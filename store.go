package cli

import (
	"context"
)

type valueSrc struct {
	count  int
	source string
	value  string
}

type resultStore struct {
	rules  rules
	values map[string]valueSrc
}

type FromStore interface {
	Source() string
	Get(context.Context, string) (string, int)
}

func newResultStore(rules rules) *resultStore {
	return &resultStore{
		rules: rules,
	}
}

func (rs *resultStore) From(ctx context.Context, from FromStore) error {
	for _, rule := range rs.rules {
		if value, count := from.Get(ctx, rule.Name); count != 0 {
			rs.values[rule.Name] = valueSrc{
				source: from.Source(),
				count:  count,
				value:  value,
			}
		}
	}
	return nil
}

func (rs *resultStore) Get(ctx context.Context, name string) (string, int) {
	value, ok := rs.values[name]
	if ok {
		return value.value, value.count
	}
	return "", 0
}

func (rs *resultStore) Set(name, source, value string, count int) {
	rs.values[name] = valueSrc{
		source: source,
		value:  value,
		count:  count,
	}
}
