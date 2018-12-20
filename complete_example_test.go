package cli_test

import (
	"fmt"

	"github.com/harbor-pkgs/cli"
)

type Config struct {
	PowerLevel int
	Message    string
	Weapons    []string
	Verbose    int
	Name       string
	Villains   []string
	Debug      bool
}

func Example_complete() {
	var conf Config

	parser := cli.New(&cli.Config{
		Desc: "This is a demo app to showcase some features of cli",
		Name: "demo",
	})

	// Store Integers directly into a struct field with a default value
	parser.Add(&cli.Flag{
		Help:    "Set our power level to over 5000!",
		Store:   &conf.PowerLevel,
		Aliases: []string{"p"},
		Name:    "power-level",
		Env:     "POWER_LEVEL",
		Default: "10000",
	})

	// Add an environment variable as a possible source for the argument
	parser.Add(&cli.Flag{
		Help:    "Send a message to your enemies, make them quake in fear",
		Default: "over-ten-thousand",
		Store:   &conf.Message,
		Aliases: []string{"m"},
		Name:    "message",
		Env:     "MESSAGE",
	})

	// Pass a comma separated list of strings and get a []string slice
	parser.Add(&cli.Flag{
		Help:    "List of weapons to choose from separated by a comma",
		Default: "vi,emacs,pico",
		Store:   conf.Weapons,
		Aliases: []string{"s"},
		Name:    "weapons",
		Env:     "WEAPONS",
	})

	// Count the number of times an option is seen
	parser.Add(&cli.Flag{
		Help:    "Declare to the world after each evil is vanquished",
		Aliases: []string{"v"},
		Count:   &conf.Verbose,
		Name:    "verbose",
	})

	// Set bool to true if the option is present on the command line
	parser.Add(&cli.Flag{
		Help:    "Turn on universe debug, it must give up it's secrets",
		Aliases: []string{"d"},
		Name:    "debug",
		IsSet:   &conf.Debug,
	})

	// --help option is provided by default, however you can override
	parser.Add(&cli.Flag{
		Help:    "Show this help message and exit",
		Aliases: []string{"H"},
		Name:    "help",
	})

	// Add Required arguments
	parser.Add(&cli.Argument{
		Help:     "The name of the hero who fights for justice",
		Store:    &conf.Name,
		Name:     "name",
		Required: true,
	})

	// Add optional arguments that can repeat
	parser.Add(&cli.Argument{
		Help:      "List of villains to vanquish",
		Name:      "the-villains",
		Store:     conf.Villains,
		CanRepeat: true,
	})

	// ParseOrExit() is just a convenience, you can call
	// parser.Parse() directly and handle the errors
	// yourself if you have more complicated use case
	parser.ParseOrExit()

	// Demo default variables in a struct
	fmt.Printf("Power        '%d'\n", conf.PowerLevel)
	fmt.Printf("Message      '%s'\n", conf.Message)
	fmt.Printf("Weapons      '%v'\n", conf.Weapons)
	fmt.Printf("Verbose      '%d'\n", conf.Verbose)
	fmt.Printf("Name         '%s'\n", conf.Name)
	fmt.Printf("Villains     '%v'\n", conf.Villains)
	fmt.Printf("Debug        '%v'\n", conf.Debug)
	fmt.Println("")

	// TODO: Demo greedy options [SRC... DST]
}
