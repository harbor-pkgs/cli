package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/harbor-pkgs/cli"
	"github.com/stretchr/testify/require"
)

func TestHelpMessage(t *testing.T) {
	var foo, argOne, argTwo, envOne string
	var hasFlag, hasEnv, isTrue bool
	var envTwo []string
	var bar int

	p := cli.New(&cli.Config{
		Name:   "test",
		Desc:   "This is the description of the application",
		Epilog: "Copyright 2018 By Derrick J. Wippler",
	})

	p.Add(
		&cli.Argument{Name: "arg-one", Flags: cli.Required, Store: &argOne, Help: "this is a required argument"},
		&cli.Argument{Name: "arg-two", Store: &argTwo, Help: "this argument is optional"},
	)

	p.Add(
		&cli.Option{Name: "flag", IsSet: &hasFlag, Help: "this is my flag"},
		&cli.Option{Name: "foo", Store: &foo, Aliases: []string{"f"}, Help: "used to store bars"},
		&cli.Option{Name: "bar", Store: &bar, Help: "used to store number of foo's"},
		&cli.Option{Name: "true", Store: &isTrue, Help: "is very true"},
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

	compare("Usage: test [options]  <arg-one> [arg-two]")
	compare("")
	compare("This is the description of the application")
	compare("")
	compare("Arguments:")
	compare("  arg-one   this is a required argument")
	compare("  arg-two   this argument is optional")
	compare("")
	compare("Options:")
	compare("  --flag               this is my flag")
	compare("  --foo, -f <string>   used to store bars")
	compare("  --bar <int>          used to store number of foo's")
	compare("  --true <bool>        is very true")
	compare("  --help, -h           display this help message and exit")
	compare("")
	compare("Environment Variables:")
	compare("  ENV_ONE <string>      this is env one, it holds 1 thing")
	compare("  ENV_TWO <str>,<str>   this is env one, it holds a comma separate list of 2 things")
	compare("")
	compare("Copyright 2018 By Derrick J. Wippler")
}

func TestGenerateConfig(t *testing.T) {
	var bar, foo, thing string
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

	p.Add(&cli.Option{Name: "thing", Env: "THE_THING", Store: &thing, Default: "Let's call it a thingamajig",
		Help: "This is a rather simple thing"})

	// Given
	retCode, err := p.Parse(nil, []string{})

	// Then
	require.Nil(t, err)
	require.Equal(t, 0, retCode)

	config := string(p.GenerateEnvConfig())
	fmt.Println(config)

	lines := strings.Split(config, "\n")
	var i int

	compare := func(expected string) {
		require.Equal(t, expected, lines[i])
		i++
	}

	compare("# A bar to put beer into, with extra hops")
	compare("# export BAR=<string>")
	compare("")
	compare("# Lorem ipsum dolor sit amet, consect etura dipiscing elit, sed do eiusmod tempor incididunt ut")
	compare("# labore etmollit anim id est laborum.")
	compare("# export FOO=<string>")
	compare("")
	compare(`# The number of things to come (Default:"1")`)
	compare("# export COUNT=<int>")
	compare("")
	compare("# A comma separated list of endpoints our application can connect too")
	compare("# export ENDPOINTS=<str>,<str>")
	compare("")
	compare(`# This is a rather simple thing (Default:"Let's call it a thingamajig")`)
	compare("# export THE_THING=<string>")

	// Test INI Config
	config = string(p.GenerateINIConfig())
	fmt.Println(config)

	lines = strings.Split(config, "\n")
	i = 0

	compare = func(expected string) {
		require.Equal(t, expected, lines[i])
		i++
	}

	compare("# A bar to put beer into, with extra hops")
	compare("# BAR=<string>")
	compare("")
	compare("# Lorem ipsum dolor sit amet, consect etura dipiscing elit, sed do eiusmod tempor incididunt ut")
	compare("# labore etmollit anim id est laborum.")
	compare("# foo=<string>")
	compare("")
	compare(`# The number of things to come (Default:"1")`)
	compare("# count=<int>")
	compare("")
	compare("# A comma separated list of endpoints our application can connect too")
	compare("# endpoints=<str>,<str>")
	compare("")
	compare(`# This is a rather simple thing (Default:"Let's call it a thingamajig")`)
	compare("# thing=<string>")
}
