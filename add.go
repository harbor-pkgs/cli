package cli

import (
	"fmt"
	"reflect"
	"runtime"
)

type Variant interface {
	name() string
	toRule() (*rule, error)
}

type Flag struct {
	Name      string
	Help      string
	Env       string
	Default   string
	Aliases   []string
	Required  bool
	CanRepeat bool
	HelpFlag  bool
	DependsOn string
	Hidden    bool

	Store interface{}
	Count *int
	IsSet *bool
}

func (f *Flag) name() string {
	return f.Name
}

func (f *Flag) toRule() (*rule, error) {
	if f.Name == "" {
		return nil, fmt.Errorf("failed to add new flag; 'Name' is required")
	}

	r := &rule{
		Name:    f.Name,
		HelpMsg: f.Help,
		Aliases: append(f.Aliases, f.Name),
		EnvVar:  f.Env,
	}

	if f.Store != nil {
		r.SetFlag(isExpectingValue, true)
		fnc, flag, err := newStoreFunc(f.Store)
		if err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding flag '%s': %s", f.Name, err)
		}
		r.SetFlag(flag, true)
		r.StoreFuncs = append(r.StoreFuncs, fnc)
	}

	if f.Default != "" {
		r.Default = &f.Default
	}

	r.SetFlag(isFlag, true)
	r.SetFlag(isHidden, f.Hidden)
	r.SetFlag(isRequired, f.Required)
	r.SetFlag(canRepeat, f.CanRepeat)
	r.SetFlag(isHelpRule, f.HelpFlag)

	if f.Count != nil {
		r.SetFlag(canRepeat, true)
		r.StoreFuncs = append(r.StoreFuncs, toCount(f.Count))
	}
	if f.IsSet != nil {
		r.StoreFuncs = append(r.StoreFuncs, toSet(f.IsSet))
	}

	if f.IsSet == nil && f.Store == nil && f.Count == nil {
		return nil, fmt.Errorf("refusing to add flag '%s'; provide an 'IsSet', 'Store' or 'Count' field", f.Name)
	}
	return r, nil
}

type Argument struct {
	Name      string
	Help      string
	Env       string
	Default   string
	Required  bool
	CanRepeat bool

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
	}

	if a.Store != nil {
		fnc, flag, err := newStoreFunc(a.Store)
		if err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding argument '%s': %s", a.Name, err)
		}
		r.SetFlag(flag, true)
		r.StoreFuncs = append(r.StoreFuncs, fnc)
	}

	if a.Default != "" {
		r.Default = &a.Default
	}
	r.SetFlag(isArgument, true)
	r.SetFlag(isRequired, a.Required)
	r.SetFlag(canRepeat, a.CanRepeat)

	if a.Count != nil {
		r.SetFlag(canRepeat, true)
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
	Name     string
	Help     string
	Env      string
	Default  string
	Required bool

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
	}

	if r.EnvVar == "" {
		r.EnvVar = e.Name
	}

	if e.Store != nil {
		fnc, flag, err := newStoreFunc(e.Store)
		if err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding EnvVar '%s': %s", e.Name, err)
		}
		r.SetFlag(flag, true)
		r.StoreFuncs = append(r.StoreFuncs, fnc)
	}

	if e.Default != "" {
		r.Default = &e.Default
	}
	r.SetFlag(isRequired, e.Required)
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

// TODO: Support the 'Set()' interface similar to go flags
func newStoreFunc(dest interface{}) (StoreFunc, ruleFlag, error) {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		return nil, isScalar, fmt.Errorf("cannot use non pointer type '%s'; must provide a pointer", d.Kind())
	}

	// Dereference the pointer
	d = reflect.Indirect(d)

	// Determine if it's a scalar, slice or map
	switch d.Kind() {

	case reflect.Array, reflect.Slice:
		elem := reflect.TypeOf(dest).Elem().Elem()
		switch elem.Kind() {
		case reflect.Int:
			return toIntSlice(dest.(*[]int)), isList | isInt, nil
		case reflect.String:
			return toStringSlice(dest.(*[]string)), isList | isString, nil
		case reflect.Bool:
			return toBoolSlice(dest.(*[]bool)), isList | isBool, nil
		default:
			return nil, isList, fmt.Errorf("slice of type '%s' is not supported", elem.Kind())
		}

	case reflect.Map:
		key := d.Type().Key()
		elem := d.Type().Elem()
		if key.Kind() == reflect.String && elem.Kind() == reflect.String {
			return toStringMap(dest.(*map[string]string)), isMap | isString, nil
		}
		if key.Kind() == reflect.String && elem.Kind() == reflect.Int {
			return toIntMap(dest.(*map[string]int)), isMap | isInt, nil
		}
		if key.Kind() == reflect.String && elem.Kind() == reflect.Bool {
			return toBoolMap(dest.(*map[string]bool)), isMap | isBool, nil
		}
		return nil, isMap, fmt.Errorf("cannot use 'map[%s]%s'; only "+
			"'map[string]string, map[string]int, map[string]bool' currently supported", key.Kind(), elem.Kind())

	case reflect.String:
		return toString(dest.(*string)), isScalar | isString, nil

	case reflect.Bool:
		return toBool(dest.(*bool)), isScalar | isBool, nil

	case reflect.Int:
		return toInt(dest.(*int)), isScalar | isInt, nil

	// Unhandled types
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
		reflect.Float64, reflect.Interface, reflect.Ptr, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil, isScalar, fmt.Errorf("cannot store '%s'; type not supported", d.Kind())
	}
	return nil, isScalar, fmt.Errorf("unhandled type '%s'", d.Kind())
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
