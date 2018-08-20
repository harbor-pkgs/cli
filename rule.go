package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	isHidden
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

func (r *rule) SetFlag(flag ruleFlag, set bool) {
	if set {
		r.flags = r.flags | flag
	} else {
		mask := r.flags ^ flag
		r.flags &= mask
	}
}

func (r *rule) StoreValue(value interface{}, count int) error {
	for _, f := range r.StoreFuncs {
		if err := f(value, count); err != nil {
			return fmt.Errorf("while storing '%s' %s", value, err)
		}
	}
	return nil
}

func (r *rule) GenerateUsage() string {
	switch {
	case r.HasFlag(isFlag):
		if r.HasFlag(isRequired) {
			return fmt.Sprintf("%s", r.Aliases[0])
		}
		return fmt.Sprintf("[%s]", r.Aliases[0])
	case r.HasFlag(isArgument):
		// TODO: Display CanRepeat and interspersed arguments
		if r.HasFlag(isRequired) {
			return fmt.Sprintf("<%s>", r.Name)
		}
		return fmt.Sprintf("[%s]", r.Name)
	}
	return ""
}

func (r *rule) GenerateHelp() (string, string) {
	var parens []string
	paren := ""

	if !r.HasFlag(isCommand) {
		if r.Default != nil {
			parens = append(parens, fmt.Sprintf("default=%s", *r.Default))
		}
		if r.EnvVar != "" {
			parens = append(parens, fmt.Sprintf("env=%s", r.EnvVar))
		}
		if len(parens) != 0 {
			paren = fmt.Sprintf(" (%s)", strings.Join(parens, ", "))
		}
	}

	var valueType string
	// if the option expects a value optionally display this depending on type
	// TODO: Allow the user to override this?
	if r.HasFlag(isFlag) && r.HasFlag(isExpectingValue) {
		switch {
		case r.HasFlag(isScalar):
			valueType = " <string>"
		case r.HasFlag(isList):
			valueType = " <s1>,<s2>"
		case r.HasFlag(isMap):
			valueType = " <key>=<value>"
		}
	}

	if r.HasFlag(isArgument) {
		return "  " + r.Name, r.HelpMsg
	}

	var flags []string
	for _, flag := range r.Aliases {
		flags = append(flags, fmt.Sprintf("-%s", flag))
	}

	return "  " + strings.Join(flags, ", ") + valueType, r.HelpMsg + paren
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
