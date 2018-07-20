package cli

import (
	"fmt"
	"regexp"
	"strings"
)

var regexInValidRuleName = regexp.MustCompile(`[!"#$&'/()*;<>{|}\\\\~\s]`)
var regexHasNonWordPrefix = regexp.MustCompile(`^(\W+)([\w|-]*)$`)

type rules []*rule

func (r rules) Len() int {
	return len(r)
}

func (r rules) Less(left, right int) bool {
	return r[left].Sequence < r[right].Sequence
}

func (r rules) Swap(left, right int) {
	r[left], r[right] = r[right], r[left]
}

func (r rules) ValidateRules() error {
	var greedyRule *rule
	for idx, rule := range r {
		// Duplicate rule check
		next := idx + 1
		if next < len(r) {
			for ; next < len(r); next++ {
				// If the names are the same
				if rule.Name == r[next].Name {
					return fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
				}
				// If the alias is a duplicate
				for _, alias := range r[next].Aliases {
					var duplicate string

					// if rule.Aliases contains 'alias'
					for _, item := range rule.Aliases {
						if item == alias {
							duplicate = alias
						}
					}
					if len(duplicate) != 0 {
						return fmt.Errorf("duplicate alias '%s' for '%s' redefined by '%s'",
							duplicate, rule.Name, r[next].Name)
					}
				}
				if rule.Name == r[next].Name {
					return fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
				}
			}
		}

		if rule.Name == "" {
			return fmt.Errorf("refusing to parse %s with no name'", rule.Type())
		}

		if regexHasNonWordPrefix.MatchString(rule.Name) {
			return fmt.Errorf("'%s' is an invalid name for an '%s'", rule.Name, rule.Type())
		}

		// Check for invalid option and argument names
		if regexInValidRuleName.MatchString(rule.Name) {
			if !strings.HasPrefix(rule.Name, subCmdNamePrefix) {
				return fmt.Errorf("bad %s '%s'; contains invalid characters",
					rule.Type(), rule.Name)
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

func (r rule) Type() string {
	switch {
	case r.HasFlag(IsFlag):
		return "flag"
	case r.HasFlag(IsArgument):
		return "argument"
	case r.HasFlag(IsCommand):
		return "command"
	}
	return "unknown"
}

func (r rules) GetRule(name string) *rule {
	for _, rule := range r {
		if rule.Name == name {
			return rule
		}
	}
	return nil
}

func (r rules) RuleWithFlag(flag ruleFlag) *rule {
	for _, rule := range r {
		if rule.HasFlag(flag) {
			return rule
		}
	}
	return nil
}
