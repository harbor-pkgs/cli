package cli_test

import (
	"fmt"
	"sort"
	"strings"
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

func TestParserAddHelpWithFlag(t *testing.T) {
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo"})

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

func TestFlagDefaultScalar(t *testing.T) {
	var foo string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bash"})

	// Given no value
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bash", foo)

	// Given a value
	retCode, err = p.Parse(nil, []string{"--foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)
}

func TestFlagDefaultList(t *testing.T) {
	var foo []string
	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bash,bar,foo"})

	// Given no value
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	sort.Strings(foo)
	assert.Equal(t, []string{"bar", "bash", "foo"}, foo)

	// Given a value
	retCode, err = p.Parse(nil, []string{"--foo", "bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, []string{"bar"}, foo)
}

func TestFlagDefaultMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Default: "bar=foo,foo=bar"})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 0, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
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

func TestFlagWithSlice(t *testing.T) {
	var foo []string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar,bang", "-f", "foo"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bar,bang", "foo"}, foo)
}

func TestFlagWithMap(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar=foo", "-f", "foo=bar"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
}

func TestFlagWithMapAndJSON(t *testing.T) {
	var foo map[string]string
	var count int

	p := cli.New(nil)
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", `{"bar":"foo"}`, "-f", `{"foo": "bar", "bash": "bang"}`})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	require.Contains(t, foo, "bar")
	require.Contains(t, foo, "foo")
	require.Contains(t, foo, "bash")
	assert.Equal(t, foo["bar"], "foo")
	assert.Equal(t, foo["foo"], "bar")
	assert.Equal(t, foo["bash"], "bang")
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

func TestHelpMessage(t *testing.T) {
	var bar, foo, argOne, argTwo string
	var hasFlag bool

	p := cli.New(&cli.Config{
		Name:   "test",
		Desc:   "This is the description of the application",
		Epilog: "Copyright 2018 By Derrick J. Wippler",
	})
	p.Add(&cli.Argument{Name: "arg-one", Required: true, Store: &argOne, Help: "this is a required argument"})
	p.Add(&cli.Argument{Name: "arg-two", Store: &argTwo, Help: "this argument is optional"})

	p.Add(&cli.Flag{Name: "flag", IfExists: &hasFlag, Help: "this is my flag"})
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Aliases: []string{"f"}, Help: "used to store bars"})
	p.Add(&cli.Flag{Name: "bar", Store: &bar, Help: "used to store foos"})

	// Given
	retCode, err := p.Parse(nil, []string{"-h"})

	// Then
	require.NotNil(t, err)
	require.Equal(t, cli.ErrorRetCode, retCode)
	require.Equal(t, true, cli.IsHelpError(err))

	help := p.GenerateHelp()
	fmt.Println(help)

	lines := strings.Split(help, "\n")
	var i int

	compare := func(expected string) {
		require.Equal(t, expected, lines[i])
		i++
	}

	compare("Usage: test [flags]  <arg-one> [arg-two]")
	compare("")
	compare("This is the description of the application")
	compare("")
	compare("Arguments:")
	compare("  arg-one   this is a required argument")
	compare("  arg-two   this argument is optional")
	compare("")
	compare("Flags:")
	compare("  -flag               this is my flag")
	compare("  -foo, -f <string>   used to store bars")
	compare("  -bar <string>       used to store foos")
	compare("  -help, -h           display this help message and exit")
	compare("")
	compare("Copyright 2018 By Derrick J. Wippler")
}
