package cli

import (
	"bytes"
	"fmt"
	"strings"
)

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
		result.WriteString("\nEnvironment Vars:\n")
		result.WriteString(envVars)
	}

	if p.cfg.Epilog != "" {
		result.WriteString("\n" + WordWrap(p.cfg.Epilog, 0, p.cfg.WordWrap))
	}
	return result.String()
}

// TODO: Document this method
func (p *Parser) GenerateEnvConfig() string {
	var result bytes.Buffer

	for _, rule := range p.rules.SortRulesWithFlag(isEnvVar) {
		if rule.HelpMsg != "" {
			for _, line := range strings.Split(WordWrap(rule.HelpMsg, 0, p.cfg.WordWrap), "\n") {
				result.WriteString("# " + line + "\n")
			}
		}
		result.WriteString(rule.GenerateEnvUsage())
		result.WriteString(rule.EnvVar + "=\n\n")
	}

	return result.String()
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
