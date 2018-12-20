package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/require"
)

func TestHelpMessage(t *testing.T) {
	var bar, foo, argOne, argTwo, envOne string
	var hasFlag, hasEnv bool
	var envTwo []string

	p := cli.New(&cli.Config{
		Name:   "test",
		Desc:   "This is the description of the application",
		Epilog: "Copyright 2018 By Derrick J. Wippler",
	})

	p.Add(
		&cli.Argument{Name: "arg-one", Required: true, Store: &argOne, Help: "this is a required argument"},
		&cli.Argument{Name: "arg-two", Store: &argTwo, Help: "this argument is optional"},
	)

	p.Add(
		&cli.Flag{Name: "flag", IsSet: &hasFlag, Help: "this is my flag"},
		&cli.Flag{Name: "foo", Store: &foo, Aliases: []string{"f"}, Help: "used to store bars"},
		&cli.Flag{Name: "bar", Store: &bar, Help: "used to store foos"},
	)

	p.Add(
		&cli.EnvVar{Name: "ENV_ONE", IsSet: &hasEnv, Store: &envOne, Help: "this is env one, it holds 1 thing"},
		&cli.EnvVar{Name: "ENV_TWO", Store: &envTwo, Help: "this is env one, it holds a comma separate list of 2 things"},
	)

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
	compare("Environment Variables:")
	compare("")
	compare("Copyright 2018 By Derrick J. Wippler")

	// TODO: Test Usage when EnvVar are used
}

func TestGenerateEnvConfig(t *testing.T) {
	var bar, foo string
	var endpoints []string
	var count int

	p := cli.New(nil)
	p.Add(&cli.EnvVar{Name: "BAR", Store: &bar, Help: "A bar to put beer into, with extra hops"})
	p.Add(&cli.EnvVar{Name: "foo", Env: "FOO", Store: &foo, Help: "Lorem ipsum dolor sit amet, " +
		"consect etura dipiscing elit, sed do eiusmod tempor incididunt ut labore etmollit anim id " +
		"est laborum."})
	p.Add(&cli.EnvVar{Name: "count", Env: "COUNT", Store: &count, Default: "1", Help: "The number of things to come"})
	p.Add(&cli.EnvVar{Name: "endpoints", Env: "ENDPOINTS", Store: &endpoints,
		Help: "A comma separated list of endpoints our application can connect too"})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.Nil(t, err)
	require.Equal(t, 0, retCode)

	config := p.GenerateEnvConfig()
	fmt.Println(config)

	lines := strings.Split(config, "\n")
	var i int

	compare := func(expected string) {
		require.Equal(t, expected, lines[i])
		i++
	}

	compare(`# A bar to put beer into, with extra hops`)
	compare(`BAR=`)
	compare(``)
	compare(`# The number of things to come`)
	compare(`# Default: "1"`)
	compare(`COUNT=`)
	compare(``)
	compare(`# A comma separated list of endpoints our application can connect too`)
	compare(`ENDPOINTS=`)
	compare(``)
	compare(`# Lorem ipsum dolor sit amet, consect etura dipiscing elit, sed do eiusmod tempor incididunt ut`)
	compare(`# labore etmollit anim id est laborum.`)
	compare(`FOO=`)
}
