package cli

import (
	"fmt"
	"sort"
	"strings"
)

type sortByLen []string

func (a sortByLen) Len() int {
	return len(a)
}

func (a sortByLen) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}

func (a sortByLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type scanner struct {
	abstract *abstract
	aliases  sortByLen
	rules    ruleList
	argv     []string
	mode     Mode
}

// Scan the argv for flags and arguments and add them to our linear abstract store
func scanArgv(p *Parser) (*abstract, error) {
	// Collect all the flag aliases and sort them by length such that when looking for matching flags we match
	// long flag names before short flags (-a is not a match for -amend)
	var sortedAliases sortByLen = p.rules.GetFlagAliases()
	sort.Sort(sortedAliases)

	s := &scanner{
		abstract: newAbstract(p),
		aliases:  sortedAliases,
		rules:    p.rules,
		argv:     p.argv,
		mode:     p.cfg.Mode,
	}

	// TODO: No need for this function to be recursive, we can simply use a for loop
	// Look for stop arg, and count the number of possible positional args until stop
	// First scan for flags
	if err := s.scanFlags(0); err != nil {
		return nil, err
	}

	// TODO: Look for commands before assigning arguments
	/*for _, rule := range s.rules {
		if rule.HasFlag(isCommand) {

	}
	}*/

	// Add nodes for any args which do not have a node
	for i := range s.argv {
		node := s.abstract.AtPos(i)
		if node == nil {
			s.abstract.Add(&absNode{Pos: i})
		}
	}

	return s.abstract, nil
}

func (s *scanner) scanFlags(argPos int) error {
	if len(s.argv) == argPos {
		fmt.Println("no more args to scan")
		return nil
	}

	if s.hasMode(AllowUnPrefixedOptions) {
		if err := s.scanFlag(argPos, 0, true); err != nil {
			if !IsInvalidFlag(err) {
				return err
			}
		}
	} else {
		if charPos := hasFlagPrefix(s.argv[argPos]); charPos != 0 {
			fmt.Printf("has flag prefix: %d\n", charPos)
			// TODO: If the charPos != 2, AND allowCombinedOptions then pass in true, else pass in false
			// This allows us to disambiguate '-amend' (a bunch of combined flag) and '--amend' a single flag name
			if err := s.scanFlag(argPos, charPos, s.hasMode(AllowCombinedOptions)); err != nil {
				return err
			}
		}
	}

	return s.scanFlags(argPos + 1)
}

func (s *scanner) scanFlag(argPos, charPos int, allowCombinedOptions bool) error {
	// sanity check
	/*char := rune(s.argv[argPos][charPos])
	fmt.Printf("rune %q\n", char)
	fmt.Printf("isLetter %t\n", unicode.IsLetter(char))
	if !unicode.IsLetter(char) {
		// TODO: Should this be an error? Or should we allow upstream to decide if this error is returned to the user
		return &InvalidFlag{Msg: fmt.Sprintf("invalid character at pos '%d' for flag '%s'",
			charPos, s.argv[argPos])}
	}*/

	// Match flag to aliases by first matching the entire flag,
	// then attempt matching a subset of the flag. This allows
	// -amend to match before -a matches
	option := s.argv[argPos][charPos:]
	fmt.Printf("attempt to match '%s'\n", option)
	var rule *rule
	var end int

	for i := len(option); i > 0; i-- {
		// Find an alias that matches the prefix or complete arg
		rule, end = s.matchAliases(option[:i])

		// The option did not match any aliases
		if rule == nil {
			s.abstract.Add(&absNode{
				Flags:  isOption, // TODO: Remove isOption here?
				Pos:    argPos,
				Offset: charPos,
			})
			return nil
		}

		fmt.Printf("'%s' possible match to rule '%s' end=%d\n", option[:i], rule.Name, end)
		if rule.HasFlag(isExpectingValue) {
			// If the entire option matched the rule
			fmt.Printf("end+1: %d option: %d\n", end+1, len(option))
			if end+1 > len(option) {
				// Expect the next arg to hold our value
				fmt.Printf("'%s' matched rule '%s'\n", s.argv[argPos], rule.Name)
				// consume the next arg for the value for this option
				if len(s.argv) <= argPos+1 {
					return fmt.Errorf("expected option '%s' to have a value", s.argv[argPos])
				}
				optionNode := &absNode{
					Flags: isOption,
					Pos:   argPos,
					Value: &s.argv[argPos+1],
					Rule:  rule,
				}
				s.abstract.Add(optionNode)
				s.abstract.Add(&absNode{
					Flags:    isOption,
					Pos:      argPos + 1,
					ValueFor: optionNode,
				})
				return nil
			}
			// Is the next character an '='?
			if option[end+1] == '=' {
				if len(option) <= end+2 {
					return fmt.Errorf("expected option '%s' to have a value after '='", s.argv[argPos])
				}
				// the remainder of the option is the value
				value := option[end+2:]
				s.abstract.Add(&absNode{
					Flags: isOption,
					Pos:   argPos,
					Value: &value,
					Rule:  rule,
				})
				return nil
			}

			if s.hasMode(AllowCombinedValues) {
				// the remainder of the option is the value
				value := option[end:]
				s.abstract.Add(&absNode{
					Flags: isOption,
					Pos:   argPos,
					Value: &value,
					Rule:  rule,
				})
				return nil
			}
			// If we get here, then we matched part of the option, but it's not our option because
			// we expected a value and no value was provided. We know this because we sort our
			// matchable aliases by length such that the longest option will match first.
			s.abstract.Add(&absNode{
				Flags:  isOption,
				Pos:    argPos,
				Offset: charPos,
			})
			return nil
		}
		// Not Expecting value

		fmt.Printf("added rule '%s' at pos '%d'\n", rule.Name, argPos)
		// Add the rule and position to our abstract
		s.abstract.Add(&absNode{
			Flags: isOption,
			Pos:   argPos,
			Rule:  rule,
		})

		// Are there more un matched characters?
		if end != len(option) {
			fmt.Println("more un-matched")
			// and we can match combined flags?
			if allowCombinedOptions {
				// attempt to match the next flag
				return s.scanFlag(argPos, end, true)
			}
			// If we get here, then we matched part of the option, but it's not our option because
			// there are trailing characters and allowedCombinedFlags was not set.
			continue
		}
		return nil
		// Are there more un matched characters and we can match combined option?
		/*if charPos+1 != len(option) && allowCombinedOptions {
			// attempt to match the next option
			return s.scanFlag(argPos, charPos+1, true)
		}*/
	}
	return nil
}

// TODO: Move this method to the `Mode` object
func (s *scanner) hasMode(mode Mode) bool {
	return s.mode&mode != 0
}

func (s *scanner) matchAliases(arg string) (*rule, int) {
	fmt.Printf("looking for alias '%s'\n", arg)
	for _, alias := range s.aliases {
		if strings.HasPrefix(arg, alias) {
			return s.rules.GetRuleByAlias(alias), len(alias)
		}
	}
	return nil, 0
}

// Determine if the arg begins with an '-|--' prefix, if it does
// it returns the index after the prefix where we should start evaluating the options
func hasFlagPrefix(arg string) int {
	switch len(arg) {
	case 0, 1:
		return 0
	case 2:
		if arg == "--" {
			return 0
		}
		if arg[0] == '-' {
			return 1
		}
	}
	if arg[0] == '-' {
		if arg[1] == '-' {
			return 2
		}
		return 1
	}
	return 0
}

func (s *scanner) scanArgument(argPos int) error {
	// TODO: subcommands
	return nil
}
