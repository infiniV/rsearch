package translator

import (
	"fmt"
	"strings"

	"github.com/infiniv/rsearch/internal/parser"
	"github.com/infiniv/rsearch/internal/schema"
)

// MongoDBTranslator translates AST nodes to MongoDB query filters.
type MongoDBTranslator struct {
	boosts   []map[string]interface{}
	metadata map[string]interface{}
}

// NewMongoDBTranslator creates a new MongoDB translator.
func NewMongoDBTranslator() *MongoDBTranslator {
	return &MongoDBTranslator{}
}

// DatabaseType returns the database type.
func (m *MongoDBTranslator) DatabaseType() string {
	return "mongodb"
}

// Translate converts an AST node to a MongoDB query filter.
func (m *MongoDBTranslator) Translate(ast parser.Node, schema *schema.Schema) (*TranslatorOutput, error) {
	// Reset state for new translation
	m.boosts = make([]map[string]interface{}, 0)
	m.metadata = make(map[string]interface{})

	filter, err := m.translateNode(ast, schema)
	if err != nil {
		return nil, err
	}

	output := NewMongoDBOutput(filter)

	// Add boost metadata if any boosts were collected
	if len(m.boosts) > 0 {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata["boosts"] = m.boosts
	}

	// Add any additional metadata
	for k, v := range m.metadata {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata[k] = v
	}

	return output, nil
}

// translateNode recursively translates AST nodes to MongoDB filters.
func (m *MongoDBTranslator) translateNode(node parser.Node, schema *schema.Schema) (interface{}, error) {
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
		return nil, fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

// translateFieldQuery translates a simple field:value query.
func (m *MongoDBTranslator) translateFieldQuery(fq *parser.FieldQuery, schema *schema.Schema) (interface{}, error) {
	// Validate field exists in schema
	columnName, _, err := schema.ResolveField(fq.Field)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", fq.Field, schema.Name)
	}

	// Handle different value types
	switch v := fq.Value.(type) {
	case *parser.WildcardValue:
		// Convert wildcard pattern to regex pattern
		pattern := m.wildcardToRegex(v.Pattern)
		return map[string]interface{}{
			columnName: map[string]interface{}{
				"$regex": pattern,
			},
		}, nil

	case *parser.RegexValue:
		// Use MongoDB regex operator
		return map[string]interface{}{
			columnName: map[string]interface{}{
				"$regex":   v.Pattern,
				"$options": "",
			},
		}, nil

	case *parser.PhraseValue:
		// Phrase is exact match
		return map[string]interface{}{
			columnName: v.Phrase,
		}, nil

	default:
		// Simple equality
		value := fq.Value.Value()
		return map[string]interface{}{
			columnName: value,
		}, nil
	}
}

// wildcardToRegex converts wildcard pattern to regex pattern.
func (m *MongoDBTranslator) wildcardToRegex(pattern string) string {
	// Escape special regex characters except * and ?
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "+", "\\+")
	pattern = strings.ReplaceAll(pattern, "^", "\\^")
	pattern = strings.ReplaceAll(pattern, "$", "\\$")
	pattern = strings.ReplaceAll(pattern, "(", "\\(")
	pattern = strings.ReplaceAll(pattern, ")", "\\)")
	pattern = strings.ReplaceAll(pattern, "[", "\\[")
	pattern = strings.ReplaceAll(pattern, "]", "\\]")
	pattern = strings.ReplaceAll(pattern, "{", "\\{")
	pattern = strings.ReplaceAll(pattern, "}", "\\}")
	pattern = strings.ReplaceAll(pattern, "|", "\\|")
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")

	// Convert wildcards
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")

	// Anchor the pattern
	return "^" + pattern + "$"
}

// translateBinaryOp translates AND/OR operations.
func (m *MongoDBTranslator) translateBinaryOp(bo *parser.BinaryOp, schema *schema.Schema) (interface{}, error) {
	left, err := m.translateNode(bo.Left, schema)
	if err != nil {
		return nil, err
	}

	right, err := m.translateNode(bo.Right, schema)
	if err != nil {
		return nil, err
	}

	operator := strings.ToLower(bo.Op)
	if operator == "and" {
		return map[string]interface{}{
			"$and": []interface{}{left, right},
		}, nil
	} else if operator == "or" {
		return map[string]interface{}{
			"$or": []interface{}{left, right},
		}, nil
	}

	return nil, fmt.Errorf("unsupported binary operator: %s", bo.Op)
}

