package translator

// Node is the base interface for all AST nodes.
// This is a stub implementation until the parser is integrated.
type Node interface {
	Type() string
}

// FieldQuery represents a simple field:value query.
type FieldQuery struct {
	Field string
	Value string
}

func (f *FieldQuery) Type() string {
	return "field_query"
}

// BinaryOp represents a binary operation (AND, OR).
type BinaryOp struct {
	Op    string // "AND", "OR"
	Left  Node
	Right Node
}

func (b *BinaryOp) Type() string {
	return "binary_op"
}

// RangeQuery represents a range query like field:[start TO end].
type RangeQuery struct {
	Field          string
	Start          interface{}
	End            interface{}
	InclusiveStart bool
	InclusiveEnd   bool
}

func (r *RangeQuery) Type() string {
	return "range_query"
}

// UnaryOp represents a unary operation (NOT, +, -).
type UnaryOp struct {
	Op      string // "NOT", "+", "-"
	Operand Node
}

func (u *UnaryOp) Type() string {
	return "unary_op"
}
