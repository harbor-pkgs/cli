package cli_test

import (
	"sort"
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserNoRules(t *testing.T) {
	p := cli.New(&cli.Parser{NoHelp: true})
	retCode, err := p.Parse(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "no flags or arguments defined; call Add() before calling Parse()", err.Error())
}

func TestParserAddHelpWithFlag(t *testing.T) {
	p := cli.NewParser()
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
	p := cli.NewParser()
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
	p := cli.NewParser()
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
	p := cli.NewParser()
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

	p := cli.NewParser()
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
	p := cli.NewParser()
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
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo"})

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "expected flag '--foo' to have an argument", err.Error())
	assert.Equal(t, "", foo)
}

func TestFlagCount(t *testing.T) {
	var count int

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
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

	p := cli.NewParser()
	p.Add(&cli.Argument{Name: "bar", Store: &bar})

	// Given
	retCode, err := p.Parse(nil, []string{"bar-thing"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
}

func TestBarFooArguments(t *testing.T) {
	var bar string
	var foo string

	p := cli.NewParser()
	p.Add(&cli.Argument{Name: "bar", Store: &bar})
	p.Add(&cli.Argument{Name: "foo", Store: &foo})

	// Given
	retCode, err := p.Parse(nil, []string{"bar-thing", "foo-thing"})

	// Then
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar-thing", bar)
	assert.Equal(t, "foo-thing", foo)
}

func TestArgumentAndFlags(t *testing.T) {
	var bar string
	var foo string
	var flag string

	p := cli.NewParser()
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

// TODO: Test interspersed arguments <arg0> <arg1> <cmd> <arg0>
// TODO: Test CanRepeat post and prefix  cp <src> <src> <dst>
