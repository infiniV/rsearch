package translator

import (
	"fmt"
	"strings"

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
func (p *PostgresTranslator) Translate(ast Node, schema *schema.Schema) (*TranslatorOutput, error) {
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
func (p *PostgresTranslator) translateNode(node Node, schema *schema.Schema) (string, error) {
	switch n := node.(type) {
	case *FieldQuery:
		return p.translateFieldQuery(n, schema)
	case *BinaryOp:
		return p.translateBinaryOp(n, schema)
	case *RangeQuery:
		return p.translateRangeQuery(n, schema)
	case *WildcardQuery:
		return p.translateWildcardQuery(n, schema)
	case *RegexQuery:
		return p.translateRegexQuery(n, schema)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

// translateFieldQuery translates a simple field:value query.
func (p *PostgresTranslator) translateFieldQuery(fq *FieldQuery, schema *schema.Schema) (string, error) {
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
func (p *PostgresTranslator) translateBinaryOp(bo *BinaryOp, schema *schema.Schema) (string, error) {
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
func (p *PostgresTranslator) needsParentheses(node Node) bool {
	// Binary operations need parentheses when nested
	_, isBinaryOp := node.(*BinaryOp)
	return isBinaryOp
}

// translateRangeQuery translates range queries like field:[start TO end].
func (p *PostgresTranslator) translateRangeQuery(rq *RangeQuery, schema *schema.Schema) (string, error) {
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

// translateWildcardQuery translates wildcard queries to PostgreSQL LIKE patterns.
// Converts * to % (zero or more chars) and ? to _ (single char).
func (p *PostgresTranslator) translateWildcardQuery(wq *WildcardQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(wq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", wq.Field, schema.Name)
	}

	_ = field // Use field if needed later

	// Convert wildcard pattern to PostgreSQL LIKE pattern
	likePattern := convertWildcardToLike(wq.Pattern)

	// Add parameter
	p.paramCount++
	p.params = append(p.params, likePattern)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	// Generate SQL with LIKE operator
	return fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount), nil
}

// translateRegexQuery translates regex queries to PostgreSQL regex operator.
func (p *PostgresTranslator) translateRegexQuery(rq *RegexQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(rq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", rq.Field, schema.Name)
	}

	_ = field // Use field if needed later

	// Add parameter
	p.paramCount++
	p.params = append(p.params, rq.Pattern)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	// Generate SQL with PostgreSQL regex operator (~)
	return fmt.Sprintf("%s ~ $%d", columnName, p.paramCount), nil
}

// convertWildcardToLike converts wildcard patterns (* and ?) to PostgreSQL LIKE patterns (% and _).
// Also escapes special LIKE characters: %, _, and \.
func convertWildcardToLike(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern))

	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '*':
			// * becomes % (zero or more characters)
			result.WriteByte('%')
		case '?':
			// ? becomes _ (single character)
			result.WriteByte('_')
		case '%', '_', '\\':
			// Escape special LIKE characters
			result.WriteByte('\\')
			result.WriteByte(ch)
		default:
			result.WriteByte(ch)
		}
	}

	return result.String()
}
