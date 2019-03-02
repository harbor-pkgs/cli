package cli_test

import (
	"os"
	"sort"
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserNoRules(t *testing.T) {
	p := cli.New(&cli.Config{Mode: cli.NoHelp})
	retCode, err := p.Parse(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "no options or arguments defined; call Add() before calling Parse()", err.Error())
}

func TestParserNoStore(t *testing.T) {
	tests := []struct {
		v   cli.Variant
		err string
	}{
		{v: &cli.Option{Name: "foo"}, err: "refusing to add option 'foo'; provide an 'IsSet', 'Store' or 'Count' field"},
		{v: &cli.Argument{Name: "foo"}, err: "refusing to add argument 'foo'; provide an 'IsSet', 'Store' or 'Count' field"},
		{v: &cli.EnvVar{Name: "foo"}, err: "refusing to add envvar 'foo'; provide an 'IsSet' or 'Store' field"},
	}

	for _, test := range tests {
		p := cli.New(nil)
		p.Add(test.v)
		retCode, err := p.Parse(nil, nil)

		// Then
		assert.NotNil(t, err)
		assert.Equal(t, cli.ErrorRetCode, retCode)
		assert.Contains(t, err.Error(), test.err)
	}
}

func TestParserAddHelpWithOption(t *testing.T) {
	var hasOption bool
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", IsSet: &hasOption})

	// Given -h
	retCode, err := p.Parse(nil, []string{"-h"})

	// Then should return HelpError{}
	require.NotNil(t, err)
	assert.Equal(t, "user asked for help; inspect this error with cli.isHelpError()", err.Error())

	// Should be a help error
	_, ok := err.(*cli.HelpError)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, true, cli.IsHelpError(err))
	assert.Equal(t, true, ok)
}

func TestParserNoArgs(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo})

	// Given no arguments
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "", foo)
}

func TestFooOption(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo})

	// Given double prefix
	retCode, err := p.Parse(nil, []string{"--foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)

	// Given single prefix
	retCode, err = p.Parse(nil, []string{"-foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)
}

func TestOptionExpectedValue(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "expected option '--foo' to have a value", err.Error())
	assert.Equal(t, "", foo)
}

func TestOptionCount(t *testing.T) {
	var count int

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "verbose", Count: &count, Aliases: []string{"v"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--verbose", "-v"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
}

func TestOptionCountWithValue(t *testing.T) {
	var count int
	var foo []string

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar", "-f", "bang"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bang", "bar"}, foo)
}

func TestOptionIsRequired(t *testing.T) {
	var foo, bar string

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Flags: cli.Required, Store: &foo, Aliases: []string{"f"}})
	p.Add(&cli.Option{Name: "bar", Store: &bar, Aliases: []string{"b"}})

	// Given
	retCode, err := p.Parse(nil, []string{"-b", "bar"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "option '--foo' is required", err.Error())
}

func TestOptionReplace(t *testing.T) {
	var count int
	var foo []string

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "foo", Store: &foo, Default: "bash", Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar", "-f", "bang"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bang", "bar"}, foo)
}

// TODO: Test matching flags with no prefix if enabled

func TestBarEnvVar(t *testing.T) {
	var bar, foo, cat string

	p := cli.New(nil)
	p.Add(&cli.EnvVar{Name: "BAR", Store: &bar})
	p.Add(&cli.EnvVar{Name: "foo", Env: "FOO", Store: &foo})
	p.Add(&cli.EnvVar{Name: "cat", Env: "CAT", Store: &cat, Default: "cat-thing"})

	// Given
	os.Setenv("BAR", "bar-thing")
	defer os.Unsetenv("BAR")
	os.Setenv("FOO", "foo-thing")
	defer os.Unsetenv("FOO")
	os.Unsetenv("CAT")

	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
	assert.Equal(t, "foo-thing", foo)
	assert.Equal(t, "cat-thing", cat)
}

// TODO: Test with `Required` restriction for EnvVar

