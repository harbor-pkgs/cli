package cli

import (
	"context"
	"os"
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

func (e *envStore) Get(ctx context.Context, name string) (string, int) {
	rule := e.rules.GetRule(name)
	if rule == nil {
		return "", 0
	}

	if rule.EnvVar != "" {
		return "", 0
	}

	value := os.Getenv(rule.EnvVar)
	if value == "" {
		return "", 0
	}
	return value, 1
}
