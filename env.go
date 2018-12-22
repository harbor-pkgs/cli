package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type envStore struct {
	rules ruleList
}

func newEnvStore(rules ruleList) *envStore {
	return &envStore{
		rules: rules,
	}
}

func (e *envStore) Source() string {
	return envSource
}

func (e *envStore) Get(ctx context.Context, name string, valueType ValueType) (interface{}, int, error) {
	rule := e.rules.GetRule(name)
	if rule == nil {
		return nil, 0, nil
	}

	if rule.EnvVar == "" {
		return nil, 0, nil
	}

	value := os.Getenv(rule.EnvVar)
	if value == "" {
		return nil, 0, nil
	}

	switch valueType {
	case ScalarType:
		return value, 1, nil
	case ListType:
		r := StringToSlice(value, strings.TrimSpace)
		return r, len(r), nil
	case MapType:
		r, err := ToStringMap(value)
		if err != nil {
			return r, len(r), nil
		}
	}
	return value, 1, fmt.Errorf("unknown ValueType '%s'", valueType)
}