func TestBarArgument(t *testing.T) {
	var bar string

	p := cli.New(nil)
	p.Add(&cli.Argument{Name: "bar", Store: &bar})

	// Given
	retCode, err := p.Parse(nil, []string{"bar-thing"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
}

func TestArgumentOverride(t *testing.T) {
	var bar, foo string

	p := cli.New(nil)
	p.Add(&cli.Argument{Name: "bar", Store: &bar})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"bar-thing", "foo-thing"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
	assert.Equal(t, "foo-thing", foo)

	// Given no arguments
	retCode, err = p.Parse(nil, []string{})

	// Then Parse() should not override current values
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
	assert.Equal(t, "foo-thing", foo)
}

func TestArgumentAndOptions(t *testing.T) {
	var bar, foo, flag string

	p := cli.New(nil)
	p.Add(&cli.Option{Name: "flag", Store: &flag})
	p.Add(&cli.Argument{Name: "bar", Store: &bar})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"bar-thing", "--flag", "flag-thing", "foo-thing"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
	assert.Equal(t, "foo-thing", foo)
	assert.Equal(t, "flag-thing", flag)
}

func TestRuleNameCollision(t *testing.T) {
	var bar, foo, flag string
	p := cli.New(nil)

	// Given
	p.Add(&cli.Option{Name: "bar", Store: &flag})
	p.Add(&cli.Argument{Name: "bar", Store: &bar})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "duplicate argument or option 'bar' defined", err.Error())
}

func TestArgumentIsRequired(t *testing.T) {
	var foo, bar string

	p := cli.New(nil)
	p.Add(&cli.Argument{Name: "foo", Flags: cli.Required, Store: &foo})
	p.Add(&cli.Argument{Name: "bar", Store: &bar, Default: "foo"})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "argument 'foo' is required", err.Error())

	// Given
	retCode, err = p.Parse(nil, []string{"bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, foo, "bar")
	assert.Equal(t, bar, "foo")
}

func TestInvalidRuleNames(t *testing.T) {
	var foo, flag string
	p := cli.New(nil)

	// Given
	p.Add(&cli.Option{Name: "*bar", Store: &flag})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "'*bar' is an invalid name for option; prefixes on names are not allowed", err.Error())
}

func TestPartialMatchArgs(t *testing.T) {
	var foo, fooBar, fooBang, bang string

	p := cli.New(nil)
	p.Add(
		&cli.Option{Name: "foo", Store: &foo},
		&cli.Option{Name: "foo-bang", Store: &fooBang},
		&cli.Option{Name: "foobar", Store: &fooBar},
		&cli.Option{Name: "bang", Store: &bang},
	)

	// Given no value
	retCode, err := p.Parse(nil, []string{
		"--foobar", "one",
		"--foo-bang", "two",
		"--foo", "three",
		"--bang", "four",
	})

	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "one", fooBar)
	assert.Equal(t, "two", fooBang)
	assert.Equal(t, "three", foo)
	assert.Equal(t, "four", bang)

	retCode, err = p.Parse(nil, []string{
		"--foobar", "one",
		"--foo-bang", "two",
		"--foo", "three",
		"--me", "five",
		"--foo-me", "four",
	})
	assert.NotNil(t, err)
	assert.Equal(t, "'--me' was provided but not defined", err.Error())
}

func TestInvalidAliases(t *testing.T) {
	var foo, flag string
	p := cli.New(nil)

	// Given
	p.Add(&cli.Option{Name: "bar", Aliases: []string{"-b"}, Store: &flag})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "'-b' is an invalid alias for option; prefixes on aliases are not allowed", err.Error())
}

func TestUnknownArgs(t *testing.T) {
	var foo string
	var flag bool
	p := cli.New(nil)

	// Given
	p.Add(&cli.Option{Name: "bar", Aliases: []string{"b"}, IsSet: &flag})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{"-b", "-g", "bar"})

	// Then '-g' should not be mistaken for argument 'foo'
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "'-g' was provided but not defined", err.Error())
}

func TestCanRepeat(t *testing.T) {
	var src []string
	var dst string

	p := cli.New(nil)

	// Given
	p.Add(&cli.Argument{Name: "src", Store: &src, Flags: cli.CanRepeat})
	p.Add(&cli.Argument{Name: "dst", Store: &dst})
	retCode, err := p.Parse(nil, []string{"file1", "file2", "file3", "/folder"})

	require.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, []string{"file1", "file2", "file3"}, src)
	assert.Equal(t, "/folder", dst)

	// TODO: Test CanRepeat arguments
	// TODO: Test CanRepeat post and prefix  cp <src> <src> <dst>
}

// TODO: Errors should reference the actual option that caused the issue, not the rule definition name
//  IE: (unexpected duplicate option 'foo' provided") should be (unexpected duplicate option '-f' provided")
// TODO: Test interspersed arguments <arg0> <arg1> <cmd> <arg0>
// TODO: sub command usage should include <command> in usage line
// TODO: Test for flags that start with or contain a number -v3  -2Knds
// TODO: Test docker run -it -P 80:80 <image> <command> <-command-flag> -it -P blah
//  maybe a parser.Add(Argument{Flags: greedy|stop}) to stop processing any args or flags after
//  while scanning flags/args count the number of unknown args, once unknown args is greater than the stop, then
//  ignore flags after.
// TODO: parser.Add(Stop{Marker: "--"}) will instruct that no arg parsing will occur after the "--" marker
