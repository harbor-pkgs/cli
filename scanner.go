package cli

import (
	"fmt"
	"unicode"
)

func (p *Parser) scan(argPos int) error {
	if len(p.argv) == argPos {
		fmt.Println("no more args to scan")
		return nil
	}

	if p.AllowUnPrefixedFlags {
		if err := p.scanFlag(argPos, 0, true); err != nil {
			if !IsInvalidFlag(err) {
				return err
			}
		}
	} else {
		if charPos := p.isFlag(p.argv[argPos]); charPos != 0 {
			if err := p.scanFlag(argPos, 0, p.AllowCombinedValues); err != nil {
				return err
			}
		}
	}

	// scan for arguments
	if err := p.scanArgument(argPos); err != nil {
		return err
	}

	return nil
}

// Determine if the arg begins with an '-|--' prefix, if it does
// it returns the index after the prefix where we should start evaluating the flags
func (p *Parser) isFlag(arg string) int {
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

func (p *Parser) scanFlag(argPos, charPos int, allowCombinedFlags bool) error {
	// sanity check
	char := rune(p.argv[argPos][charPos])
	if !unicode.IsNumber(char) || !unicode.IsLetter(char) {
		return &InvalidFlag{Msg: fmt.Sprintf("invalid character at pos '%d' for flag '%s'",
			charPos, p.argv[argPos])}
	}

	// Match flags to aliases by first matching the entire flag,
	// then attempt matching a subset of the flag. This allows
	// -amend to match before -a matches
	flag := p.argv[argPos][charPos:]
	for i := len(flag); i > 0; i-- {
		// Find an alias that matches this flag
		rule, end := p.matchRule(flag[:i])
		if rule != nil {
			if rule.HasFlag(isExpectingValue) {
				// If the entire flag matched the rule
				if end+1 == len(flag) {
					// Expect the next arg to hold our value
				} else {
					// Is the next character an '='?
					if flag[end+1] == '=' {
						// the remainder of the flag is the value
					}
					if p.AllowCombinedValues {
						// the remainder of the flag is the value
					}
					// TODO: Throw an error or ignore this match? (I think we ignore)
				}
			}
			// Are there more un matched characters and we can match combined flags?
			if end+1 != len(flag) && allowCombinedFlags {
				// attempt to match the next flag
				if err := p.scanFlag(argPos, end+1, true); err != nil {
					return err
				}
			}
			// Add the rule and position to our syntax
		}
		// The flag did not match
		if allowCombinedFlags {
			return &InvalidFlag{Msg: fmt.Sprintf("unrecognized flag '%s' at pos '%d'", p.argv[argPos], charPos)}
		}
	}

	return nil
}

func (p *Parser) scanArgument(argPos int) error {
	// TODO: subcommands
	return nil
}
