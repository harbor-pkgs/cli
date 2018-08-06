package cli

import (
	"context"
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

func (rs *resultStore) Get(ctx context.Context, name string, valType ValueType) (interface{}, int) {
	value, ok := rs.values[name]
	if ok {
		return value.value, value.count
	}
	return "", 0
}

func (rs *resultStore) Set(name, source string, value interface{}, count int) {
	rs.values[name] = valueSrc{
		source: source,
		value:  value,
		count:  count,
	}
}
