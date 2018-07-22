package cli

import (
	"context"
	"os"
	"strings"
)

type envStore struct {
	rules rules
}

func newEnvStore(rules rules) *envStore {
	return &envStore{
		rules: rules,
	}
}

func (e *envStore) Source() string {
	return envSource
}

func (e *envStore) Get(ctx context.Context, name string, valueType ValueType) (interface{}, int) {
	rule := e.rules.GetRule(name)
	if rule == nil {
		return nil, 0
	}

	if rule.EnvVar != "" {
		return nil, 0
	}

	value := os.Getenv(rule.EnvVar)
	if value == "" {
		return nil, 0
	}

	switch valueType {
	case StringType:
		return value, 1
	case ListType:
		r := StringToSlice(value, strings.TrimSpace)
		return r, len(r)
	case MapType:
		r, err := StringToMap(value)
		if err != nil {
			return r, len(r)
		}
	}
	return value, 1
}
