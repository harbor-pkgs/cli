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

type Mode int64

const (
	// TODO: Support combined flag parsing (-s -o can be expressed as -so)
	// Cannot be combined with 'AllowUnknownArgs'
	AllowCombinedFlags Mode = 1 << iota
	// TODO: Support values that directly follow a flag ( '-f value' can be expressed as '-fvalue' )
	AllowCombinedValues
	// TODO: Support flags without a prefix IE `ps aux`
	// Allow UnPrefixedFlags implies 'AllowCombinedFlags'
	AllowUnPrefixedFlags
	// Arguments that don't match a flag or argument defined by the parser do not result in an error
	// Unknown args can be retrieved using Parser.UnProcessedArgs()
	AllowUnknownArgs
	// Avoid adding a --help, -h option
	NoHelp
	// Don't display help message when ParseOrExit() encounters an error
	NoHelpOnError
)

type Config struct {
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
	// TODO: If defined will log parse and type errors to this logger
	Logger StdLogger
	// Provide an error function, defaults to a function that prints the error to stdout and panics
	ErrorFunc ErrorFunc
	// Represents the parsers mode which dictates how the parser reacts to input
	Mode Mode
}

type Parser struct {
	// Parser configuration
	cfg Config
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
	// Represents the parsers mode which dictates how the parser reacts to input
	mode Mode
}

func New(config *Config) *Parser {
	var cfg Config
	if config != nil {
		cfg = *config
	}

	SetDefault(&cfg.ErrorFunc, panicFunc)
	SetDefault(&cfg.WordWrap, 100)
	SetDefault(&cfg.Name, path.Base(os.Args[0]))

	p := &Parser{
		cfg:      cfg,
		seqCount: 1,
	}

	// TODO: Consider removing this if we don't copy our subparsers
	// If rules exist, assume we are a sub parser and
	// copy all the private fields
	/*if parent.rules != nil {
		p.parent = parent
		p.argv = parent.argv
		p.syntax = parent.syntax
		p.rules = parent.rules
		p.stores = parent.stores
		p.seqCount = parent.seqCount
	}*/
	return p
}

// Returns true if the mode or set of modes is selected
func (p *Parser) HasMode(mode Mode) bool {
	return p.mode&mode != 0
}

// Set or clear a mode on the current parser
func (p *Parser) SetMode(mode Mode, set bool) {
	if set {
		p.mode = p.mode | mode
	} else {
		mask := p.mode ^ mode
		p.mode &= mask
	}
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
	p.syntax = nil

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
		if !p.HasMode(NoHelp) && p.rules.RuleWithFlag(isHelpRule) == nil {
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

	//fmt.Printf("rules: %s\n", p.rules.String())

	// Scan the argv and attempt to assign rules to argv positions, this is
	// only a best effort since a sub command might add new flags and args.
	if p.syntax, err = scanArgv(p); err != nil {
		fmt.Println("scan fail")
		// report flags that expect values
		return ErrorRetCode, err
	}

	/*if subCmd := p.nextSubCmd(); subCmd != nil {
		// Run the sub command
		// TODO: Might not need to make a copy of ourselves, just pass in the current parser
		return subCmd(ctx, New(p))
	}*/

	fmt.Printf("syntax: %s\n", p.syntax.String())
	// --help is a special case flag, as it short circuits the normal store
	// and validation of arguments. This allows the user to pass other arguments
	// along side -h and still get a help message before getting invalid arg errors
	// TODO: Support other short circuit flags besides -h, in the case a user wishes to
	// TODO: Not use -h has the help flag.
	if p.syntax.FindWithFlag(isHelpRule) != nil {
		fmt.Printf("type %s\n", reflect.TypeOf(&HelpError{}))
		return ErrorRetCode, &HelpError{}
	}

	// If we get here, we are at the top of the parent tree
	results := newResultStore(p.rules)

	// Retrieve values from any stores provided by the user first
	for _, store := range p.stores {
		if err := results.From(ctx, store); err != nil {
			return ErrorRetCode, fmt.Errorf("while reading from store '%s': %s", store.Source(), err)
		}
	}
	fmt.Printf("1 results: %+v\n", results.values)
	results.From(ctx, newEnvStore(p.rules))
	fmt.Printf("2 results: %+v\n", results.values)
	results.From(ctx, p.syntax)
	fmt.Printf("3 results: %+v\n", results.values)
	// Apply defaults and validate required values are provided then store values
	return p.validateAndStore(results)
}

func (p *Parser) validateAndStore(rs *resultStore) (int, error) {
	// TODO: Support option exclusion `--option1 | --option2`
	// TODO: Support option dependency (option2 cannot be used unless option1 is also defined)

	// If the user asked to error on unknown arguments
	if !p.HasMode(AllowUnknownArgs) {
		args := p.UnProcessedArgs()
		if len(args) != 0 {
			return ErrorRetCode, fmt.Errorf("'%s' was provided but not defined", args[0])
		}
	}

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
	if p.syntax == nil {
		return []string{}
	}

	var r []string
	for i, arg := range p.argv {
		if !p.syntax.Contains(i) {
			r = append(r, arg)
		}
	}
	return r
}

func (p *Parser) nextSubCmd() CommandFunc {
	if p.syntax == nil {
		return nil
	}

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
