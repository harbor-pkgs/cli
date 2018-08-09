package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

const (
	// Used to identify the named rule is a command
	subCmdNamePrefix = "!cmd-"

	// return code used when parser encounters an error
	ErrorRetCode = 1

	// These are the names cli uses to identify where a value came from
	cliSource     = "cli-args"
	envSource     = "cli-env"
	defaultSource = "cli-default"
)

type Parser struct {
	// Max length of each line of the generated help message
	WordWrap int
	// Custom usage provided by the user
	Usage string
	// Prefix applied to all to rules that match environment variables
	EnvPrefix string
	// A description of the application
	Desc string
	// If provided, will display the provided text at the bottom of the generated help message
	Epilog string
	// The name of the application
	Name string

	// If set to true, un-matched arguments on the command line result in parse errors
	ErrOnUnknownArgs bool
	// If set to true, avoid adding a --help, -h option
	NoHelp bool
	// Don't display help message when ParseOrExit() encounters an error
	NoHelpOnError bool
	// TODO: If defined will log parse and type errors to this logger
	Logger StdLogger
	// Provide an error function, defaults to a function that panics
	ErrorFunc ErrorFunc
	// TODO: Support combined flag parsing (-s -o can be expressed as -so)
	// Allow combined flags implies 'ErrOnUnknownArgs'
	AllowCombinedFlags bool
	// TODO: Support values that directly follow a flag ( '-f value' can be expressed as '-fvalue' )
	AllowCombinedValues bool
	// TODO: Support flags without a prefix IE `ps aux`
	// Allow UnPrefixedFlags implies 'AllowCombinedFlags'
	AllowUnPrefixedFlags bool

	// The arguments we are tasked with parsing
	argv []string
	// The current state of the syntax we have parsed
	syntax *linearSyntax
	// Sorted list of parsing rules
	rules ruleList
	// Our parent parser if this instance is a sub-parser
	parent *Parser
	// A collection of stores provided by the user for retrieving values
	stores []FromStore
	// Errors accumulated when adding flags
	errs []error
	// Each new argument is assigned a sequence depending on when they were added. This
	// allows us to infer which position the argument should be expected when parsing the command line
	seqCount int
}

func NewParser() *Parser {
	return New(&Parser{})
}

func New(parent *Parser) *Parser {
	p := &Parser{
		WordWrap:         parent.WordWrap,
		EnvPrefix:        parent.EnvPrefix,
		ErrOnUnknownArgs: parent.ErrOnUnknownArgs,
		ErrorFunc:        parent.ErrorFunc,
		Name:             parent.Name,
		Desc:             parent.Desc,
		Epilog:           parent.Epilog,
		seqCount:         1,
	}

	SetDefault(&p.ErrorFunc, panicFunc)
	SetDefault(&p.WordWrap, 100)
	SetDefault(&p.Name, path.Base(os.Args[0]))

	// TODO: Consider removing this if we don't copy our subparsers
	// If rules exist, assume we are a sub parser and
	// copy all the private fields
	if parent.rules != nil {
		p.parent = parent
		p.argv = parent.argv
		p.syntax = parent.syntax
		p.rules = parent.rules
		p.stores = parent.stores
		p.seqCount = parent.seqCount
	}
	return p
}

