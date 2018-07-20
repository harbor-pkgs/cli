package cli

type node struct {
	Pos        int
	RawFlag    string
	Value      *string
	Rule       *rule
	Count      int
	IsCmd      bool
	CmdHandled bool
}

type nodes []*node
type syntax struct {
	nodes map[int]*node
}

// Returns the all nodes that have the specified rule
func (s syntax) FindRules(rule *rule) nodes {
	var result nodes
	for _, node := range s.nodes {
		if node.Rule == rule {
			result = append(result, node)
		}
	}
	return result
}

func (s syntax) FindWithFlag(flag ruleFlag) nodes {
	var result nodes
	for _, node := range s.nodes {
		if node.Rule.HasFlag(flag) {
			result = append(result, node)
		}
	}
	return result
}

func (s *syntax) Add(node *node) {
	s.nodes[node.Pos] = node
}

func (s *syntax) Contains(pos int) bool {
	_, ok := s.nodes[pos]
	return ok
}
