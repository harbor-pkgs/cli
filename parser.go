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

// This is just so Add() won't complain we didn't provide a 'IsSet' for our auto added Help option
var hasHelp bool

type Mode int64

const (
	// TODO: Support combined option parsing (-s -o can be expressed as -so)
	// Cannot be combined with 'IgnoreUnknownArgs'
	AllowCombinedOptions Mode = 1 << iota
	// TODO: Support values that directly follow a option ( '-f value' can be expressed as '-fvalue' )
	AllowCombinedValues
	// TODO: Support options without a prefix IE `ps aux`
	// AllowUnPrefixedOptions implies 'AllowCombinedOptions'
	AllowUnPrefixedOptions
	// Arguments that don't match a option or argument defined by the parser do not result in an error
	// Unknown args can be retrieved using Parser.UnProcessedArgs()
	IgnoreUnknownArgs
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
	// The current state of the abstract we have parsed
	abstract *abstract
	// Sorted list of parsing rules
	rules ruleList
	// Our parent parser if this instance is a sub-parser
	parent *Parser
	// A collection of stores provided by the user for retrieving values
	stores []FromStore
	// Errors accumulated when adding options
	errs []error
	// Each new argument is assigned a sequence depending on when they were added. This
	// allows us to infer which position the argument should be expected when parsing the command line
	seqCount int
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
		p.abstract = parent.abstract
		p.rules = parent.rules
		p.stores = parent.stores
		p.seqCount = parent.seqCount
	}*/
	return p
}

// Returns true if the mode or set of modes is selected
func (p *Parser) HasMode(mode Mode) bool {
	return p.cfg.Mode&mode != 0
}

