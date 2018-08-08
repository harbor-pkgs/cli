package cli

import (
	"bytes"
	"fmt"
)

func (p *Parser) GenerateHelp() string {
	var result bytes.Buffer
	if p.Usage != "" {
		result.WriteString(fmt.Sprintf("Usage: %s\n", p.Usage))
	} else {
		result.WriteString(fmt.Sprintf("Usage: %s %s %s\n", p.Name,
			p.generateUsage(isFlag),
			p.generateUsage(isArgument)))
	}

	if p.Desc != "" {
		result.WriteString("\n")
		result.WriteString(WordWrap(p.Desc, 0, p.WordWrap))
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

	if p.Epilog != "" {
		result.WriteString("\n" + WordWrap(p.Epilog, 0, p.WordWrap))
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
		message := WordWrap(opt.Message, indent, p.WordWrap)
		result.WriteString(fmt.Sprintf(flagFmt, opt.Flags, message))
	}
	return result.String()
}
