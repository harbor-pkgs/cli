package cli_test

import (
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

	// Given no arguments
	retCode, err := p.Parse(nil, []string{"--foo", "bar"})

	// Parser should return 0 and no error
	assert.Nil(t, err)
	assert.Equal(t, 0, retCode)
	assert.Equal(t, "bar", foo)
}
