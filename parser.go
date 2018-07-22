package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type parseFlag int64

const (
	// Used to identify the named rule is a command
	subCmdNamePrefix = "!cmd-"

	// return code used when parser encounters an error
	errorCode = 1

	// These are the names cli uses to identify where a value came from
	cliSource     = "cli-args"
	envSource     = "cli-env"
	defaultSource = "cli-default"

	IsFormatted parseFlag = 1 << iota
	AddHelpFlag
)

type Parser struct {
	// The length of lines used to word wrap the description
	WordWrap int
	// Prefix applied to all to rules that match environment variables
	EnvPrefix string
	// A description of the application
	Desc string
	// The name of the application
	Name string
	// TODO: If defined will log parse and type errors to this logger
	Logger StdLogger
	// Provide an error function, defaults to a function that panics
	ErrorFunc ErrorFunc

	// The arguments we are tasked with parsing
	argv []string
	// The current state of the syntax we have parsed
	syntax syntax
	// Sorted list of parsing rules
	rules rules
	// flags that modify parser behavior
	flags parseFlag
	// Our parent parser if this instance is a sub-parser
	parent *Parser
	// A collection of stores provided by the user for retrieving values
	stores []FromStore
}

func New(parent *Parser) *Parser {
	p := &Parser{
		WordWrap:  parent.WordWrap,
		EnvPrefix: parent.EnvPrefix,
	}

	if parent.ErrorFunc == nil {
		p.ErrorFunc = panicFunc
	}

	// If rules exist, assume we are a sub parser and
	// copy all the private fields
	if parent.rules != nil {
		p.parent = parent
		p.argv = parent.argv
		p.syntax = parent.syntax
		p.rules = parent.rules
		p.flags = parent.flags
		p.flags = parent.flags
		p.stores = parent.stores
	} else {
		// TODO: Set flags
	}
	return p
}

func (p *Parser) hasFlag(flag parseFlag) bool {
	return p.flags&flag != 0
}

func (p *Parser) setFlag(flag parseFlag) {
	p.flags = p.flags | flag
}

