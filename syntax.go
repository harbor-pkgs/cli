package cli

import (
	"context"
	"fmt"
	"strings"
)

type node struct {
	Pos        int
	Offset     int
	Value      *string
	Rule       *rule
	IsCmd      bool
	CmdHandled bool
	ValueFor   *node
}

type nodes []*node
type linearSyntax struct {
	nodes []*node
	rules ruleList
}

func newLinearSyntax(parser *Parser) *linearSyntax {
	return &linearSyntax{
		rules: parser.rules,
	}
}

// Returns the all nodes that have the specified rule
func (s *linearSyntax) FindRules(rule *rule) nodes {
	var result nodes
	for _, node := range s.nodes {
		if node.Rule == rule {
			result = append(result, node)
		}
	}
	return result
}

func (s *linearSyntax) FindWithFlag(flag ruleFlag) nodes {
	var result nodes
	for _, node := range s.nodes {
		if node.Rule != nil && node.Rule.HasFlag(flag) {
			result = append(result, node)
		}
	}
	return result
}

func (s *linearSyntax) Add(n *node) {
	fmt.Printf("Add %+v\n", n)
	s.nodes = append(s.nodes, n)
}

// Returns true if at least one rule matched the given argument position
func (s *linearSyntax) Contains(pos int) bool {
	for _, node := range s.nodes {
		if node.Pos == pos {
			return true
		}
	}
	return false
}

func (s *linearSyntax) String() string {
	var lines []string
	for i := 0; i < len(s.nodes); i++ {
		lines = append(lines, fmt.Sprintf("[%d] %+v", i, s.nodes[i]))
	}
	return strings.Join(lines, "\n")
}

func (s *linearSyntax) Get(ctx context.Context, key string, kind Kind) (interface{}, int, error) {
	//fmt.Printf("Get(%s,%s)\n", key, kind)
	rule := s.rules.GetRule(key)
	if rule == nil {
		return "", 0, nil
	}

	// TODO: If user requests positional only arguments, eliminate args/flags we find that are not in our range

	var values []string
	var count int
	// collect all the values for this rule
	for _, node := range s.FindRules(rule) {
		count++
		if node.Value != nil {
			values = append(values, *node.Value)
		}
	}

	if len(values) == 0 {
		//fmt.Printf("Get Ret: <nil>, %d\n", count)
		return nil, count, nil
	}

	return sliceToKind(values, kind, count)
}

func (p *linearSyntax) Source() string {
	return cliSource
}
