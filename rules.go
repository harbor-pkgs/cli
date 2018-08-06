package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var regexInValidRuleName = regexp.MustCompile(`[!"#$&'/()*;<>{|}\\\\~\s]`)
var regexHasNonWordPrefix = regexp.MustCompile(`^(\W+)([\w|-]*)$`)

type ruleList []*rule

func (r ruleList) Len() int {
	return len(r)
}

func (r ruleList) Less(left, right int) bool {
	return r[left].Sequence < r[right].Sequence
}

func (r ruleList) Swap(left, right int) {
	r[left], r[right] = r[right], r[left]
}

func (r ruleList) String() string {
	var lines []string
	for i, rule := range r {
		lines = append(lines, fmt.Sprintf("[%d] %s", i, spew.Sdump(rule)))
	}
	return strings.Join(lines, "\n")
}

func (r ruleList) ValidateRules() (ruleList, error) {
	var greedyRule *rule
	for idx, rule := range r {
		// Duplicate rule check
		next := idx + 1
		if next < len(r) {
			for ; next < len(r); next++ {
				// If the names are the same
				if rule.Name == r[next].Name {
					return nil, fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
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
						return nil, fmt.Errorf("duplicate alias '%s' for '%s' redefined by '%s'",
							duplicate, rule.Name, r[next].Name)
					}
				}
				if rule.Name == r[next].Name {
					return nil, fmt.Errorf("duplicate argument or flag '%s' defined", rule.Name)
				}
			}
		}
		// TODO: rules cannot store map and slices simultaneously check for flags set and error if both exists

		if rule.Name == "" {
			return nil, fmt.Errorf("refusing to parse %s with no name'", rule.Type())
		}

		if regexHasNonWordPrefix.MatchString(rule.Name) {
			return nil, fmt.Errorf("'%s' is an invalid name for an '%s'", rule.Name, rule.Type())
		}

		// Check for invalid option and argument names
		if regexInValidRuleName.MatchString(rule.Name) {
			if !strings.HasPrefix(rule.Name, subCmdNamePrefix) {
				return nil, fmt.Errorf("bad %s '%s'; contains invalid characters",
					rule.Type(), rule.Name)
			}
		}

		if !rule.HasFlag(isArgument) {
			continue
		}

		// If we already found a greedy rule, no other argument should follow
		if greedyRule != nil {
			return nil, fmt.Errorf("'%s' is ambiguous when following greedy argument '%s'",
				rule.Name, greedyRule.Name)
		}

		// Check for ambiguous greedy arguments
		if rule.HasFlag(canRepeat) {
			if greedyRule == nil {
				greedyRule = rule
			}
		}
	}
	return r, nil
}

type ValueType string

const (
	StringType ValueType = "string"
	ListType   ValueType = "list"
	MapType    ValueType = "map"
)

func (r rule) ValueType() ValueType {
	switch {
	case r.HasFlag(isList):
		return ListType
	case r.HasFlag(isMap):
		return MapType
	}
	return StringType
}

func (r rule) Type() string {
	switch {
	case r.HasFlag(isFlag):
		return "flag"
	case r.HasFlag(isArgument):
		return "argument"
	case r.HasFlag(isCommand):
		return "command"
	}
	return "unknown"
}

func (r ruleList) GetRule(name string) *rule {
	for _, rule := range r {
		if rule.Name == name {
			return rule
		}
	}
	return nil
}

func (r ruleList) RuleWithFlag(flag ruleFlag) *rule {
	for _, rule := range r {
		if rule.HasFlag(flag) {
			return rule
		}
	}
	return nil
}