func (p *Parser) clearFlag(flag parseFlag) {
	mask := p.flags ^ flag
	p.flags &= mask
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

// Parses command line arguments using os.Args if 'args' is nil.
func (p *Parser) Parse(ctx context.Context, argv []string) (int, error) {
	var err error

	if argv == nil {
		argv = os.Args
	}

	// Sanity Check
	if len(p.rules) == 0 {
		return errorCode, errors.New("no flags or arguments defined; call Add() before calling Parse()")
	}

	// If we are the top most parent
	if p.parent == nil {
		// If user requested we add a help flag, and if one is not already defined
		if p.hasFlag(AddHelpFlag) && p.rules.RuleWithFlag(IsHelpRule) == nil {
			// TODO: Add a help flag
			// p.AddFlag("--help").Alias("-h").IsTrue().Help("Display a help message and exit")
		}
	}

	// Check for duplicate rules, and combine any rules from parent parsers
	if p.rules, err = p.validateRules(nil); err != nil {
		return errorCode, err
	}

	// Sort the rules so positional rules are evaluated last
	sort.Sort(p.rules)

	if err := p.parse(0); err != nil {
		// report missing flag values
		return errorCode, err
	}

	subCmd := p.nextSubCmd()
	if subCmd != nil {
		// Run the sub command
		// TODO: Might not need to make a copy of ourselves
		return subCmd(ctx, New(p))
	}

	// If they asked for help
	if p.AskedForHelp() {
		return errorCode, &HelpError{}
	}

	// If we get here, we are at the top of the parent tree and can collect values
	results := newResultStore(p.rules)

	results.From(ctx, p)
	results.From(ctx, newEnvStore(p.rules))

	// Retrieve values from any stores provided by the user
	for _, store := range p.stores {
		if err := results.From(ctx, store); err != nil {
			return errorCode, fmt.Errorf("while reading from store '%s': %s", store.Source(), err)
		}
	}
	// Apply defaults and validate required values are provided then store values
	return p.apply(results)
}

func (p *Parser) apply(rs *resultStore) (int, error) {
	for _, rule := range p.rules {

		// get the value and how many instances of it where provided on the command line
		value, count := p.Get(context.Background(), rule.Name, rule.ValueType())

		// if no instances of this rule where found
		if count == 0 {
			// Set the default value if provided
			if rule.Default != nil {
				rs.Set(rule.Name, defaultSource, *rule.Default, 1)
			} else {
				// and is required
				if rule.HasFlag(IsRequired) {
					return errorCode, errors.New(rule.IsRequiredMessage())
				}
			}
		}

		// if the user dis-allows the flag to be provided more than once
		if count > 1 {
			if rule.HasFlag(IsFlag) && !rule.HasFlag(IsGreedy) {
				return errorCode, fmt.Errorf("unexpected duplicate flag '%s' provided", rule.Name)
			}
		}

		// If has no value
		if value == nil {
			// TODO: Check for IsCount and ExecpectValue and return error
			// Use the count as the value
			value = fmt.Sprintf("%d", count)
		}

		// ensure the value matches one of our choices
		if len(rule.Choices) != 0 {
			switch t := value.(type) {
			case string:
				if !ContainsString(t, rule.Choices, nil) {
					return errorCode, fmt.Errorf("'%s' is an invalid argument for '%s' choose from (%s)",
						value, rule.Name, strings.Join(rule.Choices, ", "))
				}
			case []string:
				for _, i := range t {
					if !ContainsString(i, rule.Choices, nil) {
						return errorCode, fmt.Errorf("'%s' is an invalid argument for '%s' choose from (%s)",
							value, rule.Name, strings.Join(rule.Choices, ", "))
					}
				}
			}
		}

		rule.StoreValue(value, count)
	}
	return 0, nil
}

func (p *Parser) Get(ctx context.Context, key string, typ ValueType) (interface{}, int) {
	rule := p.rules.GetRule(key)
	if rule == nil {
		return "", 0
	}

	var list []string
	// Count the number of times this flag was seen
	for _, node := range p.syntax.FindRules(rule) {
		if node.Value != nil {
			list = append(list, *node.Value)
		}
	}

	if len(list) == 0 {
		return list[0], 1
	}

	// If found multiple flags with values, return them as a JSON list of strings
	value, err := json.Marshal(list)
	if err != nil {
		p.ErrorFunc(fmt.Sprintf("during Parser.Get() while marshaling values to list: %s", err))
	}
	return string(value), len(list)
}

func (p *Parser) Source() string {
	return cliSource
}

func (p *Parser) nextSubCmd() CommandFunc {
	cmdNodes := p.syntax.FindWithFlag(IsCommand)
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

func (p *Parser) AskedForHelp() bool {
	return p.syntax.FindWithFlag(IsHelpRule) != nil
}

func (p *Parser) parse(pos int) error {
	if len(p.argv) == pos {
		// No more args to parse
		return nil
	}
	var skipNextPos bool

	for _, rule := range p.rules {
		// If this is an flag rule
		if rule.HasFlag(IsFlag) {
			var count int
			// Match any aliases for this rule
			for _, alias := range rule.Aliases {
				if rule.HasFlag(IsExpectingValue) {
					// If contains an '='
					if strings.ContainsRune(p.argv[pos], '=') {
						parts := strings.Split(p.argv[0], "=")
						count = matchesFlag(p.argv[pos], alias)
						if count != 0 {
							p.syntax.Add(&node{
								Pos:     pos,
								Value:   &parts[1],
								RawFlag: parts[0],
								Rule:    rule,
								Count:   count,
							})
						}
					} else {
						count = matchesFlag(p.argv[pos], alias)
						if count != 0 {
							// consume the next arg for the value for this flag
							if len(p.argv) < pos+1 {
								return fmt.Errorf("expected '%p' to have an argument", rule.Name)
							}

							p.syntax.Add(&node{
								Pos:     pos,
								RawFlag: p.argv[pos],
								Value:   &p.argv[pos+1],
								Rule:    rule,
								Count:   count,
							})
							skipNextPos = true
						}
					}
				} else {
					count = matchesFlag(p.argv[pos], alias)
					if count != 0 {
						p.syntax.Add(&node{
							Pos:     pos,
							RawFlag: p.argv[pos],
							Value:   &p.argv[pos+1],
							Rule:    rule,
							Count:   count,
						})
					}
				}
			}
			// We matched a flag, move to the next arg position
			if count != 0 {
				break
			}
		}

		if rule.HasFlag(IsCommand) {
			if rule.Value == p.argv[pos] {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[0],
					Rule:  rule,
					IsCmd: true,
				})
			}
		}

		// If this is an argument rule
		if rule.HasFlag(IsArgument) {
			// If it's greedy
			if rule.HasFlag(IsGreedy) {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[0],
					Rule:  rule,
				})
				break
			}

			// If we haven't already match this rule with an argument
			if p.syntax.FindRules(rule) == nil {
				p.syntax.Add(&node{
					Pos:   pos,
					Value: &p.argv[0],
					Rule:  rule,
				})
				break
			}
		}
	}

	// Skip the next pos because we parsed it p a flag value
	if skipNextPos {
		pos += 1
	}

	return p.parse(pos)
}

// Validate our current rules and any parent rules
func (p *Parser) validateRules(rules rules) (rules, error) {
	var validate rules

	// If were passed some rules, append them
	if rules != nil {
		validate = append(p.rules, rules...)
	}

	// Validate with our parents rules if exist
	if p.parent != nil {
		return p.parent.validateRules(validate)
	}

	return nil, validate.ValidateRules()
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
