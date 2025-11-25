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
	boosts     []map[string]interface{}
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
	p.boosts = make([]map[string]interface{}, 0)

	whereClause, err := p.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	output := NewSQLOutput(whereClause, p.params, p.paramTypes)

	// Add boost metadata if any boosts were collected
	if len(p.boosts) > 0 {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata["boosts"] = p.boosts
	}

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
	case *parser.UnaryOp:
		return p.translateUnaryOp(n, schema)
	case *parser.ExistsQuery:
		return p.translateExistsQuery(n, schema)
	case *parser.BoostQuery:
		return p.translateBoostQuery(n, schema)
	case *parser.GroupQuery:
		return p.translateGroupQuery(n, schema)
	case *parser.RequiredQuery:
		return p.translateRequiredQuery(n, schema)
	case *parser.ProhibitedQuery:
		return p.translateProhibitedQuery(n, schema)
	case *parser.TermQuery:
		return p.translateTermQuery(n, schema)
	case *parser.PhraseQuery:
		return p.translatePhraseQuery(n, schema)
	case *parser.WildcardQuery:
		return p.translateWildcardQuery(n, schema)
	case *parser.FuzzyQuery:
		return p.translateFuzzyQuery(n, schema)
	case *parser.ProximityQuery:
		return p.translateProximityQuery(n, schema)
	case *parser.FieldGroupQuery:
		return p.translateFieldGroupQuery(n, schema)
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

	// Handle different value types
	switch v := fq.Value.(type) {
	case *parser.WildcardValue:
		// Convert wildcard pattern to LIKE pattern
		pattern := v.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		p.paramCount++
		p.params = append(p.params, pattern)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		return fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount), nil

	case *parser.RegexValue:
		// Use PostgreSQL regex operator
		p.paramCount++
		p.params = append(p.params, v.Pattern)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		return fmt.Sprintf("%s ~ $%d", columnName, p.paramCount), nil

	case *parser.PhraseValue:
		// Phrase is exact match
		p.paramCount++
		p.params = append(p.params, v.Phrase)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil

	default:
		// Simple equality
		value := fq.Value.Value()
		p.paramCount++
		p.params = append(p.params, value)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil
	}
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

	// Check for wildcard boundaries
	startIsWildcard := p.isWildcard(rq.Start)
	endIsWildcard := p.isWildcard(rq.End)

	// Handle fully bounded ranges (no wildcards)
	if !startIsWildcard && !endIsWildcard && rq.InclusiveStart && rq.InclusiveEnd {
		// Both inclusive: BETWEEN
		p.paramCount++
		p.params = append(p.params, rq.Start.Value())
		p.paramTypes = append(p.paramTypes, string(field.Type))

		p.paramCount++
		p.params = append(p.params, rq.End.Value())
		p.paramTypes = append(p.paramTypes, string(field.Type))

		return fmt.Sprintf("%s BETWEEN $%d AND $%d", columnName, p.paramCount-1, p.paramCount), nil
	}

	// Mixed or exclusive ranges, or unbounded ranges: use comparison operators
	var clauses []string

	// Start condition (if not wildcard)
	if !startIsWildcard {
		p.paramCount++
		p.params = append(p.params, rq.Start.Value())
		p.paramTypes = append(p.paramTypes, string(field.Type))

		if rq.InclusiveStart {
			clauses = append(clauses, fmt.Sprintf("%s >= $%d", columnName, p.paramCount))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s > $%d", columnName, p.paramCount))
		}
	}

	// End condition (if not wildcard)
	if !endIsWildcard {
		p.paramCount++
		p.params = append(p.params, rq.End.Value())
		p.paramTypes = append(p.paramTypes, string(field.Type))

		if rq.InclusiveEnd {
			clauses = append(clauses, fmt.Sprintf("%s <= $%d", columnName, p.paramCount))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s < $%d", columnName, p.paramCount))
		}
	}

	return strings.Join(clauses, " AND "), nil
}

// isWildcard checks if a ValueNode represents a wildcard (*).
func (p *PostgresTranslator) isWildcard(v parser.ValueNode) bool {
	if v == nil {
		return false
	}
	val := v.Value()
	if strVal, ok := val.(string); ok {
		return strVal == "*"
	}
	return false
}

// translateUnaryOp translates unary operations (+, -, NOT).
func (p *PostgresTranslator) translateUnaryOp(uo *parser.UnaryOp, schema *schema.Schema) (string, error) {
	operand, err := p.translateNode(uo.Operand, schema)
	if err != nil {
		return "", err
	}

	// Handle different operators
	switch uo.Op {
	case "+":
		// Required operator - just pass through
		return operand, nil
	case "-", "NOT":
		// Prohibited or NOT operator - add NOT
		// Wrap in parentheses if needed (for complex expressions or JSON exists checks)
		if p.needsParenthesesForNot(uo.Operand, operand) {
			operand = fmt.Sprintf("(%s)", operand)
		}
		return fmt.Sprintf("NOT %s", operand), nil
	default:
		return "", fmt.Errorf("unsupported unary operator: %s", uo.Op)
	}
}

