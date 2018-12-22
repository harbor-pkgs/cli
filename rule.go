package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"bytes"
)

var regexHasPrefix = regexp.MustCompile(`^(\W+)([\w|-]*)$`)

type CastFunc func(string, interface{}, interface{}) (interface{}, error)
type ActionFunc func(*rule, string, []string, *int) error
type StoreFunc func(interface{}, int) error
type CommandFunc func(context.Context, *Parser) (int, error)

type ruleFlag int64

// TODO: Make these flags private
const (
	allFlags  ruleFlag = 0xFFFFFFFF
	isCommand ruleFlag = 1 << iota
	isArgument
	isRequired
	isFlag
	isEnvVar
	canRepeat
	isExpectingValue
	isHidden
	hasCount
	isHelpRule

	// Type flags
	isList
	isMap
	isScalar

	// Kind flags
	isString
	isInt
	isBool
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
			return err
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

// Generate usage lines suitable for use in an example config
//
//   # The count of things to come (Default:"1")
//   # export COUNT=<int>
func (r *rule) GenerateEnvUsage(wordWrap int) []byte {
	result := r.generateConfigHelp(wordWrap)
	result.WriteString("# " + "export " + r.EnvVar + "=" + r.TypeUsage() + "\n\n")
	return result.Bytes()
}

// Generate usage lines suitable for use in an example INI config
//
//   # The count of things to come (Default:"1")
//   # count=<int>
func (r *rule) GenerateINIUsage(wordWrap int) []byte {
	result := r.generateConfigHelp(wordWrap)
	result.WriteString("# " + r.Name + "=" + r.TypeUsage() + "\n\n")
	return result.Bytes()
}

// Generate help lines suitable for use in an example config
//
//   # The count of things to come (Default:"1")
func (r *rule) generateConfigHelp(wordWrap int) bytes.Buffer {
	var result bytes.Buffer

	// Append the default value to the end of the help string
	helpMsg := r.HelpMsg
	if r.Default != nil {
		helpMsg += ` (Default:"` + *r.Default + `")`
	}

	// Word wrap the help string
	for _, line := range strings.Split(WordWrap(helpMsg, 0, wordWrap), "\n") {
		result.WriteString("# " + line + "\n")
	}
	return result
}

func (r *rule) TypeUsage() string {
	switch {
	case r.HasFlag(isScalar):
		switch {
		case r.HasFlag(isString):
			return "<string>"
		case r.HasFlag(isBool):
			return "<bool>"
		case r.HasFlag(isInt):
			return "<int>"
		default:
			return "<unknown>"
		}
	case r.HasFlag(isList):
		switch {
		case r.HasFlag(isString):
			return "<str>,<str>"
		case r.HasFlag(isBool):
			return "<bool>,<bool>"
		case r.HasFlag(isInt):
			return "<int>,<int>"
		default:
			return "<unknown>,<unknown>"
		}
	case r.HasFlag(isMap):
		switch {
		case r.HasFlag(isString):
			return "<key>=<string>"
		default:
			return "<key>=<unknown>"
		}
	}
	return "<unknown>"
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
	// TODO: Allow the user to override this when custom type provides Get() interface
	if r.HasFlag(isFlag) && r.HasFlag(isExpectingValue) {
		valueType = " " + r.TypeUsage()
	}

	if r.HasFlag(isArgument) {
		return "  " + r.Name, r.HelpMsg
	}

	if r.HasFlag(isEnvVar) {
		return "  " + r.Name + " " + r.TypeUsage(), r.HelpMsg
	}

	var flags []string
	for _, flag := range r.Aliases {
		if len(flag) > 2 {
			flags = append(flags, fmt.Sprintf("--%s", flag))
		} else {
			flags = append(flags, fmt.Sprintf("-%s", flag))
		}
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