// translateRangeQuery translates range queries like field:[start TO end].
func (m *MongoDBTranslator) translateRangeQuery(rq *parser.RangeQuery, schema *schema.Schema) (interface{}, error) {
	// Validate field exists in schema
	columnName, _, err := schema.ResolveField(rq.Field)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", rq.Field, schema.Name)
	}

	// Check for wildcard boundaries
	startIsWildcard := m.isWildcard(rq.Start)
	endIsWildcard := m.isWildcard(rq.End)

	rangeFilter := make(map[string]interface{})

	// Start condition (if not wildcard)
	if !startIsWildcard {
		if rq.InclusiveStart {
			rangeFilter["$gte"] = rq.Start.Value()
		} else {
			rangeFilter["$gt"] = rq.Start.Value()
		}
	}

	// End condition (if not wildcard)
	if !endIsWildcard {
		if rq.InclusiveEnd {
			rangeFilter["$lte"] = rq.End.Value()
		} else {
			rangeFilter["$lt"] = rq.End.Value()
		}
	}

	return map[string]interface{}{
		columnName: rangeFilter,
	}, nil
}

// isWildcard checks if a ValueNode represents a wildcard (*).
func (m *MongoDBTranslator) isWildcard(v parser.ValueNode) bool {
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
func (m *MongoDBTranslator) translateUnaryOp(uo *parser.UnaryOp, schema *schema.Schema) (interface{}, error) {
	operand, err := m.translateNode(uo.Operand, schema)
	if err != nil {
		return nil, err
	}

	// Handle different operators
	switch uo.Op {
	case "+":
		// Required operator - just pass through
		return operand, nil
	case "-", "NOT":
		// Prohibited or NOT operator
		// For simple field queries, use $ne
		if operandMap, ok := operand.(map[string]interface{}); ok {
			// Check if it's a simple field equality
			if len(operandMap) == 1 {
				for field, value := range operandMap {
					// Skip if field starts with $ (operator)
					if !strings.HasPrefix(field, "$") {
						// Simple field equality - use $ne
						if _, ok := value.(map[string]interface{}); ok {
							// Already has operators, wrap in $nor
							return map[string]interface{}{
								"$nor": []interface{}{operand},
							}, nil
						}
						// Simple value - use $ne
						return map[string]interface{}{
							field: map[string]interface{}{
								"$ne": value,
							},
						}, nil
					}
				}
			}
		}
		// Complex expressions - use $nor
		return map[string]interface{}{
			"$nor": []interface{}{operand},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported unary operator: %s", uo.Op)
	}
}

// translateExistsQuery translates existence checks (_exists_:field).
func (m *MongoDBTranslator) translateExistsQuery(eq *parser.ExistsQuery, schema *schema.Schema) (interface{}, error) {
	// Validate field exists in schema
	columnName, _, err := schema.ResolveField(eq.Field)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", eq.Field, schema.Name)
	}

	// MongoDB exists check - field exists and is not null
	return map[string]interface{}{
		columnName: map[string]interface{}{
			"$exists": true,
			"$ne":     nil,
		},
	}, nil
}

// translateBoostQuery translates boost queries (query^boost).
// For MongoDB, boost is stored in metadata; the filter is the same as the wrapped query.
func (m *MongoDBTranslator) translateBoostQuery(bq *parser.BoostQuery, schema *schema.Schema) (interface{}, error) {
	// Translate the wrapped query
	filter, err := m.translateNode(bq.Query, schema)
	if err != nil {
		return nil, err
	}

	// Store boost metadata with snake_case query type
	queryType := m.toSnakeCase(bq.Query.Type())
	boostInfo := map[string]interface{}{
		"query": queryType,
		"boost": bq.Boost,
	}
	m.boosts = append(m.boosts, boostInfo)

	return filter, nil
}

// toSnakeCase converts CamelCase to snake_case
func (m *MongoDBTranslator) toSnakeCase(s string) string {
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
func (m *MongoDBTranslator) translateGroupQuery(gq *parser.GroupQuery, schema *schema.Schema) (interface{}, error) {
	// Group queries in MongoDB just pass through the inner query
	return m.translateNode(gq.Query, schema)
}

// translateRequiredQuery translates +term (required term).
func (m *MongoDBTranslator) translateRequiredQuery(rq *parser.RequiredQuery, schema *schema.Schema) (interface{}, error) {
	// Required terms pass through - they must match
	return m.translateNode(rq.Query, schema)
}

// translateProhibitedQuery translates -term (prohibited term).
func (m *MongoDBTranslator) translateProhibitedQuery(pq *parser.ProhibitedQuery, schema *schema.Schema) (interface{}, error) {
	inner, err := m.translateNode(pq.Query, schema)
	if err != nil {
		return nil, err
	}

	// For simple field queries, use $ne
	if innerMap, ok := inner.(map[string]interface{}); ok {
		if len(innerMap) == 1 {
			for field, value := range innerMap {
				if !strings.HasPrefix(field, "$") {
					// Simple field equality - use $ne
					if _, ok := value.(map[string]interface{}); ok {
						// Already has operators, wrap in $nor
						return map[string]interface{}{
							"$nor": []interface{}{inner},
						}, nil
					}
					return map[string]interface{}{
						field: map[string]interface{}{
							"$ne": value,
						},
					}, nil
				}
			}
		}
	}

	// Complex expressions - use $nor
	return map[string]interface{}{
		"$nor": []interface{}{inner},
	}, nil
}

// translateTermQuery translates standalone terms (uses default field).
func (m *MongoDBTranslator) translateTermQuery(tq *parser.TermQuery, schema *schema.Schema) (interface{}, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return nil, fmt.Errorf("standalone term '%s' requires a default field in schema", tq.Term)
	}

	columnName, _, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return nil, fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	return map[string]interface{}{
		columnName: tq.Term,
	}, nil
}

// translatePhraseQuery translates standalone phrases (uses default field).
func (m *MongoDBTranslator) translatePhraseQuery(pq *parser.PhraseQuery, schema *schema.Schema) (interface{}, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return nil, fmt.Errorf("standalone phrase '%s' requires a default field in schema", pq.Phrase)
	}

	columnName, _, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return nil, fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	return map[string]interface{}{
		columnName: pq.Phrase,
	}, nil
}

