package cli

import (
	"context"
	"os"
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

func (e *envStore) Get(ctx context.Context, name string, flags Flags) (interface{}, int, error) {
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

	return convToKind([]string{value}, flags, 1)
}
