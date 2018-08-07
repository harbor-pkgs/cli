package cli

import (
	"context"
	"fmt"
)

type valueSrc struct {
	count  int
	source string
	value  interface{}
}

type resultStore struct {
	rules  ruleList
	values map[string]valueSrc
}

type FromStore interface {
	Source() string
	// Returns the value and the number of times the value was seen
	// in the store (not a count of the values found if non scalar value)
	Get(context.Context, string, ValueType) (interface{}, int, error)
}

func newResultStore(rules ruleList) *resultStore {
	return &resultStore{
		values: make(map[string]valueSrc),
		rules:  rules,
	}
}

func (rs *resultStore) From(ctx context.Context, from FromStore) error {
	for _, rule := range rs.rules {
		value, count, err := from.Get(ctx, rule.Name, rule.ValueType())
		fmt.Printf("[%s]Get(%s, %s) - '%v' '%d'\n", from.Source(), rule.Name, rule.ValueType(), value, count)
		if err != nil {
			return err
		}
		if count != 0 {
			rs.values[rule.Name] = valueSrc{
				source: from.Source(),
				count:  count,
				value:  value,
			}
		}
	}
	return nil
}

func (rs *resultStore) Get(ctx context.Context, name string, valType ValueType) (interface{}, int, error) {
	value, ok := rs.values[name]
	if ok {
		return value.value, value.count, nil
	}
	return "", 0, nil
}

func (rs *resultStore) Set(name, source string, value interface{}, count int) {
	rs.values[name] = valueSrc{
		source: source,
		value:  value,
		count:  count,
	}
}
