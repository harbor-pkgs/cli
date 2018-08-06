package cli

import (
	"fmt"
	"reflect"
	"runtime"
)

type Variant interface {
	toRule() (*rule, error)
}

type Flag struct {
	Name       string
	Help       string
	Env        string
	Default    string
	Aliases    []string
	IsRequired bool
	CanRepeat  bool
	IsHelpFlag bool
	DependsOn  string

	Store       interface{}
	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	IfExists    *bool
	Bool        *bool
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
	r.SetFlag(isFlag)

	if f.Store != nil {
		r.SetFlag(isExpectingValue)
		fnc, err := newStoreFunc(f.Store)
		if err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding flag '%s': %s", f.Name, err)
		}
		r.StoreFuncs = append(r.StoreFuncs, fnc)
	}

	if f.Default != "" {
		r.Default = &f.Default
	}
	if f.IsRequired {
		r.SetFlag(isRequired)
	}
	if f.CanRepeat {
		r.SetFlag(canRepeat)
	}
	if f.IsHelpFlag {
		r.SetFlag(isHelpRule)
	}
	if f.Count != nil {
		r.SetFlag(canRepeat)
		r.StoreFuncs = append(r.StoreFuncs, toCount(f.Count))
	}
	if f.IfExists != nil {
		r.StoreFuncs = append(r.StoreFuncs, toExists(f.IfExists))
	}
	return r, nil
}

type Argument struct {
	Name       string
	Help       string
	Env        string
	Default    string
	IsRequired bool
	CanRepeat  bool

	Store       interface{}
	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	IfExists    *bool
	Bool        *bool
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
		fnc, err := newStoreFunc(a.Store)
		if err != nil {
			return nil, fmt.Errorf("invalid 'Store' while adding argument '%s': %s", a.Name, err)
		}
		r.StoreFuncs = append(r.StoreFuncs, fnc)
	}
	if a.Default != "" {
		r.Default = &a.Default
	}
	r.SetFlag(isArgument)
	if a.IsRequired {
		r.SetFlag(isRequired)
	}
	if a.CanRepeat {
		r.SetFlag(canRepeat)
	}
	if a.Count != nil {
		r.SetFlag(canRepeat)
		r.StoreFuncs = append(r.StoreFuncs, toCount(a.Count))
	}
	if a.IfExists != nil {
		r.StoreFuncs = append(r.StoreFuncs, toExists(a.IfExists))
	}

	return r, nil
}

func newStoreFunc(dest interface{}) (StoreFunc, error) {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("cannot use non pointer type '%s'; must provide a pointer", d.Kind())
	}
	d = reflect.Indirect(d)
	switch d.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		elem := reflect.TypeOf(dest).Elem()
		switch elem.Kind() {
		case reflect.Int:
			return toIntSlice(dest.([]int)), nil
		case reflect.String:
			return toStringSlice(dest.([]string)), nil
		default:
			return nil, fmt.Errorf("slice of type '%s' is not supported", d.Kind())
		}
	case reflect.String:
		fmt.Println("isString")
		return toString(dest.(*string)), nil
	case reflect.Bool:
		return toBool(dest.(*bool)), nil
	case reflect.Int:
		return toInt(dest.(*int)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32,
		reflect.Float64, reflect.Interface, reflect.Ptr, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return nil, fmt.Errorf("cannot use '%s'; type not supported", d.Kind())
	}
	return nil, fmt.Errorf("unhandled type '%s'", d.Kind())
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
	r.SetFlag(isCommand)
	return r
}

func (p *Parser) Add(v Variant) {
	// TODO: Support adding multiple variants with the same Add() call
	rule, err := v.toRule()
	if err != nil {
		// Extract the line number and file name that called 'Add'
		_, file, line, _ := runtime.Caller(1)
		// Add the error to the parser, to be reported when `Parse()` is called
		p.errs = append(p.errs, fmt.Errorf("%s:%d - %s", file, line, err))
		return
	}

	fmt.Println("add rule")
	p.rules = append(p.rules, rule)
}

// Returns the file, function and line number of the function that called logrus
/*func GetLogrusCaller() *stack.FrameInfo {
	var frames [32]uintptr

	// iterate until we find non logrus function
	length := runtime.Callers(5, frames[:])
	for idx := 0; idx < (length - 1); idx++ {
		pc := uintptr(frames[idx]) - 1
		fn := runtime.FuncForPC(pc)
		funcName := fn.Name()
		if strings.Contains(strings.ToLower(funcName), "sirupsen/logrus") {
			continue
		}
		filePath, lineNo := fn.FileLine(pc)
		return &stack.FrameInfo{
			Func:   stack.FuncName(fn),
			File:   filePath,
			LineNo: lineNo,
		}
	}
	return &stack.FrameInfo{}
}*/
