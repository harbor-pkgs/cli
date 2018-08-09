package cli

// Returns true if the error was because help flag was found during parsing
func IsHelpError(err error) bool {
	obj, ok := err.(isHelpError)
	return ok && obj.IsHelpError()
}

type isHelpError interface {
	IsHelpError() bool
}

type HelpError struct{}

func (e *HelpError) Error() string {
	return "user asked for help; inspect this error with cli.isHelpError()"
}

func (e *HelpError) IsHelpError() bool {
	return true
}

func IsInvalidFlag(err error) bool {
	obj, ok := err.(isHelpError)
	return ok && obj.IsHelpError()
}

type isInvalidFlag interface {
	IsInvalidFlag() bool
}

type InvalidFlag struct {
	Msg string
}

func (e *InvalidFlag) Error() string {
	return e.Msg
}

func (e *InvalidFlag) IsInvalidFlag() bool {
	return true
}
