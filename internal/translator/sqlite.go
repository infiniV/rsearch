package translator

import (
	"fmt"
	"strings"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
)

// SQLiteTranslator translates AST nodes to SQLite queries.
type SQLiteTranslator struct {
	params     []interface{}
	paramTypes []string
	boosts     []map[string]interface{}
}

// NewSQLiteTranslator creates a new SQLite translator.
func NewSQLiteTranslator() *SQLiteTranslator {
	return &SQLiteTranslator{}
}

// DatabaseType returns the database type.
func (s *SQLiteTranslator) DatabaseType() string {
	return "sqlite"
}

// Translate converts an AST node to a SQLite query.
func (s *SQLiteTranslator) Translate(ast parser.Node, schema *schema.Schema) (*TranslatorOutput, error) {
	// Reset state for new translation
	s.params = make([]interface{}, 0)
	s.paramTypes = make([]string, 0)
	s.boosts = make([]map[string]interface{}, 0)

	whereClause, err := s.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	output := NewSQLOutput(whereClause, s.params, s.paramTypes)

	// Add boost metadata if any boosts were collected
	if len(s.boosts) > 0 {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata["boosts"] = s.boosts
	}

	return output, nil
}

// translateNode recursively translates AST nodes.
func (s *SQLiteTranslator) translateNode(node parser.Node, schema *schema.Schema) (string, error) {
	switch n := node.(type) {
	case *parser.FieldQuery:
		return s.translateFieldQuery(n, schema)
	case *parser.BinaryOp:
		return s.translateBinaryOp(n, schema)
	case *parser.RangeQuery:
		return s.translateRangeQuery(n, schema)
	case *parser.UnaryOp:
		return s.translateUnaryOp(n, schema)
	case *parser.ExistsQuery:
		return s.translateExistsQuery(n, schema)
	case *parser.BoostQuery:
		return s.translateBoostQuery(n, schema)
	case *parser.GroupQuery:
		return s.translateGroupQuery(n, schema)
	case *parser.RequiredQuery:
		return s.translateRequiredQuery(n, schema)
	case *parser.ProhibitedQuery:
		return s.translateProhibitedQuery(n, schema)
	case *parser.TermQuery:
		return s.translateTermQuery(n, schema)
	case *parser.PhraseQuery:
		return s.translatePhraseQuery(n, schema)
	case *parser.WildcardQuery:
		return s.translateWildcardQuery(n, schema)
	case *parser.FuzzyQuery:
		return s.translateFuzzyQuery(n, schema)
	case *parser.ProximityQuery:
		return s.translateProximityQuery(n, schema)
	case *parser.FieldGroupQuery:
		return s.translateFieldGroupQuery(n, schema)
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

// translateFieldQuery translates a simple field:value query.
func (s *SQLiteTranslator) translateFieldQuery(fq *parser.FieldQuery, schema *schema.Schema) (string, error) {
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
		s.params = append(s.params, pattern)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		return fmt.Sprintf("%s LIKE ?", columnName), nil

	case *parser.RegexValue:
		// Use SQLite REGEXP operator (requires user-defined function)
		s.params = append(s.params, v.Pattern)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		return fmt.Sprintf("%s REGEXP ?", columnName), nil

	case *parser.PhraseValue:
		// Phrase is exact match
		s.params = append(s.params, v.Phrase)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = ?", columnName), nil

	default:
		// Simple equality
		value := fq.Value.Value()
		s.params = append(s.params, value)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		return fmt.Sprintf("%s = ?", columnName), nil
	}
}

// translateBinaryOp translates AND/OR operations.
func (s *SQLiteTranslator) translateBinaryOp(bo *parser.BinaryOp, schema *schema.Schema) (string, error) {
	left, err := s.translateNode(bo.Left, schema)
	if err != nil {
		return "", err
	}

	right, err := s.translateNode(bo.Right, schema)
	if err != nil {
		return "", err
	}

	// Determine if we need parentheses
	leftNeedsParens := s.needsParentheses(bo.Left)
	rightNeedsParens := s.needsParentheses(bo.Right)

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
func (s *SQLiteTranslator) needsParentheses(node parser.Node) bool {
	// Binary operations need parentheses when nested
	_, isBinaryOp := node.(*parser.BinaryOp)
	return isBinaryOp
}

// translateRangeQuery translates range queries like field:[start TO end].
func (s *SQLiteTranslator) translateRangeQuery(rq *parser.RangeQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(rq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", rq.Field, schema.Name)
	}

	// Check for wildcard boundaries
	startIsWildcard := s.isWildcard(rq.Start)
	endIsWildcard := s.isWildcard(rq.End)

	// Handle fully bounded ranges (no wildcards)
	if !startIsWildcard && !endIsWildcard && rq.InclusiveStart && rq.InclusiveEnd {
		// Both inclusive: BETWEEN
		s.params = append(s.params, rq.Start.Value())
		s.paramTypes = append(s.paramTypes, string(field.Type))

		s.params = append(s.params, rq.End.Value())
		s.paramTypes = append(s.paramTypes, string(field.Type))

		return fmt.Sprintf("%s BETWEEN ? AND ?", columnName), nil
	}

	// Mixed or exclusive ranges, or unbounded ranges: use comparison operators
	var clauses []string

	// Start condition (if not wildcard)
	if !startIsWildcard {
		s.params = append(s.params, rq.Start.Value())
		s.paramTypes = append(s.paramTypes, string(field.Type))

		if rq.InclusiveStart {
			clauses = append(clauses, fmt.Sprintf("%s >= ?", columnName))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s > ?", columnName))
		}
	}

	// End condition (if not wildcard)
	if !endIsWildcard {
		s.params = append(s.params, rq.End.Value())
		s.paramTypes = append(s.paramTypes, string(field.Type))

		if rq.InclusiveEnd {
			clauses = append(clauses, fmt.Sprintf("%s <= ?", columnName))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s < ?", columnName))
		}
	}

	return strings.Join(clauses, " AND "), nil
}

// isWildcard checks if a ValueNode represents a wildcard (*).
func (s *SQLiteTranslator) isWildcard(v parser.ValueNode) bool {
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
func (s *SQLiteTranslator) translateUnaryOp(uo *parser.UnaryOp, schema *schema.Schema) (string, error) {
	operand, err := s.translateNode(uo.Operand, schema)
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
		if s.needsParenthesesForNot(uo.Operand, operand) {
			operand = fmt.Sprintf("(%s)", operand)
		}
		return fmt.Sprintf("NOT %s", operand), nil
	default:
		return "", fmt.Errorf("unsupported unary operator: %s", uo.Op)
	}
}

// needsParenthesesForNot determines if operand needs parentheses in NOT context
func (s *SQLiteTranslator) needsParenthesesForNot(node parser.Node, sql string) bool {
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
func (s *SQLiteTranslator) translateExistsQuery(eq *parser.ExistsQuery, schema *schema.Schema) (string, error) {
	// Validate field exists in schema
	columnName, field, err := schema.ResolveField(eq.Field)
	if err != nil {
		return "", fmt.Errorf("field %s not found in schema %s", eq.Field, schema.Name)
	}

	// For JSON fields, need special handling
	if field.Type == "json" {
		// JSON fields: check IS NOT NULL and json_extract is not null
		return fmt.Sprintf("%s IS NOT NULL AND json_extract(%s, '$') IS NOT NULL", columnName, columnName), nil
	}

	// For regular fields: simple IS NOT NULL check
	return fmt.Sprintf("%s IS NOT NULL", columnName), nil
}

// translateBoostQuery translates boost queries (query^boost).
// For SQL databases, boost is stored in metadata; the SQL is the same as the wrapped query.
func (s *SQLiteTranslator) translateBoostQuery(bq *parser.BoostQuery, schema *schema.Schema) (string, error) {
	// Translate the wrapped query
	sql, err := s.translateNode(bq.Query, schema)
	if err != nil {
		return "", err
	}

	// Store boost metadata with snake_case query type
	queryType := s.toSnakeCase(bq.Query.Type())
	boostInfo := map[string]interface{}{
		"query": queryType,
		"boost": bq.Boost,
	}
	s.boosts = append(s.boosts, boostInfo)

	return sql, nil
}

// toSnakeCase converts CamelCase to snake_case
func (s *SQLiteTranslator) toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
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
func (s *SQLiteTranslator) translateGroupQuery(gq *parser.GroupQuery, schema *schema.Schema) (string, error) {
	inner, err := s.translateNode(gq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", inner), nil
}

// translateRequiredQuery translates +term (required term).
func (s *SQLiteTranslator) translateRequiredQuery(rq *parser.RequiredQuery, schema *schema.Schema) (string, error) {
	// Required terms pass through - they must match
	return s.translateNode(rq.Query, schema)
}

// translateProhibitedQuery translates -term (prohibited term).
func (s *SQLiteTranslator) translateProhibitedQuery(pq *parser.ProhibitedQuery, schema *schema.Schema) (string, error) {
	inner, err := s.translateNode(pq.Query, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("NOT %s", inner), nil
}

// translateTermQuery translates standalone terms (uses default field).
func (s *SQLiteTranslator) translateTermQuery(tq *parser.TermQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone term '%s' requires a default field in schema", tq.Term)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	s.params = append(s.params, tq.Term)
	s.paramTypes = append(s.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = ?", columnName), nil
}

// translatePhraseQuery translates standalone phrases (uses default field).
func (s *SQLiteTranslator) translatePhraseQuery(pq *parser.PhraseQuery, schema *schema.Schema) (string, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return "", fmt.Errorf("standalone phrase '%s' requires a default field in schema", pq.Phrase)
	}

	columnName, field, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return "", fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	s.params = append(s.params, pq.Phrase)
	s.paramTypes = append(s.paramTypes, string(field.Type))
	return fmt.Sprintf("%s = ?", columnName), nil
}

// translateWildcardQuery translates standalone wildcards (uses default field).
func (s *SQLiteTranslator) translateWildcardQuery(wq *parser.WildcardQuery, schema *schema.Schema) (string, error) {
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

	s.params = append(s.params, pattern)
	s.paramTypes = append(s.paramTypes, string(field.Type))
	return fmt.Sprintf("%s LIKE ?", columnName), nil
}

// translateFuzzyQuery translates fuzzy search queries (term~distance).
func (s *SQLiteTranslator) translateFuzzyQuery(fq *parser.FuzzyQuery, schema *schema.Schema) (string, error) {
	// SQLite does not have built-in fuzzy search support
	return "", fmt.Errorf("fuzzy search not supported in SQLite. Use wildcard patterns instead (e.g., '%s*')", fq.Term)
}

// translateProximityQuery translates proximity search queries ("phrase"~distance).
func (s *SQLiteTranslator) translateProximityQuery(pq *parser.ProximityQuery, schema *schema.Schema) (string, error) {
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
		return "", fmt.Errorf("proximity search requires FTS5. Enable in schema or use phrase match instead")
	}

	// SQLite FTS5 uses NEAR() function for proximity
	// Format: NEAR(term1 term2, N) where N is maximum distance
	nearQuery := fmt.Sprintf("NEAR(%s, %d)", pq.Phrase, pq.Distance)
	s.params = append(s.params, nearQuery)
	s.paramTypes = append(s.paramTypes, string(field.Type))

	return fmt.Sprintf("%s MATCH ?", columnName), nil
}

// translateFieldGroupQuery translates field:(value1 OR value2) queries.
func (s *SQLiteTranslator) translateFieldGroupQuery(fgq *parser.FieldGroupQuery, schema *schema.Schema) (string, error) {
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
			s.params = append(s.params, inner.Term)
			s.paramTypes = append(s.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s = ?", columnName)
		case *parser.WildcardQuery:
			pattern := inner.Pattern
			pattern = strings.ReplaceAll(pattern, "*", "%")
			pattern = strings.ReplaceAll(pattern, "?", "_")
			s.params = append(s.params, pattern)
			s.paramTypes = append(s.paramTypes, string(field.Type))
			clause = fmt.Sprintf("%s LIKE ?", columnName)
		case *parser.BinaryOp:
			// Recursively translate the binary operation
			clause, err = s.translateFieldGroupBinaryOp(inner, columnName, field, schema)
			if err != nil {
				return "", err
			}
		default:
			clause, err = s.translateNode(q, schema)
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
func (s *SQLiteTranslator) translateFieldGroupBinaryOp(bo *parser.BinaryOp, columnName string, field *schema.Field, schema *schema.Schema) (string, error) {
	var leftClause, rightClause string
	var err error

	// Handle left side
	switch left := bo.Left.(type) {
	case *parser.TermQuery:
		s.params = append(s.params, left.Term)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s = ?", columnName)
	case *parser.WildcardQuery:
		pattern := left.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		s.params = append(s.params, pattern)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		leftClause = fmt.Sprintf("%s LIKE ?", columnName)
	case *parser.BinaryOp:
		leftClause, err = s.translateFieldGroupBinaryOp(left, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		leftClause, err = s.translateNode(bo.Left, schema)
		if err != nil {
			return "", err
		}
	}

	// Handle right side
	switch right := bo.Right.(type) {
	case *parser.TermQuery:
		s.params = append(s.params, right.Term)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s = ?", columnName)
	case *parser.WildcardQuery:
		pattern := right.Pattern
		pattern = strings.ReplaceAll(pattern, "*", "%")
		pattern = strings.ReplaceAll(pattern, "?", "_")
		s.params = append(s.params, pattern)
		s.paramTypes = append(s.paramTypes, string(field.Type))
		rightClause = fmt.Sprintf("%s LIKE ?", columnName)
	case *parser.BinaryOp:
		rightClause, err = s.translateFieldGroupBinaryOp(right, columnName, field, schema)
		if err != nil {
			return "", err
		}
	default:
		rightClause, err = s.translateNode(bo.Right, schema)
		if err != nil {
			return "", err
		}
	}

	operator := strings.ToUpper(bo.Op)
	return fmt.Sprintf("(%s %s %s)", leftClause, operator, rightClause), nil
}
