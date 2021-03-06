package cli

import (
	"bytes"
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

type Flags int64

func (f *Flags) Has(flag Flags) bool {
	return *f&flag != 0
}

func (f *Flags) Set(flag Flags, set bool) {
	if set {
		*f = *f | flag
	} else {
		mask := *f ^ flag
		*f &= mask
	}
}

const (
	allFlags  Flags = 0xFFFFFFFF
	isCommand Flags = 1 << iota
	isArgument
	isOption
	isEnvVar
	isExpectingValue
	isHelpRule // TODO: This should be a generic special case flag
	cmdHandled

	// Public flags
	Required
	CanRepeat
	NoSplit
	Hidden

	// Kind flags
	ScalarKind
	SliceKind
	MapKind
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
	Usage       string
	Flags       Flags
}

func (r *rule) HasFlag(flag Flags) bool {
	return r.Flags&flag != 0
}

func (r *rule) SetFlag(flag Flags, set bool) {
	if set {
		r.Flags = r.Flags | flag
	} else {
		mask := r.Flags ^ flag
		r.Flags &= mask
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
	case r.HasFlag(isOption):
		if r.HasFlag(Required) {
			return fmt.Sprintf("%s", r.Aliases[0])
		}
		return fmt.Sprintf("[%s]", r.Aliases[0])
	case r.HasFlag(isArgument):
		// TODO: Display CanRepeat and interspersed arguments
		if r.HasFlag(Required) {
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
	return r.Usage
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
	if r.HasFlag(isOption) && r.HasFlag(isExpectingValue) {
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
	case r.HasFlag(isOption):
		return fmt.Sprintf("option '--%s' is required", r.Name)
	}
	return fmt.Sprintf("'%s' is required", r.Name)
}

func (r rule) Type() string {
	switch {
	case r.HasFlag(isOption):
		return "option"
	case r.HasFlag(isArgument):
		return "argument"
	case r.HasFlag(isCommand):
		return "command"
	}
	return "unknown"
}

func (r rule) Kind() string {
	switch {
	case r.HasFlag(ScalarKind):
		return "ScalarKind"
	case r.HasFlag(SliceKind):
		return "SliceKind"
	case r.HasFlag(MapKind):
		return "MapKind"
	}
	return "unknown"
}
