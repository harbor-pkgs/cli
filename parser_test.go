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
	// With default parser and 'foo' flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo"})

	// Given -h
	retCode, err := p.Parse(nil, []string{"-h"})

	// Parser should return HelpError{}
	require.NotNil(t, err)
	assert.Equal(t, "user asked for help; inspect this error with cli.isHelpError()", err.Error())

	// Should be a help error
	_, ok := err.(*cli.HelpError)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, true, cli.IsHelpError(err))
	assert.Equal(t, true, ok)
}

func TestParserNoArgs(t *testing.T) {
	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo"})

	// Given no arguments
	retCode, err := p.Parse(nil, []string{})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
}

func TestFooFlag(t *testing.T) {
	var foo string
	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

	// Given double prefix
	retCode, err := p.Parse(nil, []string{"--foo", "bar"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)

	// Given single prefix
	retCode, err = p.Parse(nil, []string{"-foo", "bar"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)
}

func TestFlagExpectedValue(t *testing.T) {
	var foo string
	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo", Store: &foo})

	// Given no value
	retCode, err := p.Parse(nil, []string{"--foo"})

	// Parser should return 0 and no error
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "expected flag '--foo' to have an argument", err.Error())
	assert.Equal(t, "", foo)
}

func TestFlagCount(t *testing.T) {
	var count int

	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "verbose", Count: &count, Aliases: []string{"v"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--verbose", "-v"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
}

func TestFlagCountWithValue(t *testing.T) {
	var count int
	var foo []string

	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar", "-f", "bang"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bang", "bar"}, foo)
}

func TestFlagIsRequired(t *testing.T) {
	var foo, bar string

	// With default parser and foo flag
	p := cli.NewParser()
	p.Add(&cli.Flag{Name: "foo", Required: true, Store: &foo, Aliases: []string{"f"}})
	p.Add(&cli.Flag{Name: "bar", Store: &bar, Aliases: []string{"b"}})

	// Given
	retCode, err := p.Parse(nil, []string{"-b", "bar"})

	// Parser should return error
	assert.NotNil(t, err)
	assert.Equal(t, cli.ErrorRetCode, retCode)
	assert.Equal(t, "flag '--foo' is required", err.Error())
}

func TestFlagWithSlice(t *testing.T) {
	var foo []string
	var count int

	// With default parser and foo flag
	p := cli.NewParser()
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar,bang", "-f", "foo"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, 2, count)
	sort.Strings(foo)
	assert.Equal(t, []string{"bar,bang", "foo"}, foo)
}

func TestFlagWithMap(t *testing.T) {
	var foo map[string]string
	var count int

	// With default parser and foo flag
	p := cli.NewParser()
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", "bar=foo", "-f", "foo=bar"})

	// Parser should return 0 and no error
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

	// With default parser and foo flag
	p := cli.NewParser()
	// Count implies 'CanRepeat=true'
	p.Add(&cli.Flag{Name: "foo", Store: &foo, Count: &count, Aliases: []string{"f"}})

	// Given
	retCode, err := p.Parse(nil, []string{"--foo", `{"bar":"foo"}`, "-f", `{"foo": "bar", "bash": "bang"}`})

	// Parser should return 0 and no error
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

// TODO: Test matching flags with no prefix if enabled
// TODO: Test default values with all supported scalar and slice values
// TODO: Test StringToSlice
// TODO: Test From alternate sources
