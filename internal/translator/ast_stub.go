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

// FuzzyQuery represents a fuzzy search query like field:term~ or field:term~1.
type FuzzyQuery struct {
	Field    string
	Term     string
	Distance int // Edit distance (default 2 if not specified)
}

func (f *FuzzyQuery) Type() string {
	return "fuzzy_query"
}