// translateWildcardQuery translates standalone wildcards (uses default field).
func (m *MongoDBTranslator) translateWildcardQuery(wq *parser.WildcardQuery, schema *schema.Schema) (interface{}, error) {
	// Use default field from schema options
	if schema.Options.DefaultField == "" {
		return nil, fmt.Errorf("standalone wildcard '%s' requires a default field in schema", wq.Pattern)
	}

	columnName, _, err := schema.ResolveField(schema.Options.DefaultField)
	if err != nil {
		return nil, fmt.Errorf("default field %s not found in schema %s", schema.Options.DefaultField, schema.Name)
	}

	// Convert wildcard pattern to regex pattern
	pattern := m.wildcardToRegex(wq.Pattern)

	return map[string]interface{}{
		columnName: map[string]interface{}{
			"$regex": pattern,
		},
	}, nil
}

// translateFuzzyQuery translates fuzzy search queries (term~distance).
func (m *MongoDBTranslator) translateFuzzyQuery(fq *parser.FuzzyQuery, schema *schema.Schema) (interface{}, error) {
	// Determine field - use provided field or default
	fieldName := fq.Field
	if fieldName == "" {
		if schema.Options.DefaultField == "" {
			return nil, fmt.Errorf("fuzzy search '%s~%d' requires a field or default field in schema", fq.Term, fq.Distance)
		}
		fieldName = schema.Options.DefaultField
	}

	_, _, err := schema.ResolveField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", fieldName, schema.Name)
	}

	// Check if fuzzy search is enabled
	if !schema.Options.EnabledFeatures.Fuzzy {
		return nil, fmt.Errorf("fuzzy search requires text index. Enable in schema or use wildcards instead")
	}

	// MongoDB with text index: use $text search
	filter := map[string]interface{}{
		"$text": map[string]interface{}{
			"$search": fq.Term,
		},
	}

	// Store fuzzy distance in metadata for reference
	m.metadata["fuzzy_distance"] = fq.Distance

	return filter, nil
}

