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
	metadata   map[string]interface{}
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
	p.metadata = make(map[string]interface{})

	whereClause, err := p.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	output := NewSQLOutput(whereClause, p.params, p.paramTypes)
	output.Metadata = p.metadata
	return output, nil
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
	case *parser.BoostQuery:
		return p.translateBoostQuery(n, schema)
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

	// Note: Searchable field removed as it's not in the current schema design
	_ = field // Use field if needed later

	// Add parameter
	p.paramCount++
	p.params = append(p.params, fq.Value)
	p.paramTypes = append(p.paramTypes, string(string(field.Type)))

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

	// Note: Searchable field removed as it's not in the current schema design
	_ = field // Use field if needed later

	// Handle inclusive vs exclusive ranges
	if rq.InclusiveStart && rq.InclusiveEnd {
		// Both inclusive: BETWEEN
		p.paramCount++
		p.params = append(p.params, rq.Start)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		p.paramCount++
		p.params = append(p.params, rq.End)
		p.paramTypes = append(p.paramTypes, string(field.Type))

		return fmt.Sprintf("%s BETWEEN $%d AND $%d", columnName, p.paramCount-1, p.paramCount), nil
	}

	// Mixed or exclusive ranges: use comparison operators
	var clauses []string

	// Start condition
	p.paramCount++
	p.params = append(p.params, rq.Start)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	if rq.InclusiveStart {
		clauses = append(clauses, fmt.Sprintf("%s >= $%d", columnName, p.paramCount))
	} else {
		clauses = append(clauses, fmt.Sprintf("%s > $%d", columnName, p.paramCount))
	}

	// End condition
	p.paramCount++
	p.params = append(p.params, rq.End)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	if rq.InclusiveEnd {
		clauses = append(clauses, fmt.Sprintf("%s <= $%d", columnName, p.paramCount))
	} else {
		clauses = append(clauses, fmt.Sprintf("%s < $%d", columnName, p.paramCount))
	}

	return strings.Join(clauses, " AND "), nil
}

// translateBoostQuery translates boosted queries.
// PostgreSQL doesn't natively support relevance scoring like search engines,
// so we translate the inner query normally and store the boost value in metadata
// for application-level use.
func (p *PostgresTranslator) translateBoostQuery(bq *parser.BoostQuery, schema *schema.Schema) (string, error) {
	// Translate the inner query
	innerSQL, err := p.translateNode(bq.Query, schema)
	if err != nil {
		return "", err
	}

	// Store boost information in metadata
	if p.metadata["boosts"] == nil {
		p.metadata["boosts"] = make([]map[string]interface{}, 0)
	}

	boosts := p.metadata["boosts"].([]map[string]interface{})
	boosts = append(boosts, map[string]interface{}{
		"query": bq.Query.Type(),
		"boost": bq.Boost,
	})
	p.metadata["boosts"] = boosts

	// Return the inner SQL without modification
	// SQL databases don't have native relevance scoring
	return innerSQL, nil
}
