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
	syntax  *linearSyntax
	aliases sortByLen
	rules   ruleList
	argv    []string
	mode    Mode
}

func Scan(p *Parser) (*linearSyntax, error) {
	// Collect all the flag aliases and sort them by length such that when looking for matching flags we match
	// flag names before short aliases (-a is not a match for -amend)
	var sortedAliases sortByLen = p.rules.GetAliases()
	sort.Sort(sortedAliases)

	s := &scanner{
		syntax:  newLinearSyntax(p),
		aliases: sortedAliases,
		rules:   p.rules,
		argv:    p.argv,
		mode:    p.mode,
	}

	if err := s.scanPos(0); err != nil {
		return nil, err
	}

	return s.syntax, nil
}

func (s *scanner) scanPos(argPos int) error {
	if len(s.argv) == argPos {
		fmt.Println("no more args to scan")
		return nil
	}

	if s.hasMode(AllowUnPrefixedFlags) {
		if err := s.scanFlag(argPos, 0, true); err != nil {
			if !IsInvalidFlag(err) {
				return err
			}
		}
	} else {
		if charPos := hasFlagPrefix(s.argv[argPos]); charPos != 0 {
			fmt.Printf("has flag prefix: %d\n", charPos)
			// TODO: If the charPos != 2, AND allowCombinedFlags then pass in true, else pass in false
			// This allows us to disambiguate '-amend' (a bunch of combined flags) and '--amend' a single flag name
			if err := s.scanFlag(argPos, charPos, s.hasMode(AllowCombinedFlags)); err != nil {
				return err
			}
		}
	}

	// scan for arguments
	if err := s.scanArgument(argPos); err != nil {
		return err
	}

	return s.scanPos(argPos + 1)
}

func (s *scanner) scanFlag(argPos, charPos int, allowCombinedFlags bool) error {
	// sanity check
	/*char := rune(s.argv[argPos][charPos])
	fmt.Printf("rune %q\n", char)
	fmt.Printf("isLetter %t\n", unicode.IsLetter(char))
	if !unicode.IsLetter(char) {
		// TODO: Should this be an error? Or should we allow upstream to decide if this error is returned to the user
		return &InvalidFlag{Msg: fmt.Sprintf("invalid character at pos '%d' for flag '%s'",
			charPos, s.argv[argPos])}
	}*/

	// Match flags to aliases by first matching the entire flag,
	// then attempt matching a subset of the flag. This allows
	// -amend to match before -a matches
	flag := s.argv[argPos][charPos:]
	fmt.Printf("attempt to match '%s'\n", flag)
	var rule *rule
	var end int

	for i := len(flag); i > 0; i-- {
		// Find an alias that matches the prefix or complete arg
		rule, end = s.matchAliases(flag[:i])

		// The flag did not match any aliases
		if rule == nil {
			s.syntax.Add(&node{
				Pos:    argPos,
				Offset: charPos,
			})
			return nil
		}

		fmt.Printf("mached '%s' to rule '%s' end=%d\n", flag[:i], rule.Name, end)
		if rule.HasFlag(isExpectingValue) {
			// If the entire flag matched the rule
			fmt.Printf("end+1: %d flag: %d\n", end+1, len(flag))
			if end+1 > len(flag) {
				// Expect the next arg to hold our value
				fmt.Printf("'%s' matched rule '%s'\n", s.argv[argPos], rule.Name)
				// consume the next arg for the value for this flag
				if len(s.argv) <= argPos+1 {
					return fmt.Errorf("expected flag '%s' to have a value", s.argv[argPos])
				}
				flagNode := &node{
					Pos:   argPos,
					Value: &s.argv[argPos+1],
					Rule:  rule,
				}
				s.syntax.Add(flagNode)
				s.syntax.Add(&node{
					Pos:      argPos + 1,
					ValueFor: flagNode,
				})
				return nil
			}
			// Is the next character an '='?
			if flag[end+1] == '=' {
				if len(flag) <= end+2 {
					return fmt.Errorf("expected flag '%s' to have a value after '='", s.argv[argPos])
				}
				// the remainder of the flag is the value
				value := flag[end+2:]
				s.syntax.Add(&node{
					Pos:   argPos,
					Value: &value,
					Rule:  rule,
				})
				return nil
			}

			if s.hasMode(AllowCombinedValues) {
				// the remainder of the flag is the value
				value := flag[end:]
				s.syntax.Add(&node{
					Pos:   argPos,
					Value: &value,
					Rule:  rule,
				})
			}
			// If we get here, then we matched part of the flag, but it's not our flag because
			// we expected a value and no value was provided. We know this because we sort our
			// matchable aliases by length such that the longest flag will match first.
		}
		// Not Expecting value

		fmt.Printf("added rule '%s' at pos '%d'\n", rule.Name, argPos)
		// Add the rule and position to our syntax
		s.syntax.Add(&node{
			Pos:  argPos,
			Rule: rule,
		})

		// Are there more un matched characters?
		if end != len(flag) {
			fmt.Println("more un-matched")
			// and we can match combined flags?
			if allowCombinedFlags {
				// attempt to match the next flag
				return s.scanFlag(argPos, end, true)
			}
			// If we get here, then we matched part of the flag, but it's not our flag because
			// there are trailing characters and allowedCombinedFlags was not set.
			continue
		}
		return nil
		// Are there more un matched characters and we can match combined flags?
		/*if charPos+1 != len(flag) && allowCombinedFlags {
			// attempt to match the next flag
			return s.scanFlag(argPos, charPos+1, true)
		}*/
	}
	return nil
}

func (s *scanner) hasMode(mode Mode) bool {
	return s.mode&mode != 0
}

func (s *scanner) matchAliases(arg string) (*rule, int) {
	fmt.Printf("looking for alias '%s'\n", arg)
	for _, alias := range s.aliases {
		i := strings.Index(arg, alias)
		if i != -1 {
			return s.rules.GetRuleByAlias(alias), i + len(arg)
		}
	}
	return nil, 0
}

// Determine if the arg begins with an '-|--' prefix, if it does
// it returns the index after the prefix where we should start evaluating the flags
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
