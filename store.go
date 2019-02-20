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
	Get(context.Context, string, Flags) (interface{}, int, error)
}

func newResultStore(rules ruleList) *resultStore {
	return &resultStore{
		values: make(map[string]valueSrc),
		rules:  rules,
	}
}

func (rs *resultStore) From(ctx context.Context, from FromStore) error {
	for _, r := range rs.rules {
		value, count, err := from.Get(ctx, r.Name, r.Flags)
		if err != nil {
			return err
		}

		fmt.Printf("[%s] Get(%s, %s) - '%v' '%d'\n", from.Source(), r.Name, r.Kind(), value, count)

		// If store did not provide a value for this rule
		if count == 0 {
			continue
		}

		// Validate the 'value' is of the correct kind, This protects
		// `StoreFunc` from receiving the incorrect kind.
		if notValidKind(r, value) {
			return fmt.Errorf("value for '%s' from '%s'; expected 'string' type but got '%T'",
				r.Name, from.Source(), value)
		}

		rs.values[r.Name] = valueSrc{
			source: from.Source(),
			count:  count,
			value:  value,
		}
	}
	return nil
}

func notValidKind(r *rule, value interface{}) bool {
	switch {
	case r.HasFlag(ScalarKind):
		if _, ok := value.(string); !ok {
			return true
		}
	case r.HasFlag(SliceKind):
		if _, ok := value.([]string); !ok {
			return true
		}
	case r.HasFlag(MapKind):
		if _, ok := value.(map[string]string); !ok {
			return true
		}
	}
	return false
}

func (rs *resultStore) Get(ctx context.Context, name string, flags Flags) (interface{}, int, error) {
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
