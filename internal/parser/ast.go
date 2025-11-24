package parser

// Node is the interface for all AST nodes
type Node interface {
	Type() string
	Position() Position
}

// ValueNode represents a value in the query (term, phrase, wildcard, regex)
type ValueNode interface {
	Value() interface{}
	IsValueNode()
}

// BinaryOp represents a binary operation (AND, OR)
type BinaryOp struct {
	Op    string // "AND", "OR"
	Left  Node
	Right Node
	Pos   Position
}

func (n *BinaryOp) Type() string     { return "BinaryOp" }
func (n *BinaryOp) Position() Position { return n.Pos }

// UnaryOp represents a unary operation (NOT, !)
type UnaryOp struct {
	Op      string // "NOT", "!"
	Operand Node
	Pos     Position
}

func (n *UnaryOp) Type() string     { return "UnaryOp" }
func (n *UnaryOp) Position() Position { return n.Pos }

// RequiredQuery represents a required term (+term)
type RequiredQuery struct {
	Query Node
	Pos   Position
}

func (n *RequiredQuery) Type() string     { return "RequiredQuery" }
func (n *RequiredQuery) Position() Position { return n.Pos }

// ProhibitedQuery represents a prohibited term (-term)
type ProhibitedQuery struct {
	Query Node
	Pos   Position
}

func (n *ProhibitedQuery) Type() string     { return "ProhibitedQuery" }
func (n *ProhibitedQuery) Position() Position { return n.Pos }

// FieldQuery represents a field:value query
type FieldQuery struct {
	Field string
	Value ValueNode
	Pos   Position
}

func (n *FieldQuery) Type() string     { return "FieldQuery" }
func (n *FieldQuery) Position() Position { return n.Pos }

// FieldGroupQuery represents field:(a OR b)
type FieldGroupQuery struct {
	Field   string
	Queries []Node
	Pos     Position
}

func (n *FieldGroupQuery) Type() string     { return "FieldGroupQuery" }
func (n *FieldGroupQuery) Position() Position { return n.Pos }

// RangeQuery represents a range query [start TO end] or {start TO end}
type RangeQuery struct {
	Field          string
	Start          ValueNode
	End            ValueNode
	InclusiveStart bool
	InclusiveEnd   bool
	Pos            Position
}

func (n *RangeQuery) Type() string     { return "RangeQuery" }
func (n *RangeQuery) Position() Position { return n.Pos }

// FuzzyQuery represents a fuzzy search (term~distance)
type FuzzyQuery struct {
	Field    string
	Term     string
	Distance int // 0 means default (2)
	Pos      Position
}

func (n *FuzzyQuery) Type() string     { return "FuzzyQuery" }
func (n *FuzzyQuery) Position() Position { return n.Pos }

// ProximityQuery represents a proximity search ("phrase"~distance)
type ProximityQuery struct {
	Field    string
	Phrase   string
	Distance int
	Pos      Position
}

func (n *ProximityQuery) Type() string     { return "ProximityQuery" }
func (n *ProximityQuery) Position() Position { return n.Pos }

// ExistsQuery represents an existence check (_exists_:field)
type ExistsQuery struct {
	Field string
	Pos   Position
}

func (n *ExistsQuery) Type() string     { return "ExistsQuery" }
func (n *ExistsQuery) Position() Position { return n.Pos }

// BoostQuery represents a boosted query (query^boost)
type BoostQuery struct {
	Query Node
	Boost float64
	Pos   Position
}

func (n *BoostQuery) Type() string     { return "BoostQuery" }
func (n *BoostQuery) Position() Position { return n.Pos }

// TermValue represents a simple term
type TermValue struct {
	Term string
	Pos  Position
}

func (n *TermValue) Value() interface{} { return n.Term }
func (n *TermValue) IsValueNode()       {}

// PhraseValue represents a quoted phrase
type PhraseValue struct {
	Phrase string
	Pos    Position
}

func (n *PhraseValue) Value() interface{} { return n.Phrase }
func (n *PhraseValue) IsValueNode()       {}

// WildcardValue represents a wildcard pattern (* or ?)
type WildcardValue struct {
	Pattern string
	Pos     Position
}

func (n *WildcardValue) Value() interface{} { return n.Pattern }
func (n *WildcardValue) IsValueNode()       {}

// RegexValue represents a regex pattern /pattern/
type RegexValue struct {
	Pattern string
	Pos     Position
}

func (n *RegexValue) Value() interface{} { return n.Pattern }
func (n *RegexValue) IsValueNode()       {}

// NumberValue represents a numeric value
type NumberValue struct {
	Number string
	Pos    Position
}

func (n *NumberValue) Value() interface{} { return n.Number }
func (n *NumberValue) IsValueNode()       {}

// WildcardQuery represents a standalone wildcard query (not in a field context)
type WildcardQuery struct {
	Pattern string
	Pos     Position
}

func (n *WildcardQuery) Type() string     { return "WildcardQuery" }
func (n *WildcardQuery) Position() Position { return n.Pos }

// TermQuery represents a standalone term query
type TermQuery struct {
	Term string
	Pos  Position
}

func (n *TermQuery) Type() string     { return "TermQuery" }
func (n *TermQuery) Position() Position { return n.Pos }

// PhraseQuery represents a standalone phrase query
type PhraseQuery struct {
	Phrase string
	Pos    Position
}

func (n *PhraseQuery) Type() string     { return "PhraseQuery" }
func (n *PhraseQuery) Position() Position { return n.Pos }

// GroupQuery represents a grouped query (...)
type GroupQuery struct {
	Query Node
	Pos   Position
}

func (n *GroupQuery) Type() string     { return "GroupQuery" }
func (n *GroupQuery) Position() Position { return n.Pos }
