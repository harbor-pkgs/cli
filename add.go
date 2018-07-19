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

	Int         *int
	String      *string
	Count       *int
	StringSlice []string
	IntSlice    []int
	Bool        *bool
	StoreBool   *bool
}

func (f *Flag) toRule() *rule {
	var funcs []StoreFunc

	if f.Int != nil {
		funcs = append(funcs, toInt(f.Int))
	}
	if f.String != nil {
		funcs = append(funcs, toString(f.String))
	}
	if f.Count != nil {
		funcs = append(funcs, toCount(f.Count))
	}
	if f.StringSlice != nil {
		funcs = append(funcs, toStringSlice(f.StringSlice))
	}
	if f.IntSlice != nil {
		funcs = append(funcs, toIntSlice(f.IntSlice))
	}
	if f.Bool != nil {
		funcs = append(funcs, toBool(f.Bool))
	}
	if f.StoreBool != nil {
		funcs = append(funcs, toStoreBool(f.StoreBool))
	}

	return &rule{
		Name:       f.Name,
		HelpMsg:    f.Help,
		EnvVar:     f.Env,
		Default:    &f.Default,
		StoreFuncs: funcs,
	}
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
	StoreTrue   *bool
	Bool        *bool
}

func (f *Argument) Type() int {
	return argType
}

type Command struct {
}

func (f *Command) Type() int {
	return cmdType
}

func (p *Parser) Add(v Variant) {
	rule = v.toRule()
}
