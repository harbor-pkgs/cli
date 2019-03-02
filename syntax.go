package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type node struct {
	Pos    int
	Offset int
	Value  *string
	Rule   *rule
	Flags  Flags
	/*IsCmd      bool
	CmdHandled bool*/
	ValueFor *node
}

// Returns true if this node associated with a rule
func (n *node) HasRule() bool {
	return n.Rule != nil || n.ValueFor != nil
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

func (s *linearSyntax) FindWithFlag(flag Flags) nodes {
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

// Returns true if the argument at the specified position is an option
func (s *linearSyntax) AtPos(pos int) *node {
	for _, node := range s.nodes {
		if node.Pos == pos {
			return node
		}
	}
	return nil
}

// Returns true if at least one rule matched the given argument position
func (s *linearSyntax) Contains(pos int) bool {
	for _, node := range s.nodes {
		if node.Pos == pos {
			return node.HasRule()
		}
	}
	return false
}

func (s *linearSyntax) String() string {
	var lines []string
	for i := 0; i < len(s.nodes); i++ {
		var name string
		if s.nodes[i].Rule != nil {
			name = s.nodes[i].Rule.Name
		}
		lines = append(lines, fmt.Sprintf("[%d] '%s' - %+v", i, name, s.nodes[i]))
	}
	return strings.Join(lines, "\n")
}

func (s *linearSyntax) Get(ctx context.Context, key string, flags Flags) (interface{}, int, error) {
	//fmt.Printf("Get(%s,%s)\n", key, kind)
	rule := s.rules.GetRule(key)
	if rule == nil {
		return "", 0, nil
	}

	// TODO: If user requests positional only arguments, eliminate args/options we find that are not in our range

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

	return convToKind(values, flags, count)
}

// Given a list of string values, attempt to convert them to a kind
func convToKind(values []string, flags Flags, count int) (interface{}, int, error) {
	switch {
	case flags.Has(ScalarKind):
		//fmt.Printf("Get Ret: %s, %d\n", values[0], count)
		return values[0], count, nil
	case flags.Has(SliceKind):
		fmt.Printf("flag: %t\n", flags.Has(NoSplit))
		// If only one item is provided, it must be a comma separated list
		if count == 1 && !flags.Has(NoSplit) {
			return ToSlice(values[0]), count, nil
		}
		return values, count, nil
	case flags.Has(MapKind):
		// each string in the list should be a key value pair
		// either in the form `key=value` or `{"key": "value"}`
		r := make(map[string]string)
		for _, value := range values {
			kv, err := ToStringMap(value)
			if err != nil {
				return nil, 0, fmt.Errorf("map conversion: %s", err)
			}
			// Merge the key values for each of the items
			for k, v := range kv {
				r[k] = v
			}
		}
		//fmt.Printf("Get Ret: %s, %d\n", r, count)
		return r, count, nil
	}
	return nil, 0, errors.New("invalid rule; missing (ScalarKind|SliceKind|MapKind) in flags")
}

func (p *linearSyntax) Source() string {
	return cliSource
}
