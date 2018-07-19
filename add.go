package cli

const (
	flagType int = 1 << iota
	argType
	cmdType
)

type Variant interface {
	Type() int
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
	StoreTrue   *bool
	Bool        *bool
}

func (f *Flag) Type() int {
	return flagType
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

}
