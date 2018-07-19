package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type valueSrc struct {
	source string
	value  string
}

type resultStore struct {
	rules  rules
	values map[string]valueSrc
}

type FromStore interface {
	Source() string
	Get(context.Context, string) (string, bool)
}

func newResultStore(rules rules) *resultStore {
	return &resultStore{
		rules: rules,
	}
}

func (rs *resultStore) From(ctx context.Context, from FromStore) error {
	for _, rule := range rs.rules {
		if value, ok := from.Get(ctx, rule.Name); ok {
			rs.values[rule.Name] = valueSrc{
				source: from.Source(),
				value:  value,
			}
		}
	}
	return nil
}

func (rs *resultStore) Apply() (int, error) {
	for _, rule := range rs.rules {

		// apply default value if provided
		if _, ok := rs.values[rule.Name]; !ok {
			// Set the default value if provided
			if rule.Default != nil {
				rs.values[rule.Name] = valueSrc{
					source: defaultSource,
					value:  *rule.Default,
				}
			}
		}

		// if has no value
		value, ok := rs.values[rule.Name]
		if !ok {
			// and is required
			if rule.HasFlag(IsRequired) {
				return errorCode, errors.New(rule.IsRequiredMessage())
			}
			continue
		}

		// ensure the value matches one of our choices
		if len(rule.Choices) != 0 {
			if !ContainsString(value.value, rule.Choices, nil) {
				return errorCode, fmt.Errorf("'%s' is an invalid argument for '%s' choose from (%s)",
					value.value, rule.Name, strings.Join(rule.Choices, ", "))
			}
		}

		rule.StoreValue(value.value)
	}
	return 0, nil
}