// translateProximityQuery translates proximity search queries ("phrase"~distance).
func (m *MongoDBTranslator) translateProximityQuery(pq *parser.ProximityQuery, schema *schema.Schema) (interface{}, error) {
	// Determine field - use provided field or default
	fieldName := pq.Field
	if fieldName == "" {
		if schema.Options.DefaultField == "" {
			return nil, fmt.Errorf("proximity search requires a field or default field in schema")
		}
		fieldName = schema.Options.DefaultField
	}

	_, _, err := schema.ResolveField(fieldName)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", fieldName, schema.Name)
	}

	// Check if proximity search is enabled
	if !schema.Options.EnabledFeatures.Proximity {
		return nil, fmt.Errorf("proximity search requires text index. Enable in schema or use phrase match instead")
	}

	// MongoDB with text index: use $text search with phrase
	// Wrap phrase in quotes for exact phrase matching
	searchPhrase := fmt.Sprintf("\"%s\"", pq.Phrase)

	filter := map[string]interface{}{
		"$text": map[string]interface{}{
			"$search": searchPhrase,
		},
	}

	// Store proximity distance in metadata for reference
	m.metadata["proximity_distance"] = pq.Distance

	return filter, nil
}

// translateFieldGroupQuery translates field:(value1 OR value2) queries.
func (m *MongoDBTranslator) translateFieldGroupQuery(fgq *parser.FieldGroupQuery, schema *schema.Schema) (interface{}, error) {
	if len(fgq.Queries) == 0 {
		return nil, fmt.Errorf("empty field group query")
	}

	// Validate field exists in schema
	columnName, _, err := schema.ResolveField(fgq.Field)
	if err != nil {
		return nil, fmt.Errorf("field %s not found in schema %s", fgq.Field, schema.Name)
	}

	// Translate each inner query, wrapping terms as field queries
	var filters []interface{}
	for _, q := range fgq.Queries {
		var filter interface{}
		var err error

		switch inner := q.(type) {
		case *parser.TermQuery:
			// Convert term to field query
			filter = map[string]interface{}{
				columnName: inner.Term,
			}
		case *parser.WildcardQuery:
			pattern := m.wildcardToRegex(inner.Pattern)
			filter = map[string]interface{}{
				columnName: map[string]interface{}{
					"$regex": pattern,
				},
			}
		case *parser.BinaryOp:
			// Recursively translate the binary operation
			filter, err = m.translateFieldGroupBinaryOp(inner, columnName, schema)
			if err != nil {
				return nil, err
			}
		default:
			filter, err = m.translateNode(q, schema)
			if err != nil {
				return nil, err
			}
		}
		filters = append(filters, filter)
	}

	if len(filters) == 1 {
		return filters[0], nil
	}

	// Join with OR (default for field groups)
	return map[string]interface{}{
		"$or": filters,
	}, nil
}

// translateFieldGroupBinaryOp handles binary operations within field groups.
func (m *MongoDBTranslator) translateFieldGroupBinaryOp(bo *parser.BinaryOp, columnName string, schema *schema.Schema) (interface{}, error) {
	var leftFilter, rightFilter interface{}
	var err error

	// Handle left side
	switch left := bo.Left.(type) {
	case *parser.TermQuery:
		leftFilter = map[string]interface{}{
			columnName: left.Term,
		}
	case *parser.WildcardQuery:
		pattern := m.wildcardToRegex(left.Pattern)
		leftFilter = map[string]interface{}{
			columnName: map[string]interface{}{
				"$regex": pattern,
			},
		}
	case *parser.BinaryOp:
		leftFilter, err = m.translateFieldGroupBinaryOp(left, columnName, schema)
		if err != nil {
			return nil, err
		}
	default:
		leftFilter, err = m.translateNode(bo.Left, schema)
		if err != nil {
			return nil, err
		}
	}

	// Handle right side
	switch right := bo.Right.(type) {
	case *parser.TermQuery:
		rightFilter = map[string]interface{}{
			columnName: right.Term,
		}
	case *parser.WildcardQuery:
		pattern := m.wildcardToRegex(right.Pattern)
		rightFilter = map[string]interface{}{
			columnName: map[string]interface{}{
				"$regex": pattern,
			},
		}
	case *parser.BinaryOp:
		rightFilter, err = m.translateFieldGroupBinaryOp(right, columnName, schema)
		if err != nil {
			return nil, err
		}
	default:
		rightFilter, err = m.translateNode(bo.Right, schema)
		if err != nil {
			return nil, err
		}
	}

	operator := strings.ToLower(bo.Op)
	if operator == "and" {
		return map[string]interface{}{
			"$and": []interface{}{leftFilter, rightFilter},
		}, nil
	} else if operator == "or" {
		return map[string]interface{}{
			"$or": []interface{}{leftFilter, rightFilter},
		}, nil
	}

	return nil, fmt.Errorf("unsupported binary operator in field group: %s", bo.Op)
}
