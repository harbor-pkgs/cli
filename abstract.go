package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type absNode struct {
	Pos    int
	Offset int
	Value  *string
	Rule   *rule
	Flags  Flags
	/*IsCmd      bool
	CmdHandled bool*/
	ValueFor *absNode
}

// Returns true if this node associated with a rule
func (n *absNode) HasRule() bool {
	return n.Rule != nil || n.ValueFor != nil
}

type nodeList []*absNode

// TODO: Rename this 'abstract'
type abstract struct {
	nodes []*absNode
	rules ruleList
}

func newAbstract(parser *Parser) *abstract {
	return &abstract{
		rules: parser.rules,
	}
}

// Returns the all nodes that have the specified rule
func (a *abstract) FindRules(rule *rule) nodeList {
	var result nodeList
	for _, node := range a.nodes {
		if node.Rule == rule {
			result = append(result, node)
		}
	}
	return result
}

func (a *abstract) FindWithFlag(flag Flags) nodeList {
	var result nodeList
	for _, node := range a.nodes {
		if node.Rule != nil && node.Rule.HasFlag(flag) {
			result = append(result, node)
		}
	}
	return result
}

func (a *abstract) Add(n *absNode) {
	fmt.Printf("Add %+v\n", n)
	a.nodes = append(a.nodes, n)
}

// Return nodes that have no rule attached
func (a *abstract) UnknownArgs() nodeList {
	var results nodeList
	for k, v := range a.nodes {
		if v != nil {
			if !v.HasRule() {
				results = append(results, v)
			}
		}
		results = append(results, &absNode{Pos: k})
	}
	return results
}

// Returns true if the argument at the specified position is an option
func (a *abstract) AtPos(pos int) *absNode {
	for _, node := range a.nodes {
		if node.Pos == pos {
			return node
		}
	}
	return nil
}

// Returns true if at least one rule matched the given argument position
// TODO: Remove?
func (a *abstract) Contains(pos int) bool {
	for _, node := range a.nodes {
		if node.Pos == pos {
			return node.HasRule()
		}
	}
	return false
}

func (a *abstract) String() string {
	var lines []string
	for i := 0; i < len(a.nodes); i++ {
		var name string
		if a.nodes[i].Rule != nil {
			name = a.nodes[i].Rule.Name
		}
		lines = append(lines, fmt.Sprintf("[%d] '%s' - %+v", i, name, a.nodes[i]))
	}
	return strings.Join(lines, "\n")
}

func (a *abstract) Get(ctx context.Context, key string, flags Flags) (interface{}, int, error) {
	//fmt.Printf("Get(%s,%s)\n", key, kind)
	rule := a.rules.GetRule(key)
	if rule == nil {
		return "", 0, nil
	}

	// TODO: If user requests positional only arguments, eliminate args/options we find that are not in our range

	var values []string
	var count int
	// collect all the values for this rule
	for _, node := range a.FindRules(rule) {
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

func (a *abstract) Source() string {
	return cliSource
}
