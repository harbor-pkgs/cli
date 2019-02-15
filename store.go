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
	Get(context.Context, string, Kind) (interface{}, int, error)
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
		fmt.Printf("[%s] Get(%s, %s) - '%v' '%d'\n", from.Source(), rule.Name, rule.ValueType(), value, count)
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

func (rs *resultStore) Get(ctx context.Context, name string, valType Kind) (interface{}, int, error) {
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

// Given a list of string values, attempt to convert them to a kind
func sliceToKind(values []string, kind Kind, count int) (interface{}, int, error) {
	switch kind {
	case ScalarKind:
		//fmt.Printf("Get Ret: %s, %d\n", values[0], count)
		return values[0], count, nil
	case ListKind:
		// If only one item is provided, it must be a comma separated list
		if count == 1 {
			return ToSlice(values[0]), count, nil
		}
		return values, count, nil
	case MapKind:
		// each string in the list should be a key value pair
		// either in the form `key=value` or `{"key": "value"}`
		r := make(map[string]string)
		for _, value := range values {
			kv, err := ToStringMap(value)
			if err != nil {
				return nil, 0, fmt.Errorf("map conversion: %s", err)
			}
			// Merge the key values for each of the items
			for k, v := range kv {
				r[k] = v
			}
		}
		//fmt.Printf("Get Ret: %s, %d\n", r, count)
		return r, count, nil
	}
	return nil, 0, fmt.Errorf("no such kind '%s'", kind)
}