func (p *Parser) ParseOrExit() {
	retCode, err := p.Parse(context.Background(), os.Args)
	if err != nil {
		if IsHelpError(err) {
			// TODO: Print Help message
			fmt.Printf("HELP MESSAGE HERE\n")
			os.Exit(retCode)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(retCode)
	}
}

// TODO: Support out of band command bash completions and in-band bash completions
// Parses command line arguments using os.Args if 'args' is nil.
func (p *Parser) Parse(ctx context.Context, argv []string) (int, error) {
	// Clear any previously parsed syntax
	p.syntax = newLinearSyntax()

	// Report Add() errors
	if len(p.errs) != 0 {
		return ErrorRetCode, p.errs[0]
	}

	// Sanity Check
	if len(p.rules) == 0 {
		return ErrorRetCode, errors.New("no flags or arguments defined; call Add() before calling Parse()")
	}

	fmt.Printf("parse()\n")

	// If we are the top most parent
	if p.parent == nil {
		// Allowing a sub parser to change our args can cause panics when collecting values
		p.argv = os.Args
		if argv != nil {
			p.argv = argv
		}
		// If user requested we add a help flag, and if one is not already defined
		if !p.NoHelp && p.rules.RuleWithFlag(isHelpRule) == nil {
			p.Add(&Flag{
				Help:     "display this help message and exit",
				Name:     "help",
				HelpFlag: true,
				Aliases:  []string{"h"},
			})
		}
	}

	var err error
	// Combine any rules from any parent parsers and check for duplicate rules
	if p.rules, err = p.validateRules(nil); err != nil {
		fmt.Println("validate fail")
		return ErrorRetCode, err
	}

	// Sort the rules so argument/command rules are evaluated last
	sort.Sort(p.rules)

	// Sort the aliases such that we evaluate longer alias names first
	for _, r := range p.rules {
		// TODO: Sort by length first, then alpha
		sort.Sort(sort.Reverse(sort.StringSlice(r.Aliases)))
	}

	fmt.Printf("rules: %s\n", p.rules.String())

	if err := p.parse(0); err != nil {
		fmt.Println("parse fail")
		// report flags that expect values
		return ErrorRetCode, err
	}

	subCmd := p.nextSubCmd()
	if subCmd != nil {
		// Run the sub command
		// TODO: Might not need to make a copy of ourselves, just pass in the current parser
		return subCmd(ctx, New(p))
	}

	// If the user asked to error on unknown arguments
	if p.ErrOnUnknownArgs {
		args := p.UnProcessedArgs()
		if len(args) != 0 {
			return ErrorRetCode, fmt.Errorf("provided but not defined '%s'", args[0])
		}
	}

	fmt.Printf("syntax: %s\n", p.syntax.String())
	// If they asked for help
	if p.syntax.FindWithFlag(isHelpRule) != nil {
		fmt.Printf("type %s\n", reflect.TypeOf(&HelpError{}))
		return ErrorRetCode, &HelpError{}
	}

	// If we get here, we are at the top of the parent tree and can collect values
	results := newResultStore(p.rules)

	// Retrieve values from any stores provided by the user
	for _, store := range p.stores {
		if err := results.From(ctx, store); err != nil {
			return ErrorRetCode, fmt.Errorf("while reading from store '%s': %s", store.Source(), err)
		}
	}
	fmt.Printf("1 results: %+v\n", results.values)
	results.From(ctx, newEnvStore(p.rules))
	fmt.Printf("2 results: %+v\n", results.values)
	results.From(ctx, p)
	fmt.Printf("3 results: %+v\n", results.values)
	// Apply defaults and validate required values are provided then store values
	return p.validateAndStore(results)
}

func (p *Parser) validateAndStore(rs *resultStore) (int, error) {
	// TODO: Support option exclusion `--option1 | --option2`
	// TODO: Support option dependency (option2 cannot be used unless option1 is also defined)

	fmt.Printf("4 results: %+v\n", rs.values)
	for _, rule := range p.rules {
		// get the value and how many instances of it where provided via the command line
		value, count, err := rs.Get(context.Background(), rule.Name, rule.ValueType())
		if err != nil {
			return ErrorRetCode, err
		}
		fmt.Printf("[validate]Get(%s,%s) - '%v' %d\n", rule.Name, rule.ValueType(), value, count)

		// if no instances of this rule where found
		if count == 0 {
			// Set the default value if provided
			if rule.Default != nil {
				value = *rule.Default
				fmt.Printf("default: %+v\n", value)
			} else {
				// and is required
				if rule.HasFlag(isRequired) {
					return ErrorRetCode, errors.New(rule.IsRequiredMessage())
				}
			}
		}

		// if the user dis-allows the flag to be provided more than once
		if count > 1 {
			if rule.HasFlag(isFlag) && !rule.HasFlag(canRepeat) {
				return ErrorRetCode, fmt.Errorf("unexpected duplicate flag '%s' provided", rule.Name)
			}
		}

		// ensure the value matches one of our choices
		if len(rule.Choices) != 0 {
			switch t := value.(type) {
			case string:
				if !ContainsString(t, rule.Choices, nil) {
					return ErrorRetCode, fmt.Errorf("'%s' is an invalid argument for '%s' choose from (%s)",
						value, rule.Name, strings.Join(rule.Choices, ", "))
				}
			case []string:
				for _, i := range t {
					if !ContainsString(i, rule.Choices, nil) {
						return ErrorRetCode, fmt.Errorf("'%s' is an invalid argument for '%s' choose from (%s)",
							value, rule.Name, strings.Join(rule.Choices, ", "))
					}
				}
			}
		}

		fmt.Printf("Store(%v,%d)\n", value, count)
		rule.StoreValue(value, count)
	}
	return 0, nil
}

func (p *Parser) AddStore(store FromStore) {
	p.stores = append(p.stores, store)
}

// Returns a list of all unknown arguments found on the command line if `ErrOnUnknownArgs = true`
func (p *Parser) UnProcessedArgs() []string {
	var r []string
	for i, arg := range p.argv {
		if !p.syntax.Contains(i) {
			r = append(r, arg)
		}
	}
	return r
}

func (p *Parser) Get(ctx context.Context, key string, valueType ValueType) (interface{}, int, error) {
	//fmt.Printf("Get(%s,%s)\n", key, valueType)
	rule := p.rules.GetRule(key)
	if rule == nil {
		return "", 0, nil
	}

	// TODO: If user requests positional only arguments, eliminate args/flags we find that are not in our range

	var values []string
	var count int
	// collect all the values for this rule
	for _, node := range p.syntax.FindRules(rule) {
		count++
		if node.Value != nil {
			values = append(values, *node.Value)
		}
	}

	if len(values) == 0 {
		//fmt.Printf("Get Ret: <nil>, %d\n", count)
		return nil, count, nil
	}

	switch valueType {
	case ScalarType:
		//fmt.Printf("Get Ret: %s, %d\n", values[0], count)
		return values[0], count, nil
	case ListType:
		//fmt.Printf("Get Ret: %s, %d\n", values, count)
		return values, len(values), nil
	case MapType:
		// each string in the list should be a key value pair
		// either in the form `key=value` or `{"key": "value"}`
		r := make(map[string]string)
		for _, value := range values {
			kv, err := StringToMap(value)
			if err != nil {
				return nil, 0, fmt.Errorf("during Parser.Get() map conversion: %s", err)
			}
			// Merge the key values for each of the items
			for k, v := range kv {
				r[k] = v
			}
		}
		//fmt.Printf("Get Ret: %s, %d\n", r, count)
		return r, count, nil
	}
	return nil, 0, fmt.Errorf("no such ValueType '%s'", valueType)
}

func (p *Parser) Source() string {
	return cliSource
}

func (p *Parser) nextSubCmd() CommandFunc {
	cmdNodes := p.syntax.FindWithFlag(isCommand)
	if cmdNodes != nil && len(cmdNodes) != 0 {
		for _, node := range cmdNodes {
			if !node.CmdHandled {
				node.CmdHandled = true
				return cmdNodes[0].Rule.CommandFunc
			}
		}
	}
	return nil
}

func (p *Parser) parse(pos int) error {
	if len(p.argv) == pos {
		fmt.Println("no more args")
		// No more args to parse
		return nil
	}
	fmt.Printf("argv[%d] ='%s'\n", pos, p.argv[pos])
	var skipNextPos bool

	for _, rule := range p.rules {
		fmt.Printf("Rule: %s\n", rule.Name)
		// If this is an flag rule
		if rule.HasFlag(isFlag) {
			fmt.Println("isFlag")
			var count int
			// Match any aliases for this rule
			for _, alias := range rule.Aliases {
				//fmt.Printf("alias: %s\n", alias)
				if rule.HasFlag(isExpectingValue) {
					//fmt.Println("isExpectingValue")
					// If contains an '='
					if strings.ContainsRune(p.argv[pos], '=') {
						//fmt.Println("has Equal")
						parts := strings.Split(p.argv[0], "=")
						count = matchesFlag(p.argv[pos], alias)
						if count != 0 {
							fmt.Printf("'%s' matched rule '%s'\n", p.argv[pos], rule.Name)
							p.syntax.Add(&node{
								Pos:     pos,
								Value:   &parts[1],
								RawFlag: parts[0],
								Rule:    rule,
								Count:   count,
							})
						}
					} else {
						//fmt.Println("no Equal")
						count = matchesFlag(p.argv[pos], alias)
						if count != 0 {
							fmt.Printf("'%s' matched rule '%s'\n", p.argv[pos], rule.Name)
							// consume the next arg for the value for this flag
							if len(p.argv) <= pos+1 {
								return fmt.Errorf("expected flag '%s' to have an argument", p.argv[pos])
							}

							flagNode := &node{
								Pos:     pos,
								RawFlag: p.argv[pos],
								Value:   &p.argv[pos+1],
								Rule:    rule,
								Count:   count,
							}
							p.syntax.Add(flagNode)
							p.syntax.Add(&node{
								Pos:      pos + 1,
								ValueFor: flagNode,
							})
							skipNextPos = true
						}
					}
				} else {
					count = matchesFlag(p.argv[pos], alias)
					if count != 0 {
						fmt.Printf("'%s' matched rule '%s'\n", p.argv[pos], rule.Name)
						p.syntax.Add(&node{
							Pos:     pos,
							RawFlag: p.argv[pos],
							Rule:    rule,
							Count:   count,
						})
					}
				}
			}
			// We matched a flag, move to the next arg position
			if count != 0 {
				fmt.Println("break")
				break
			}
		}

		if rule.HasFlag(isCommand) {
			if rule.Value == p.argv[pos] {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[pos],
					Rule:  rule,
					IsCmd: true,
				})
			}
		}

		// If this is an argument rule
		if rule.HasFlag(isArgument) {
			// If it's greedy
			if rule.HasFlag(canRepeat) {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[pos],
					Rule:  rule,
				})
				break
			}

			// If we haven't already match this rule with an argument
			if p.syntax.FindRules(rule) == nil {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[pos],
					Rule:  rule,
				})
				break
			}
		}
	}

	// Move to the next argument
	pos += 1

	// Skip an additional pos because we parsed it as a flag value
	if skipNextPos {
		pos += 1
	}

	return p.parse(pos)
}

// Validate our current rules and any parent rules
func (p *Parser) validateRules(rules ruleList) (ruleList, error) {
	combinedRules := p.rules

	// If were passed some rules, append them
	if rules != nil {
		combinedRules = append(combinedRules, rules...)
	}

	// Validate with our parents rules if exist
	if p.parent != nil {
		return p.parent.validateRules(combinedRules)
	}

	return combinedRules.ValidateRules()
}

func matchesFlag(arg, flag string) int {
	if "-"+flag == arg {
		return 1
	}
	if "--"+flag == arg {
		return 1
	}

	// handle -vvvvvv or -ffffff type args
	if len(flag) == 1 {
		var count int
		if strings.HasPrefix(arg, "-"+flag) {
			for _, rune := range flag[1:] {
				if string(rune) == flag {
					count++
				}
			}
		}
		if count == len(arg)+1 {
			return count
		}
	}
	return 0
}
