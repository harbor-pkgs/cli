package cli

import (
	"fmt"
	"strings"
)

type node struct {
	Pos        int
	RawFlag    string
	Value      *string
	Rule       *rule
	Count      int
	IsCmd      bool
	CmdHandled bool
}

func (n *node) String() string {
	if n.Rule != nil {
		return fmt.Sprintf("node(rule: %s)", n.Rule.Name)
	}
	return fmt.Sprintf("node(rule: <none>)")
}

type nodes []*node
type linearSyntax struct {
	nodes map[int]*node
}

func newLinearSyntax() *linearSyntax {
	return &linearSyntax{
		nodes: make(map[int]*node),
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
		if node.Rule.HasFlag(flag) {
			result = append(result, node)
		}
	}
	return result
}

func (s *linearSyntax) Add(node *node) {
	s.nodes[node.Pos] = node
}

func (s *linearSyntax) Contains(pos int) bool {
	_, ok := s.nodes[pos]
	return ok
}

func (s *linearSyntax) String() string {
	var lines []string
	for i, node := range s.nodes {
		lines = append(lines, fmt.Sprintf("[%d] %s", i, node))
	}
	return strings.Join(lines, "\n")
}
