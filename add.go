package cli

type Variant interface {
	toRule() *rule
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

	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	IfExists    *bool
	Bool        *bool
}

func (f *Flag) toRule() *rule {
	var funcs []StoreFunc

	r := &rule{
		Name:    f.Name,
		HelpMsg: f.Help,
		EnvVar:  f.Env,
		Default: &f.Default,
	}

	r.SetFlag(IsFlag)
	if f.IsRequired {
		r.SetFlag(IsRequired)
	}

	if f.CanRepeat {
		r.SetFlag(IsGreedy)
	}

	if f.IsHelpFlag {
		r.SetFlag(IsHelpRule)
	}

	if f.Int != nil {
		r.SetFlag(IsExpectingValue)
		funcs = append(funcs, toInt(f.Int))
	}
	if f.String != nil {
		r.SetFlag(IsExpectingValue)
		funcs = append(funcs, toString(f.String))
	}
	if f.Count != nil {
		r.SetFlag(IsGreedy)
		funcs = append(funcs, toCount(f.Count))
	}
	if f.StringSlice != nil {
		r.SetFlag(IsExpectingValue)
		funcs = append(funcs, toStringSlice(f.StringSlice))
	}
	if f.IntSlice != nil {
		r.SetFlag(IsExpectingValue)
		funcs = append(funcs, toIntSlice(f.IntSlice))
	}
	if f.IfExists != nil {
		funcs = append(funcs, toExists(f.IfExists))
	}
	if f.Bool != nil {
		r.SetFlag(IsExpectingValue)
		funcs = append(funcs, toBool(f.Bool))
	}

	r.StoreFuncs = funcs

	return r
}

type Argument struct {
	Name       string
	Help       string
	Env        string
	Default    string
	Aliases    []string
	IsRequired bool
	CanRepeat  bool

	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	IfExists    *bool
	Bool        *bool
}

func (a *Argument) toRule() *rule {
	var funcs []StoreFunc

	r := &rule{
		Name:       a.Name,
		HelpMsg:    a.Help,
		EnvVar:     a.Env,
		Default:    &a.Default,
		StoreFuncs: funcs,
	}

	r.SetFlag(IsArgument)
	if a.IsRequired {
		r.SetFlag(IsRequired)
	}
	if a.CanRepeat {
		r.SetFlag(IsGreedy)
	}

	if a.Int != nil {
		funcs = append(funcs, toInt(a.Int))
	}
	if a.String != nil {
		funcs = append(funcs, toString(a.String))
	}
	if a.Count != nil {
		r.SetFlag(IsGreedy)
		funcs = append(funcs, toCount(a.Count))
	}
	if a.StringSlice != nil {
		r.SetFlag(IsList)
		funcs = append(funcs, toStringSlice(a.StringSlice))
	}
	if a.IntSlice != nil {
		r.SetFlag(IsList)
		funcs = append(funcs, toIntSlice(a.IntSlice))
	}
	if a.IfExists != nil {
		funcs = append(funcs, toExists(a.IfExists))
	}
	if a.Bool != nil {
		funcs = append(funcs, toBool(a.Bool))
	}

	// TODO: Support map
	r.StoreFuncs = funcs

	return r
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
	r.SetFlag(IsCommand)
	return r
}

func (p *Parser) Add(v Variant) {
	// TODO: Support adding multiple variants with the same Add() call
	p.rules = append(p.rules, v.toRule())
}
