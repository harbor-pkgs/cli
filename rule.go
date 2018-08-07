package cli

import (
	"context"
	"fmt"
	"regexp"
)

var regexHasPrefix = regexp.MustCompile(`^(\W+)([\w|-]*)$`)

type CastFunc func(string, interface{}, interface{}) (interface{}, error)
type ActionFunc func(*rule, string, []string, *int) error
type StoreFunc func(interface{}, int) error
type CommandFunc func(context.Context, *Parser) (int, error)

type ruleFlag int64

// TODO: Make these flags private
const (
	isCommand ruleFlag = 1 << iota
	isArgument
	isRequired
	isFlag
	canRepeat
	isExpectingValue
	hasCount
	isHelpRule
	isList
	isMap
	isScalar
)

type rule struct {
	Sequence    int
	Name        string
	HelpMsg     string
	Value       string
	Default     *string
	Aliases     []string
	EnvVar      string
	Choices     []string
	StoreFuncs  []StoreFunc
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

func (r *rule) StoreValue(value interface{}, count int) error {
	for _, f := range r.StoreFuncs {
		if err := f(value, count); err != nil {
			return fmt.Errorf("while storing '%s' %s", value, err)
		}
	}
	return nil
}

func (r *rule) IsRequiredMessage() string {
	switch {
	case r.HasFlag(isArgument):
		return fmt.Sprintf("argument '%s' is required", r.Name)
	case r.HasFlag(isFlag):
		return fmt.Sprintf("flag '--%s' is required", r.Name)
	}
	return fmt.Sprintf("'%s' is required", r.Name)
}
