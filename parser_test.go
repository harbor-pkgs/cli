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
	assert.Equal(t, "no flags or arguments defined; call Add() before calling Parse()", err.Error())
}

func TestParserNoStore(t *testing.T) {
	tests := []struct {
		v   cli.Variant
		err string
	}{
		{v: &cli.Flag{Name: "foo"}, err: "refusing to add flag 'foo'; provide an 'IsSet', 'Store' or 'Count' field"},
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

func TestParserAddHelpWithFlag(t *testing.T) {
	var hasFlag bool
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", IsSet: &hasFlag})

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
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

	// Given no arguments
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "", foo)
}

func TestFooFlag(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

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

func TestFlagExpectedValue(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "expected flag '--foo' to have a value", err.Error())
	assert.Equal(t, "", foo)
}

func TestFlagCount(t *testing.T) {
	var count int

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "verbose", Count: &count, Aliases: []string{"v"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--verbose", "-v"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
}

func TestFlagCountWithValue(t *testing.T) {
	var count int
	var foo []string

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar", "-f", "bang"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bang", "bar"}, foo)
}

func TestFlagIsRequired(t *testing.T) {
	var foo, bar string

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Required: true, Store: &foo, Aliases: []string{"f"}})
	p.Add(&cli.Flag{Name: "bar", Store: &bar, Aliases: []string{"b"}})

	// Given
	retCode, err := p.Parse(nil, []string{"-b", "bar"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "flag '--foo' is required", err.Error())
}


func TestFlagReplace(t *testing.T) {
	var count int
	var foo []string

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bash", Count: &count, Aliases: []string{"f"}})

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

func TestBarFooArguments(t *testing.T) {
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

	// Given
	retCode, err = p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "", bar)
	assert.Equal(t, "", foo)
}

func TestArgumentAndFlags(t *testing.T) {
	var bar, foo, flag string

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "flag", Store: &flag})
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
	p.Add(&cli.Flag{Name: "bar", Store: &flag})
	p.Add(&cli.Argument{Name: "bar", Store: &bar})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "duplicate argument or flag 'bar' defined", err.Error())
}

func TestArgumentIsRequired(t *testing.T) {
	var foo, bar string

	p := cli.New(nil)
	p.Add(&cli.Argument{Name: "foo", Required: true, Store: &foo})
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
	p.Add(&cli.Flag{Name: "*bar", Store: &flag})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "'*bar' is an invalid name for flag; prefixes on names are not allowed", err.Error())
}

func TestInvalidAliases(t *testing.T) {
	var foo, flag string
	p := cli.New(nil)

	// Given
	p.Add(&cli.Flag{Name: "bar", Aliases: []string{"-b"}, Store: &flag})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "'-b' is an invalid alias for flag; prefixes on aliases are not allowed", err.Error())
}

// TODO: Test interspersed arguments <arg0> <arg1> <cmd> <arg0>
// TODO: Test CanRepeat arguments
// TODO: Test CanRepeat post and prefix  cp <src> <src> <dst>
// TODO: sub command usage should include <command> in usage line
// TODO: Test for flags that start with or contain a number -v3  -2Knds
