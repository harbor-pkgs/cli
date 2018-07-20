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
	IsGreedy   bool

	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	HasValue    *bool
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

	if f.IsGreedy {
		r.SetFlag(IsGreedy)
	}

	if f.Int != nil {
		funcs = append(funcs, toInt(f.Int))
	}
	if f.String != nil {
		funcs = append(funcs, toString(f.String))
	}
	if f.Count != nil {
		r.SetFlag(IsGreedy)
		funcs = append(funcs, toCount(f.Count))
	}
	if f.StringSlice != nil {
		funcs = append(funcs, toStringSlice(f.StringSlice))
	}
	if f.IntSlice != nil {
		funcs = append(funcs, toIntSlice(f.IntSlice))
	}
	if f.HasValue != nil {
		funcs = append(funcs, hasValue(f.HasValue))
	}
	if f.Bool != nil {
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
	IsGreedy   bool

	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	HasValue    *bool
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
	if a.IsGreedy {
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
		funcs = append(funcs, toStringSlice(a.StringSlice))
	}
	if a.IntSlice != nil {
		funcs = append(funcs, toIntSlice(a.IntSlice))
	}
	if a.HasValue != nil {
		funcs = append(funcs, hasValue(a.HasValue))
	}
	if a.Bool != nil {
		funcs = append(funcs, toBool(a.Bool))
	}

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
	p.rules = append(p.rules, v.toRule())
}