// Set or clear a mode on the current parser
func (p *Parser) SetMode(mode Mode, set bool) {
	if set {
		p.cfg.Mode = p.cfg.Mode | mode
	} else {
		mask := p.cfg.Mode ^ mode
		p.cfg.Mode &= mask
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
	// Report Add() errors
	if len(p.errs) != 0 {
		return ErrorRetCode, p.errs[0]
	}

	// Sanity Check
	if len(p.rules) == 0 {
		return ErrorRetCode, errors.New("no options or arguments defined; call Add() before calling Parse()")
	}

	fmt.Printf("parse()\n")

	// If we are the top most parent
	if p.parent == nil {
		// If user requested we add a help option, and if one is not already defined
		if !p.HasMode(NoHelp) && p.rules.GetRuleByFlag(isHelpRule) == nil {
			p.Add(&Flag{
				Help:    "display this help message and exit",
				Name:    "help",
				Flags:   isHelpRule,
				IsSet:   &hasHelp,
				Aliases: []string{"h"},
			})
		}
	}

	var err error
	// Combine any rules from any parent parsers and check for duplicate rules
	if p.rules, err = p.validateRules(nil); err != nil {
		fmt.Println("validate fail")
		return ErrorRetCode, err
	}

	// Sort the aliases such that we evaluate longer alias names first
	for _, r := range p.rules {
		// TODO: Sort by length first, then alpha
		sort.Sort(sort.Reverse(sort.StringSlice(r.Aliases)))
	}

	//fmt.Printf("rules: %s\n", p.rules.String())

	// TODO: Create LexNode's from the argv input
	p.lex, err = toLex(argv)

	// TODO: Find flags or expand flags in the lex

	// TODO: Find arguments and sub-commands in lex

	// Scan the argv and attempt to assign rules to argv positions, this is
	// only a best effort since a sub command might add new options and args.
	/*if p.abstract, err = scanArgv(p); err != nil {
		fmt.Println("scan fail")
		// report options that expect values
		return ErrorRetCode, err
	}

	if subCmd := p.nextSubCmd(); subCmd != nil {
		// Run the sub command
		// TODO: Might not need to make a copy of ourselves, just pass in the current parser
		return subCmd(ctx, New(p))
	}

	fmt.Printf("abstract: %s\n", p.abstract.String())*/
	// --help is a special case option, as it short circuits the normal store
	// and validation of arguments. This allows the user to pass other arguments
	// along side -h and still get a help message before getting invalid arg errors
	// TODO: Support other short circuit options besides -h, in the case a user wishes to
	// TODO: Not use -h has the help option.
	if p.lex.FindWithFlag(isHelpRule) != nil {
		fmt.Printf("type %s\n", reflect.TypeOf(&HelpError{}))
		return ErrorRetCode, &HelpError{}
	}

	// If we get here, we are at the top of the parent tree and we can assign positional arguments
	/*if err := p.applyArguments(); err != nil {
		return ErrorRetCode, err
	}*/

	// Given the lex, create a store of results
	results := newResultStore(p.lex)

	// TODO: Put all the stores in `p.stores` and process them in this for loop.
	//  This might provide future features like, having a user store take precedence over
	//  an Env store.

	// Retrieve values from external stores provided by the user first
	for _, store := range p.stores {
		if err := results.From(ctx, store); err != nil {
			return ErrorRetCode, fmt.Errorf("while reading from store '%s': %s", store.Source(), err)
		}
	}
	fmt.Printf("User store: %+v\n", results.values)

	if err := results.From(ctx, newEnvStore(p.rules)); err != nil {
		return ErrorRetCode, err
	}
	fmt.Printf("Env store: %+v\n", results.values)

	if err := results.From(ctx, p.lex); err != nil {
		return ErrorRetCode, err
	}
	fmt.Printf("Lex store: %+v\n", results.values)

	// Apply defaults and validate required values are provided then store values
	return p.validateAndStore(results, p.lex)
}

func (p *Parser) validateAndStore(rs *resultStore) (int, error) {
	// TODO: Support option exclusion `--option1 | --option2`
	// TODO: Support option dependency (option2 cannot be used unless option1 is also defined)

	// If the user asked to error on unknown arguments
	if !p.HasMode(IgnoreUnknownArgs) {
		args := p.UnProcessedArgs()
		if len(args) != 0 {
			// TODO: Review if this is the correct wording for an unknown argument
			return ErrorRetCode, fmt.Errorf("'%s' was provided but not defined", args[0])
		}
	}

	fmt.Printf("4 results: %+v\n", rs.values)
	for _, rule := range p.rules {
		// get the value and how many instances of it where provided via the command line
		value, count, err := rs.Get(context.Background(), rule.Name, rule.Flags)
		if err != nil {
			return ErrorRetCode, err
		}
		fmt.Printf("[validate]Get(%s,%s) - '%v' %d\n", rule.Name, rule.Kind(), value, count)

		// if no instances of this rule where found
		if count == 0 {
			// Set the default value if provided
			if rule.Default != nil {
				if value, count, err = convToKind([]string{*rule.Default}, rule.Flags, 1); err != nil {
					return ErrorRetCode, err
				}
				fmt.Printf("default: %+v\n", value)
			} else {
				// and is required
				if rule.HasFlag(Required) {
					return ErrorRetCode, errors.New(rule.IsRequiredMessage())
				}
				// Nothing else to be done; no value to set
				continue
			}
		}

		// if the user dis-allows the option to be provided more than once
		if count > 1 {
			if rule.HasFlag(isOption) && !rule.HasFlag(CanRepeat) {
				return ErrorRetCode, fmt.Errorf("unexpected duplicate option '%s' provided", rule.Name)
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
		if err = rule.StoreValue(value, count); err != nil {
			return ErrorRetCode, fmt.Errorf("invalid value for %s '%s': %s", rule.Type(), rule.Name, err)
		}
	}
	return 0, nil
}

func (p *Parser) AddStore(store FromStore) {
	p.stores = append(p.stores, store)
}

func (p *Parser) applyArguments() error {
	// TODO: Multiple arguments with greedy flag is not allowed
	// TODO: Arguments with default values that follow greedy arguments are not allowed. (ambiguous)
	rules := p.rules.GetRulesWithFlag(isArgument)
	args := p.abstract.UnknownArgs()

	// If there are not enough args left for each argument rule
	if len(args) <= len(rules) {
		// Simple algo, each rule is assigned to each args until the args run out
		return nil
	}

	// Either there is a greedy argument, or we have unknown args

	// Assign rules to arg until we find a greedy rule, gobble up all the args
	// then start at the bottom of the rules and work our way back up to the greedy rule

	// Returns true if we found a greedy rule
	if apply(rules, args) {
		// If we find greedy
		// Reverse the order of the args and rules
		// and apply again until we hit the greedy rule and stop
		apply(rules, args)
	}
	return nil
}

func apply(rules ruleList, args nodeList) bool {
	// TODO: Fix me
	return false
}

// Returns a list of all unknown arguments found on the command line if `ErrOnUnknownArgs = true`
// TODO: Use UnknownArgs() from abstract
func (p *Parser) UnProcessedArgs() []string {
	if p.abstract == nil {
		return []string{}
	}

	var r []string
	for i, arg := range p.argv {
		if !p.abstract.Contains(i) {
			r = append(r, arg)
		}
	}
	return r
}

func (p *Parser) nextSubCmd() CommandFunc {
	if p.abstract == nil {
		return nil
	}

	cmdNodes := p.abstract.FindWithFlag(isCommand)
	if cmdNodes != nil && len(cmdNodes) != 0 {
		for _, node := range cmdNodes {
			if !node.Flags.Has(cmdHandled) {
				node.Flags.Set(cmdHandled, true)
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
