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

func (s *linearSyntax) Get(ctx context.Context, key string, valueType ValueType) (interface{}, int, error) {
	//fmt.Printf("Get(%s,%s)\n", key, valueType)
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

	switch valueType {
	case ScalarType:
		//fmt.Printf("Get Ret: %s, %d\n", values[0], count)
		return values[0], count, nil
	case ListType:
		//fmt.Printf("Get Ret: %s, %d\n", values, count)
		return values, len(values), nil
	case MapType:
		// each string in the list should be a key value pair
		// either in the form `key=value` or `{"key": "value"}`
		r := make(map[string]string)
		for _, value := range values {
			kv, err := StringToMap(value)
			if err != nil {
				return nil, 0, fmt.Errorf("during Parser.Get() map conversion: %s", err)
			}
			// Merge the key values for each of the items
			for k, v := range kv {
				r[k] = v
			}
		}
		//fmt.Printf("Get Ret: %s, %d\n", r, count)
		return r, count, nil
	}
	return nil, 0, fmt.Errorf("no such ValueType '%s'", valueType)
}

func (p *linearSyntax) Source() string {
	return cliSource
}
