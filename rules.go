package cli

import (
	"fmt"
	"regexp"
	"strings"
)

var regexInValidRuleName = regexp.MustCompile(`[!"#$&'/()*;<>{|}\\\\~\s]`)

type rules []*rule

func (s rules) Len() int {
	return len(s)
}

func (s rules) Less(left, right int) bool {
	return s[left].Sequence < s[right].Sequence
}

func (s rules) Swap(left, right int) {
	s[left], s[right] = s[right], s[left]
}

func (s rules) ValidateRules() error {
	var greedyRule *rule
	for idx, rule := range s {
		// Duplicate rule check
		next := idx + 1
		if next < len(s) {
			for ; next < len(s); next++ {
				// If the names are the same
				if rule.Name == s[next].Name {
					return fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
				}
				// If the alias is a duplicate
				for _, alias := range s[next].Aliases {
					var duplicate string

					// if rule.Aliases contains 'alias'
					for _, item := range rule.Aliases {
						if item == alias {
							duplicate = alias
						}
					}
					if len(duplicate) != 0 {
						return fmt.Errorf("duplicate alias '%s' for '%s' redefined by '%s'",
							duplicate, rule.Name, s[next].Name)
					}
				}
				if rule.Name == s[next].Name {
					return fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
				}
			}
		}

		// Check for invalid option and argument names
		if regexInValidRuleName.MatchString(rule.Name) {
			if !strings.HasPrefix(rule.Name, subCmdNamePrefix) {
				return fmt.Errorf("bad argument or flag '%s'; contains invalid characters",
					rule.Name)
			}
		}

		if !rule.HasFlag(IsArgument) {
			continue
		}

		// If we already found a greedy rule, no other argument should follow
		if greedyRule != nil {
			return fmt.Errorf("'%s' is ambiguous when following greedy argument '%s'",
				rule.Name, greedyRule.Name)
		}

		// Check for ambiguous greedy arguments
		if rule.HasFlag(IsGreedy) {
			if greedyRule == nil {
				greedyRule = rule
			}
		}
	}
	return nil
}

func (s rules) GetRule(name string) *rule {
	for _, rule := range s {
		if rule.Name == name {
			return rule
		}
	}
	return nil
}

func (s rules) RuleWithFlag(flag ruleFlag) *rule {
	for _, rule := range s {
		if rule.HasFlag(flag) {
			return rule
		}
	}
	return nil
}