// needsParenthesesForNot determines if operand needs parentheses in NOT context
func (p *PostgresTranslator) needsParenthesesForNot(node parser.Node, sql string) bool {
	// Binary operations always need parentheses
	if _, isBinaryOp := node.(*parser.BinaryOp); isBinaryOp {
		return true
	}
	// Multi-clause SQL (like JSON exists checks with AND) needs parentheses
	if strings.Contains(sql, " AND ") || strings.Contains(sql, " OR ") {
		return true
	}
	return false
}

// translateExistsQuery translates existence checks (_exists_:field).
func (p *PostgresTranslator) translateExistsQuery(eq *parser.ExistsQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(eq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", eq.Field, schema.Name)
	}

	// For JSON/JSONB fields, need special handling
	if field.Type == "json" {
		// JSON fields: check IS NOT NULL and not the JSON null value
		return fmt.Sprintf("%s IS NOT NULL AND %s != 'null'::jsonb", columnName, columnName), nil
	}

	// For regular fields: simple IS NOT NULL check
	return fmt.Sprintf("%s IS NOT NULL", columnName), nil
}

// translateBoostQuery translates boost queries (query^boost).
// For SQL databases, boost is stored in metadata; the SQL is the same as the wrapped query.
func (p *PostgresTranslator) translateBoostQuery(bq *parser.BoostQuery, schema *schema.Schema) (string, error) {
	// Translate the wrapped query
	sql, err := p.translateNode(bq.Query, schema)
	if err != nil {
		return "", err
	}

	// Store boost metadata with snake_case query type
	queryType := p.toSnakeCase(bq.Query.Type())
	boostInfo := map[string]interface{}{
		"query": queryType,
		"boost": bq.Boost,
	}
	p.boosts = append(p.boosts, boostInfo)

	return sql, nil
}

