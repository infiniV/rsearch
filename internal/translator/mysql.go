package translator

import (
	"fmt"
	"strings"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
)

// MySQLTranslator translates AST nodes to MySQL queries.
type MySQLTranslator struct {
	params     []interface{}
	paramTypes []string
	boosts     []map[string]interface{}
}

// NewMySQLTranslator creates a new MySQL translator.
func NewMySQLTranslator() *MySQLTranslator {
	return &MySQLTranslator{}
}

// DatabaseType returns the database type.
func (m *MySQLTranslator) DatabaseType() string {
	return "mysql"
}

// Translate converts an AST node to a MySQL query.
func (m *MySQLTranslator) Translate(ast parser.Node, schema *schema.Schema) (*TranslatorOutput, error) {
	// Reset state for new translation
	m.params = make([]interface{}, 0)
	m.paramTypes = make([]string, 0)
	m.boosts = make([]map[string]interface{}, 0)

	whereClause, err := m.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	output := NewSQLOutput(whereClause, m.params, m.paramTypes)

	// Add boost metadata if any boosts were collected
	if len(m.boosts) > 0 {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata["boosts"] = m.boosts
	}

	return output, nil
}

// translateNode recursively translates AST nodes.
func (m *MySQLTranslator) translateNode(node parser.Node, schema *schema.Schema) (string, error) {
	switch n := node.(type) {
	case *parser.FieldQuery:
		return m.translateFieldQuery(n, schema)
	case *parser.BinaryOp:
		return m.translateBinaryOp(n, schema)
	case *parser.RangeQuery:
		return m.translateRangeQuery(n, schema)
	case *parser.UnaryOp:
		return m.translateUnaryOp(n, schema)
	case *parser.ExistsQuery:
		return m.translateExistsQuery(n, schema)
	case *parser.BoostQuery:
		return m.translateBoostQuery(n, schema)
	case *parser.GroupQuery:
		return m.translateGroupQuery(n, schema)
	case *parser.RequiredQuery:
		return m.translateRequiredQuery(n, schema)
	case *parser.ProhibitedQuery:
		return m.translateProhibitedQuery(n, schema)
	case *parser.TermQuery:
		return m.translateTermQuery(n, schema)
	case *parser.PhraseQuery:
		return m.translatePhraseQuery(n, schema)
	case *parser.WildcardQuery:
		return m.translateWildcardQuery(n, schema)
	case *parser.FuzzyQuery:
		return m.translateFuzzyQuery(n, schema)
	case *parser.ProximityQuery:
		return m.translateProximityQuery(n, schema)
	case *parser.FieldGroupQuery:
		return m.translateFieldGroupQuery(n, schema)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

// translateFieldQuery translates a simple field:value query.
func (m *MySQLTranslator) translateFieldQuery(fq *parser.FieldQuery, schema *schema.Schema) (string, error) {
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
		m.params = append(m.params, pattern)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		return fmt.Sprintf("%s LIKE ?", columnName), nil

	case *parser.RegexValue:
		// Use MySQL REGEXP operator
		m.params = append(m.params, v.Pattern)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		return fmt.Sprintf("%s REGEXP ?", columnName), nil

	case *parser.PhraseValue:
		// Phrase is exact match
		m.params = append(m.params, v.Phrase)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = ?", columnName), nil

	default:
		// Simple equality
		value := fq.Value.Value()
		m.params = append(m.params, value)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = ?", columnName), nil
	}
}

// translateBinaryOp translates AND/OR operations.
func (m *MySQLTranslator) translateBinaryOp(bo *parser.BinaryOp, schema *schema.Schema) (string, error) {
	left, err := m.translateNode(bo.Left, schema)
	if err != nil {
		return "", err
	}

	right, err := m.translateNode(bo.Right, schema)
	if err != nil {
		return "", err
	}

	// Determine if we need parentheses
	leftNeedsParens := m.needsParentheses(bo.Left)
	rightNeedsParens := m.needsParentheses(bo.Right)

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
func (m *MySQLTranslator) needsParentheses(node parser.Node) bool {
	// Binary operations need parentheses when nested
	_, isBinaryOp := node.(*parser.BinaryOp)
	return isBinaryOp
}

// translateRangeQuery translates range queries like field:[start TO end].
func (m *MySQLTranslator) translateRangeQuery(rq *parser.RangeQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(rq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", rq.Field, schema.Name)
	}

	// Check for wildcard boundaries
	startIsWildcard := m.isWildcard(rq.Start)
	endIsWildcard := m.isWildcard(rq.End)

	// Handle fully bounded ranges (no wildcards)
	if !startIsWildcard && !endIsWildcard && rq.InclusiveStart && rq.InclusiveEnd {
		// Both inclusive: BETWEEN
		m.params = append(m.params, rq.Start.Value())
		m.paramTypes = append(m.paramTypes, string(field.Type))

		m.params = append(m.params, rq.End.Value())
		m.paramTypes = append(m.paramTypes, string(field.Type))

		return fmt.Sprintf("%s BETWEEN ? AND ?", columnName), nil
	}

	// Mixed or exclusive ranges, or unbounded ranges: use comparison operators
	var clauses []string

	// Start condition (if not wildcard)
	if !startIsWildcard {
		m.params = append(m.params, rq.Start.Value())
		m.paramTypes = append(m.paramTypes, string(field.Type))

		if rq.InclusiveStart {
			clauses = append(clauses, fmt.Sprintf("%s >= ?", columnName))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s > ?", columnName))
		}
	}

	// End condition (if not wildcard)
	if !endIsWildcard {
		m.params = append(m.params, rq.End.Value())
		m.paramTypes = append(m.paramTypes, string(field.Type))

		if rq.InclusiveEnd {
			clauses = append(clauses, fmt.Sprintf("%s <= ?", columnName))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s < ?", columnName))
		}
	}

	return strings.Join(clauses, " AND "), nil
}

// isWildcard checks if a ValueNode represents a wildcard (*).
func (m *MySQLTranslator) isWildcard(v parser.ValueNode) bool {
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
func (m *MySQLTranslator) translateUnaryOp(uo *parser.UnaryOp, schema *schema.Schema) (string, error) {
	operand, err := m.translateNode(uo.Operand, schema)
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
		if m.needsParenthesesForNot(uo.Operand, operand) {
			operand = fmt.Sprintf("(%s)", operand)
		}
		return fmt.Sprintf("NOT %s", operand), nil
	default:
		return "", fmt.Errorf("unsupported unary operator: %s", uo.Op)
	}
}

// needsParenthesesForNot determines if operand needs parentheses in NOT context
func (m *MySQLTranslator) needsParenthesesForNot(node parser.Node, sql string) bool {
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
func (m *MySQLTranslator) translateExistsQuery(eq *parser.ExistsQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(eq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", eq.Field, schema.Name)
	}

	// For JSON fields, need special handling
	if field.Type == "json" {
		// JSON fields: check IS NOT NULL and not the JSON null value
		return fmt.Sprintf("%s IS NOT NULL AND JSON_TYPE(%s) != 'NULL'", columnName, columnName), nil
	}

	// For regular fields: simple IS NOT NULL check
	return fmt.Sprintf("%s IS NOT NULL", columnName), nil
}

// translateBoostQuery translates boost queries (query^boost).
// For SQL databases, boost is stored in metadata; the SQL is the same as the wrapped query.
func (m *MySQLTranslator) translateBoostQuery(bq *parser.BoostQuery, schema *schema.Schema) (string, error) {
	// Translate the wrapped query
	sql, err := m.translateNode(bq.Query, schema)
	if err != nil {
		return "", err
	}

	// Store boost metadata with snake_case query type
	queryType := m.toSnakeCase(bq.Query.Type())
	boostInfo := map[string]interface{}{
		"query": queryType,
		"boost": bq.Boost,
	}
	m.boosts = append(m.boosts, boostInfo)

	return sql, nil
}

// toSnakeCase converts CamelCase to snake_case
func (m *MySQLTranslator) toSnakeCase(s string) string {
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
func (m *MySQLTranslator) translateGroupQuery(gq *parser.GroupQuery, schema *schema.Schema) (string, error) {
	inner, err := m.translateNode(gq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", inner), nil
}

// translateRequiredQuery translates +term (required term).
func (m *MySQLTranslator) translateRequiredQuery(rq *parser.RequiredQuery, schema *schema.Schema) (string, error) {
	// Required terms pass through - they must match
	return m.translateNode(rq.Query, schema)
}

// translateProhibitedQuery translates -term (prohibited term).
func (m *MySQLTranslator) translateProhibitedQuery(pq *parser.ProhibitedQuery, schema *schema.Schema) (string, error) {
	inner, err := m.translateNode(pq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("NOT %s", inner), nil
}

// translateTermQuery translates standalone terms (uses default field).
func (m *MySQLTranslator) translateTermQuery(tq *parser.TermQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone term '%s' requires a default field in schema", tq.Term)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	m.params = append(m.params, tq.Term)
	m.paramTypes = append(m.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = ?", columnName), nil
}

// translatePhraseQuery translates standalone phrases (uses default field).
func (m *MySQLTranslator) translatePhraseQuery(pq *parser.PhraseQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone phrase '%s' requires a default field in schema", pq.Phrase)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	m.params = append(m.params, pq.Phrase)
	m.paramTypes = append(m.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = ?", columnName), nil
}

// translateWildcardQuery translates standalone wildcards (uses default field).
func (m *MySQLTranslator) translateWildcardQuery(wq *parser.WildcardQuery, schema *schema.Schema) (string, error) {
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

	m.params = append(m.params, pattern)
	m.paramTypes = append(m.paramTypes, string(field.Type))
	return fmt.Sprintf("%s LIKE ?", columnName), nil
}

// translateFuzzyQuery translates fuzzy search queries (term~distance).
func (m *MySQLTranslator) translateFuzzyQuery(fq *parser.FuzzyQuery, schema *schema.Schema) (string, error) {
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
		return "", fmt.Errorf("fuzzy search requires SOUNDEX function. Enable in schema or use wildcards instead")
	}

	// MySQL uses SOUNDEX for fuzzy matching (phonetic similarity)
	// Note: This is different from Levenshtein distance but provides similar functionality
	m.params = append(m.params, fq.Term)
	m.paramTypes = append(m.paramTypes, string(field.Type))

	return fmt.Sprintf("SOUNDEX(%s) = SOUNDEX(?)", columnName), nil
}

// translateProximityQuery translates proximity search queries ("phrase"~distance).
func (m *MySQLTranslator) translateProximityQuery(pq *parser.ProximityQuery, schema *schema.Schema) (string, error) {
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

	// MySQL uses MATCH...AGAINST for full-text search
	// Note: The column must have a FULLTEXT index
	m.params = append(m.params, pq.Phrase)
	m.paramTypes = append(m.paramTypes, string(field.Type))

	return fmt.Sprintf("MATCH(%s) AGAINST(? IN BOOLEAN MODE)", columnName), nil
}

// translateFieldGroupQuery translates field:(value1 OR value2) queries.
func (m *MySQLTranslator) translateFieldGroupQuery(fgq *parser.FieldGroupQuery, schema *schema.Schema) (string, error) {
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
			m.params = append(m.params, inner.Term)
			m.paramTypes = append(m.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s = ?", columnName)
		case *parser.WildcardQuery:
			pattern := inner.Pattern
			pattern = strings.ReplaceAll(pattern, "*", "%")
			pattern = strings.ReplaceAll(pattern, "?", "_")
			m.params = append(m.params, pattern)
			m.paramTypes = append(m.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s LIKE ?", columnName)
		case *parser.BinaryOp:
			// Recursively translate the binary operation
			clause, err = m.translateFieldGroupBinaryOp(inner, columnName, field, schema)
			if err != nil {
				return "", err
			}
		default:
			clause, err = m.translateNode(q, schema)
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
func (m *MySQLTranslator) translateFieldGroupBinaryOp(bo *parser.BinaryOp, columnName string, field *schema.Field, schema *schema.Schema) (string, error) {
	var leftClause, rightClause string
	var err error

	// Handle left side
	switch left := bo.Left.(type) {
	case *parser.TermQuery:
		m.params = append(m.params, left.Term)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s = ?", columnName)
	case *parser.WildcardQuery:
		pattern := left.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		m.params = append(m.params, pattern)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s LIKE ?", columnName)
	case *parser.BinaryOp:
		leftClause, err = m.translateFieldGroupBinaryOp(left, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		leftClause, err = m.translateNode(bo.Left, schema)
		if err != nil {
			return "", err
		}
	}

	// Handle right side
	switch right := bo.Right.(type) {
	case *parser.TermQuery:
		m.params = append(m.params, right.Term)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s = ?", columnName)
	case *parser.WildcardQuery:
		pattern := right.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		m.params = append(m.params, pattern)
		m.paramTypes = append(m.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s LIKE ?", columnName)
	case *parser.BinaryOp:
		rightClause, err = m.translateFieldGroupBinaryOp(right, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		rightClause, err = m.translateNode(bo.Right, schema)
		if err != nil {
			return "", err
		}
	}

	operator := strings.ToUpper(bo.Op)
	return fmt.Sprintf("(%s %s %s)", leftClause, operator, rightClause), nil
}
