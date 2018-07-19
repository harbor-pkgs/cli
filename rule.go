package cli

import (
	"context"
	"fmt"
	"regexp"
)

var regexHasPrefix = regexp.MustCompile(`^(\W+)([\w|-]*)$`)

type CastFunc func(string, interface{}, interface{}) (interface{}, error)
type ActionFunc func(*rule, string, []string, *int) error
type StoreFunc func(string) error
type CommandFunc func(context.Context, *Parser) (int, error)

type ruleFlag int64

const (
	IsCommand ruleFlag = 1 << iota
	IsArgument
	IsRequired
	IsFlag
	IsGreedy
	IsExpectingValue
	IsCountFlag
	IsHelpRule
)

type rule struct {
	Sequence    int
	Name        string
	RuleDesc    string
	Value       string
	Default     *string
	Aliases     []string
	EnvVar      string
	Choices     []string
	StoreValue  StoreFunc
	CommandFunc CommandFunc
	flags       ruleFlag
}

func (r *rule) HasFlag(flag ruleFlag) bool {
	return r.flags&flag != 0
}

func (r *rule) SetFlag(flag ruleFlag) {
	r.flags = r.flags | flag
}

func (r *rule) ClearFlag(flag ruleFlag) {
	mask := r.flags ^ flag
	r.flags &= mask
}

func (r *rule) IsRequiredMessage() string {
	switch {
	case r.HasFlag(IsArgument):
		return fmt.Sprintf("argument '%s' is required", r.Name)
	case r.HasFlag(IsFlag):
		return fmt.Sprintf("flag '%s' is required", r.Name)
	}
	return fmt.Sprintf("'%s' is required", r.Name)
}
