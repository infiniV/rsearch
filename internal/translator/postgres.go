package translator

import (
	"fmt"
	"strings"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
)

// PostgresTranslator translates AST nodes to PostgreSQL queries.
type PostgresTranslator struct {
	paramCount int
	params     []interface{}
	paramTypes []string
}

// NewPostgresTranslator creates a new PostgreSQL translator.
func NewPostgresTranslator() *PostgresTranslator {
	return &PostgresTranslator{}
}

// DatabaseType returns the database type.
func (p *PostgresTranslator) DatabaseType() string {
	return "postgres"
}

// Translate converts an AST node to a PostgreSQL query.
func (p *PostgresTranslator) Translate(ast parser.Node, schema *schema.Schema) (*TranslatorOutput, error) {
	// Reset state for new translation
	p.paramCount = 0
	p.params = make([]interface{}, 0)
	p.paramTypes = make([]string, 0)

	whereClause, err := p.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	return NewSQLOutput(whereClause, p.params, p.paramTypes), nil
}

// translateNode recursively translates AST nodes.
func (p *PostgresTranslator) translateNode(node parser.Node, schema *schema.Schema) (string, error) {
	switch n := node.(type) {
	case *parser.FieldQuery:
		return p.translateFieldQuery(n, schema)
	case *parser.BinaryOp:
		return p.translateBinaryOp(n, schema)
	case *parser.RangeQuery:
		return p.translateRangeQuery(n, schema)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

// translateFieldQuery translates a simple field:value query.
func (p *PostgresTranslator) translateFieldQuery(fq *parser.FieldQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(fq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", fq.Field, schema.Name)
	}

	// Extract value from ValueNode
	value := fq.Value.Value()

	// Add parameter
	p.paramCount++
	p.params = append(p.params, value)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	// Generate SQL with parameterized query (use resolved column name)
	return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil
}

// translateBinaryOp translates AND/OR operations.
func (p *PostgresTranslator) translateBinaryOp(bo *parser.BinaryOp, schema *schema.Schema) (string, error) {
	left, err := p.translateNode(bo.Left, schema)
	if err != nil {
		return "", err
	}

	right, err := p.translateNode(bo.Right, schema)
	if err != nil {
		return "", err
	}

	// Determine if we need parentheses
	leftNeedsParens := p.needsParentheses(bo.Left)
	rightNeedsParens := p.needsParentheses(bo.Right)

	if leftNeedsParens {
		left = fmt.Sprintf("(%s)", left)
	}
	if rightNeedsParens {
		right = fmt.Sprintf("(%s)", right)
	}

	operator := strings.ToUpper(bo.Op)
	return fmt.Sprintf("%s %s %s", left, operator, right), nil
}

// needsParentheses determines if a node needs parentheses.
func (p *PostgresTranslator) needsParentheses(node parser.Node) bool {
	// Binary operations need parentheses when nested
	_, isBinaryOp := node.(*parser.BinaryOp)
	return isBinaryOp
}

// translateRangeQuery translates range queries like field:[start TO end].
func (p *PostgresTranslator) translateRangeQuery(rq *parser.RangeQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(rq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", rq.Field, schema.Name)
	}

	// Extract values from ValueNodes
	startValue := rq.Start.Value()
	endValue := rq.End.Value()

	// Handle wildcard bounds (unbounded ranges)
	startIsWildcard := startValue == "*"
	endIsWildcard := endValue == "*"

	// Handle single-sided unbounded ranges
	if startIsWildcard && !endIsWildcard {
		// Only upper bound: field < or <= value
		p.paramCount++
		p.params = append(p.params, endValue)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		op := "<"
		if rq.InclusiveEnd {
			op = "<="
		}
		return fmt.Sprintf("%s %s $%d", columnName, op, p.paramCount), nil
	}

	if !startIsWildcard && endIsWildcard {
		// Only lower bound: field > or >= value
		p.paramCount++
		p.params = append(p.params, startValue)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		op := ">"
		if rq.InclusiveStart {
			op = ">="
		}
		return fmt.Sprintf("%s %s $%d", columnName, op, p.paramCount), nil
	}

	// Both bounds specified
	if rq.InclusiveStart && rq.InclusiveEnd {
		// Both inclusive: BETWEEN
		p.paramCount++
		p.params = append(p.params, startValue)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		p.paramCount++
		p.params = append(p.params, endValue)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		return fmt.Sprintf("%s BETWEEN $%d AND $%d", columnName, p.paramCount-1, p.paramCount), nil
	}

	// Mixed or exclusive ranges: use comparison operators
	var clauses []string

	// Start condition
	p.paramCount++
	p.params = append(p.params, startValue)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	if rq.InclusiveStart {
		clauses = append(clauses, fmt.Sprintf("%s >= $%d", columnName, p.paramCount))
	} else {
		clauses = append(clauses, fmt.Sprintf("%s > $%d", columnName, p.paramCount))
	}

	// End condition
	p.paramCount++
	p.params = append(p.params, endValue)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	if rq.InclusiveEnd {
		clauses = append(clauses, fmt.Sprintf("%s <= $%d", columnName, p.paramCount))
	} else {
		clauses = append(clauses, fmt.Sprintf("%s < $%d", columnName, p.paramCount))
	}

	return strings.Join(clauses, " AND "), nil
}
