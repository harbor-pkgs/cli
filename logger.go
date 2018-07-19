package cli

var DefaultLogger = &NullLogger{}

// We only need part of the standard logging functions
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

type NullLogger struct{}

func (n *NullLogger) Print(...interface{})          {}
func (n *NullLogger) Printf(string, ...interface{}) {}
func (n *NullLogger) Println(...interface{})        {}
