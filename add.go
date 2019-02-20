package cli

import (
	"fmt"
	"runtime"
)

type Variant interface {
	name() string
	toRule() (*rule, error)
}

// Any struct the implements this interface can be used by `Add()` to
// store the value for a option or argument.
// TODO: Reference an example
type SetValue interface {
	Set(string) error
}

type Option struct {
	Name      string
	Help      string
	Env       string
	Default   string
	Aliases   []string
	Flags     Flags
	DependsOn string // TODO: Implement dependency

	Store interface{}

	/*
		// Scalars
		String *string
		Int    *int
		Bool   *bool

		// Slices
		StringSlice *[]string
		IntSlice    *[]int
		BoolSlice   *[]bool

		// Maps
		StringMap *map[string]string
		IntMap    *map[string]int
		BoolMap   *map[string]bool

		// Set interface
		Set SetValue
	*/

	// Informational
	Count *int
	IsSet *bool
}

func (f *Option) name() string {
	return f.Name
}

func (f *Option) toRule() (*rule, error) {
	if f.Name == "" {
		return nil, fmt.Errorf("failed to add new option; 'Name' is required")
	}

	r := &rule{
		Name:    f.Name,
		HelpMsg: f.Help,
		Aliases: append(f.Aliases, f.Name),
		EnvVar:  f.Env,
		Flags:   f.Flags,
	}

	if f.Store != nil {
		r.SetFlag(isExpectingValue, true)
		if err := newStoreFunc(r, f.Store); err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding option '%s': %s", f.Name, err)
		}
	}

	if f.Default != "" {
		r.Default = &f.Default
	}

	r.SetFlag(isOption, true)

	if f.Count != nil {
		r.SetFlag(CanRepeat, true)
		r.StoreFuncs = append(r.StoreFuncs, toCount(f.Count))
	}
	if f.IsSet != nil {
		r.StoreFuncs = append(r.StoreFuncs, toSet(f.IsSet))
	}

	// TODO: Should check for a StoreFunc() instead
	if f.IsSet == nil && f.Store == nil && f.Count == nil {
		return nil, fmt.Errorf("refusing to add option '%s'; provide an 'IsSet', 'Store' or 'Count' field", f.Name)
	}
	return r, nil
}

type Argument struct {
	Name    string
	Help    string
	Env     string
	Default string
	Flags   Flags

	Store interface{}
	Count *int
	IsSet *bool
}

func (a *Argument) name() string {
	return a.Name
}

func (a *Argument) toRule() (*rule, error) {
	if a.Name == "" {
		return nil, fmt.Errorf("failed to add new argument; 'Name' is required")
	}

	r := &rule{
		Name:    a.Name,
		HelpMsg: a.Help,
		EnvVar:  a.Env,
		Flags:   a.Flags,
	}

	if a.Store != nil {
		if err := newStoreFunc(r, a.Store); err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding argument '%s': %s", a.Name, err)
		}
	}

	if a.Default != "" {
		r.Default = &a.Default
	}
	r.SetFlag(isArgument, true)

	if a.Count != nil {
		// TODO: Test can repeat for args
		r.SetFlag(CanRepeat, true)
		r.StoreFuncs = append(r.StoreFuncs, toCount(a.Count))
	}
	if a.IsSet != nil {
		r.StoreFuncs = append(r.StoreFuncs, toSet(a.IsSet))
	}

	if a.IsSet == nil && a.Store == nil && a.Count == nil {
		return nil, fmt.Errorf("refusing to add argument '%s'; provide an 'IsSet', 'Store' or 'Count' field", a.Name)
	}

	return r, nil
}

type EnvVar struct {
	Name    string
	Help    string
	Env     string
	Default string
	Flags   Flags // TODO: Test required for env

	Store interface{}
	IsSet *bool
}

func (e *EnvVar) name() string {
	return e.Name
}

func (e *EnvVar) toRule() (*rule, error) {
	if e.Name == "" {
		return nil, fmt.Errorf("failed to add new EnvVar; 'Name' is required")
	}

	r := &rule{
		Name:    e.Name,
		HelpMsg: e.Help,
		EnvVar:  e.Env,
		Flags:   e.Flags,
	}

	if r.EnvVar == "" {
		r.EnvVar = e.Name
	}

	if e.Store != nil {
		if err := newStoreFunc(r, e.Store); err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding EnvVar '%s': %s", e.Name, err)
		}
	}

	if e.Default != "" {
		r.Default = &e.Default
	}
	r.SetFlag(isExpectingValue, true)
	r.SetFlag(isEnvVar, true)

	if e.IsSet != nil {
		r.StoreFuncs = append(r.StoreFuncs, toSet(e.IsSet))
	}

	if e.IsSet == nil && e.Store == nil {
		return nil, fmt.Errorf("refusing to add envvar '%s'; provide an 'IsSet' or 'Store' field", e.Name)
	}

	return r, nil
}

type Command struct {
	Name string
	Help string
	Func CommandFunc
}

func (a *Command) toRule() *rule {
	r := &rule{
		Name:        subCmdNamePrefix + a.Name,
		HelpMsg:     a.Help,
		CommandFunc: a.Func,
	}
	r.SetFlag(isCommand, true)
	return r
}

func (p *Parser) Add(variants ...Variant) {
	for _, v := range variants {
		rule, err := v.toRule()
		if err != nil {
			// Extract the line number and file name that called 'Add'
			_, file, line, _ := runtime.Caller(1)
			// Add the error to the parser, to be reported when `Parse()` is called
			p.errs = append(p.errs, fmt.Errorf("%s:%d - %s", file, line, err))
			return
		}
		rule.Sequence = p.seqCount
		p.seqCount++

		// Ensure arguments and commands are at the bottom of the rules when sorted by sequence
		// TODO: Sorting the rules might not matter anymore, find out
		if rule.HasFlag(isArgument | isCommand) {
			rule.Sequence += 10000
		}

		fmt.Printf("Add(%s)\n", rule.Name)
		p.rules = append(p.rules, rule)
	}
}

func (p *Parser) Replace(variants ...Variant) error {
	for _, v := range variants {
		idx := p.rules.GetRuleIndex(v.name())
		if idx == -1 {
			return fmt.Errorf("unable to replace '%s'; not found", v.name())
		}
		// Add the new rule
		rule, err := v.toRule()
		if err != nil {
			// Extract the line number and file name that called 'Add'
			_, file, line, _ := runtime.Caller(1)
			// Add the error to the parser, to be reported when `Parse()` is called
			p.errs = append(p.errs, fmt.Errorf("%s:%d - %s", file, line, err))
			return nil
		}
		fmt.Printf("Replace(%s)\n", rule.Name)

		// Preserve the sequence and replace the rule
		rule.Sequence = p.rules[idx].Sequence
		p.rules[idx] = rule

		// Any previously parsed syntax is now invalid
		p.syntax = nil
	}
	return nil
}