// toSnakeCase converts CamelCase to snake_case
func (p *PostgresTranslator) toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// translateGroupQuery translates parenthesized expressions.
func (p *PostgresTranslator) translateGroupQuery(gq *parser.GroupQuery, schema *schema.Schema) (string, error) {
	inner, err := p.translateNode(gq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", inner), nil
}

// translateRequiredQuery translates +term (required term).
func (p *PostgresTranslator) translateRequiredQuery(rq *parser.RequiredQuery, schema *schema.Schema) (string, error) {
	// Required terms pass through - they must match
	return p.translateNode(rq.Query, schema)
}

// translateProhibitedQuery translates -term (prohibited term).
func (p *PostgresTranslator) translateProhibitedQuery(pq *parser.ProhibitedQuery, schema *schema.Schema) (string, error) {
	inner, err := p.translateNode(pq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("NOT %s", inner), nil
}

// translateTermQuery translates standalone terms (uses default field).
func (p *PostgresTranslator) translateTermQuery(tq *parser.TermQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone term '%s' requires a default field in schema", tq.Term)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	p.paramCount++
	p.params = append(p.params, tq.Term)
	p.paramTypes = append(p.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil
}

// translatePhraseQuery translates standalone phrases (uses default field).
func (p *PostgresTranslator) translatePhraseQuery(pq *parser.PhraseQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone phrase '%s' requires a default field in schema", pq.Phrase)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	p.paramCount++
	p.params = append(p.params, pq.Phrase)
	p.paramTypes = append(p.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil
}

// translateWildcardQuery translates standalone wildcards (uses default field).
func (p *PostgresTranslator) translateWildcardQuery(wq *parser.WildcardQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone wildcard '%s' requires a default field in schema", wq.Pattern)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	// Convert wildcard pattern to LIKE pattern
	pattern := wq.Pattern
	pattern = strings.ReplaceAll(pattern, "*", "%")
	pattern = strings.ReplaceAll(pattern, "?", "_")

	p.paramCount++
	p.params = append(p.params, pattern)
	p.paramTypes = append(p.paramTypes, string(field.Type))
	return fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount), nil
}

// translateFuzzyQuery translates fuzzy search queries (term~distance).
func (p *PostgresTranslator) translateFuzzyQuery(fq *parser.FuzzyQuery, schema *schema.Schema) (string, error) {
	// Determine field - use provided field or default
	fieldName := fq.Field
	if fieldName == "" {
		if schema.Options.DefaultField == "" {
			return "", fmt.Errorf("fuzzy search '%s~%d' requires a field or default field in schema", fq.Term, fq.Distance)
		}
		fieldName = schema.Options.DefaultField
	}

	columnName, field, err := schema.ResolveField(fieldName)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", fieldName, schema.Name)
	}

	// Check if fuzzy search is enabled
	if !schema.Options.EnabledFeatures.Fuzzy {
		return "", fmt.Errorf("fuzzy search requires pg_trgm extension. Enable in schema or use wildcards instead")
	}

	// PostgreSQL with pg_trgm: use similarity or levenshtein
	p.paramCount++
	p.params = append(p.params, fq.Term)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	p.paramCount++
	p.params = append(p.params, fq.Distance)
	p.paramTypes = append(p.paramTypes, "integer")

	return fmt.Sprintf("levenshtein(%s, $%d) <= $%d", columnName, p.paramCount-1, p.paramCount), nil
}

// translateProximityQuery translates proximity search queries ("phrase"~distance).
func (p *PostgresTranslator) translateProximityQuery(pq *parser.ProximityQuery, schema *schema.Schema) (string, error) {
	// Determine field - use provided field or default
	fieldName := pq.Field
	if fieldName == "" {
		if schema.Options.DefaultField == "" {
			return "", fmt.Errorf("proximity search requires a field or default field in schema")
		}
		fieldName = schema.Options.DefaultField
	}

	columnName, field, err := schema.ResolveField(fieldName)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", fieldName, schema.Name)
	}

	// Check if proximity search is enabled
	if !schema.Options.EnabledFeatures.Proximity {
		return "", fmt.Errorf("proximity search requires full-text search. Enable in schema or use phrase match instead")
	}

	// PostgreSQL with full-text search: use to_tsvector and <N> operator
	words := strings.Fields(pq.Phrase)
	if len(words) < 2 {
		// Fall back to simple phrase match
		p.paramCount++
		p.params = append(p.params, pq.Phrase)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = $%d", columnName, p.paramCount), nil
	}

	// Build tsquery with proximity
	p.paramCount++
	p.params = append(p.params, pq.Phrase)
	p.paramTypes = append(p.paramTypes, string(field.Type))

	return fmt.Sprintf("to_tsvector('english', %s) @@ phraseto_tsquery('english', $%d)", columnName, p.paramCount), nil
}

// translateFieldGroupQuery translates field:(value1 OR value2) queries.
func (p *PostgresTranslator) translateFieldGroupQuery(fgq *parser.FieldGroupQuery, schema *schema.Schema) (string, error) {
	if len(fgq.Queries) == 0 {
		return "", fmt.Errorf("empty field group query")
	}

	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(fgq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", fgq.Field, schema.Name)
	}

	// Translate each inner query, wrapping terms as field queries
	var clauses []string
	for _, q := range fgq.Queries {
		var clause string
		var err error

		switch inner := q.(type) {
		case *parser.TermQuery:
			// Convert term to field query
			p.paramCount++
			p.params = append(p.params, inner.Term)
			p.paramTypes = append(p.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s = $%d", columnName, p.paramCount)
		case *parser.WildcardQuery:
			pattern := inner.Pattern
			pattern = strings.ReplaceAll(pattern, "*", "%")
			pattern = strings.ReplaceAll(pattern, "?", "_")
			p.paramCount++
			p.params = append(p.params, pattern)
			p.paramTypes = append(p.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount)
		case *parser.BinaryOp:
			// Recursively translate the binary operation
			clause, err = p.translateFieldGroupBinaryOp(inner, columnName, field, schema)
			if err != nil {
				return "", err
			}
		default:
			clause, err = p.translateNode(q, schema)
			if err != nil {
				return "", err
			}
		}
		clauses = append(clauses, clause)
	}

	if len(clauses) == 1 {
		return clauses[0], nil
	}

	// Join with OR (default for field groups)
	return fmt.Sprintf("(%s)", strings.Join(clauses, " OR ")), nil
}

// translateFieldGroupBinaryOp handles binary operations within field groups.
func (p *PostgresTranslator) translateFieldGroupBinaryOp(bo *parser.BinaryOp, columnName string, field *schema.Field, schema *schema.Schema) (string, error) {
	var leftClause, rightClause string
	var err error

	// Handle left side
	switch left := bo.Left.(type) {
	case *parser.TermQuery:
		p.paramCount++
		p.params = append(p.params, left.Term)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s = $%d", columnName, p.paramCount)
	case *parser.WildcardQuery:
		pattern := left.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		p.paramCount++
		p.params = append(p.params, pattern)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount)
	case *parser.BinaryOp:
		leftClause, err = p.translateFieldGroupBinaryOp(left, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		leftClause, err = p.translateNode(bo.Left, schema)
		if err != nil {
			return "", err
		}
	}

	// Handle right side
	switch right := bo.Right.(type) {
	case *parser.TermQuery:
		p.paramCount++
		p.params = append(p.params, right.Term)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s = $%d", columnName, p.paramCount)
	case *parser.WildcardQuery:
		pattern := right.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		p.paramCount++
		p.params = append(p.params, pattern)
		p.paramTypes = append(p.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount)
	case *parser.BinaryOp:
		rightClause, err = p.translateFieldGroupBinaryOp(right, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		rightClause, err = p.translateNode(bo.Right, schema)
		if err != nil {
			return "", err
		}
	}

	operator := strings.ToUpper(bo.Op)
	return fmt.Sprintf("(%s %s %s)", leftClause, operator, rightClause), nil
}
