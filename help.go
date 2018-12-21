package cli

import (
	"bytes"
	"fmt"
)

// Returns a string that contains documentation for each flag, argument and envvar provided
// which is suitable for display to the user as your applications help message.
//
//   Usage: test [flags]  <arg-one> [arg-two]
//
//   This is the description of the application
//
//   Arguments:
//     arg-one   this is a required argument
//     arg-two   this argument is optional
//
//   Flags:
//     --help, -h           display this help message and exit
//     --flag               this is my flag
//     --foo, -f <string>   used to store bars
//     --bar <int>          used to store number of foo's
func (p *Parser) GenerateHelp() string {
	var result bytes.Buffer
	if p.cfg.Usage != "" {
		result.WriteString(fmt.Sprintf("Usage: %s\n", p.cfg.Usage))
	} else {
		result.WriteString(fmt.Sprintf("Usage: %s %s %s\n", p.cfg.Name,
			p.generateUsage(isFlag),
			p.generateUsage(isArgument)))
	}

	if p.cfg.Desc != "" {
		result.WriteString("\n")
		result.WriteString(WordWrap(p.cfg.Desc, 0, p.cfg.WordWrap))
		result.WriteString("\n")
	}

	commands := p.generateHelpSection(isCommand)
	if commands != "" {
		result.WriteString("\nCommands:\n")
		result.WriteString(commands)
	}

	argument := p.generateHelpSection(isArgument)
	if argument != "" {
		result.WriteString("\nArguments:\n")
		result.WriteString(argument)
	}

	options := p.generateHelpSection(isFlag)
	if options != "" {
		result.WriteString("\nFlags:\n")
		result.WriteString(options)
	}

	envVars := p.generateHelpSection(isEnvVar)
	if options != "" {
		result.WriteString("\nEnvironment Variables:\n")
		result.WriteString(envVars)
	}

	if p.cfg.Epilog != "" {
		result.WriteString("\n" + WordWrap(p.cfg.Epilog, 0, p.cfg.WordWrap))
	}
	return result.String()
}

// Returns a byte array that contains documentation and type signature for each flag, argument
// and envvar with an source environment variable defined. This is suitable for generating an
// environment example file for users of your app to source into their shell environment.
//
//   # The count of things to come (Default:"1")
//   # export COUNT=<int>
//
//   # A comma separated list of endpoints our application can connect too
//   # export ENDPOINTS=<str>,<str>
func (p *Parser) GenerateEnvConfig() []byte {
	var result bytes.Buffer
	for _, rule := range p.rules {
		if rule.EnvVar == "" {
			continue
		}
		result.Write(rule.GenerateEnvUsage(p.cfg.WordWrap))
	}
	return result.Bytes()
}

// Returns a byte array that contains documentation and type signature for each flag, and envvar
// added to the parser. This is suitable for generating an INI example config file
// for users of your app to modify.
//
//   # The count of things to come (Default:"1")
//   # count=<int>
//
//   # A comma separated list of endpoints our application can connect too
//   # endpoints=<str>,<str>
func (p *Parser) GenerateINIConfig() []byte {
	var result bytes.Buffer
	for _, rule := range p.rules {
		// Exclude sub commands and the help rule
		if rule.HasFlag(isCommand|isHelpRule){
			continue
		}
		result.Write(rule.GenerateINIUsage(p.cfg.WordWrap))
	}
	return result.Bytes()
}

func (p *Parser) generateUsage(flags ruleFlag) string {
	var result bytes.Buffer

	if flags == isFlag {
		return "[flags]"
	}

	for _, rule := range p.rules {
		if !rule.HasFlag(flags) {
			continue
		}
		result.WriteString(" " + rule.GenerateUsage())
	}
	return result.String()
}

func (p *Parser) generateHelpSection(flags ruleFlag) string {
	type helpMsg struct {
		Flags   string
		Message string
	}
	var result bytes.Buffer
	var options []helpMsg

	// Ask each rule to generate a Help message for the options
	maxLen := 0
	for _, rule := range p.rules {
		if !rule.HasFlag(flags) {
			continue
		}
		flags, message := rule.GenerateHelp()
		if len(flags) > maxLen {
			maxLen = len(flags)
		}
		options = append(options, helpMsg{flags, message})
	}

	// Set our indent length
	indent := maxLen + 3
	flagFmt := fmt.Sprintf("%%-%ds%%s\n", indent)

	for _, opt := range options {
		message := WordWrap(opt.Message, indent, p.cfg.WordWrap)
		result.WriteString(fmt.Sprintf(flagFmt, opt.Flags, message))
	}
	return result.String()
}
