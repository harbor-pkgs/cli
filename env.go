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

func (e *envStore) Get(ctx context.Context, name string) (string, bool) {
	rule := e.rules.GetRule(name)
	if rule == nil {
		return "", false
	}

	if rule.EnvVar != "" {
		return "", false
	}

	value := os.Getenv(rule.EnvVar)
	if value == "" {
		return "", false
	}
	return value, true
}
