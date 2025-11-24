# Fuzzy Search Implementation

**Date:** 2025-11-24
**Branch:** fuzzy
**Status:** Complete

## Overview

Implemented fuzzy search support for rsearch, enabling approximate string matching with configurable edit distance using PostgreSQL's Levenshtein distance function (pg_trgm extension).

## Changes Made

### 1. AST Node Definition

Added `FuzzyQuery` node to `/internal/translator/ast_stub.go`:

```go
type FuzzyQuery struct {
    Field    string
    Term     string
    Distance int // Edit distance (default 2 if not specified)
}
```

This node represents fuzzy search queries in the format:
- `field:term~` - auto fuzzy with distance 2
- `field:term~1` - max edit distance 1

### 2. Schema Feature Flags

Added `EnabledFeatures` struct to `/internal/schema/schema.go`:

```go
type EnabledFeatures struct {
    Fuzzy     bool `json:"fuzzy"`     // Enable fuzzy search (requires pg_trgm for PostgreSQL)
    Proximity bool `json:"proximity"` // Enable proximity search (requires full-text search)
    Regex     bool `json:"regex"`     // Enable regex search
}
```

Updated `SchemaOptions` to include the feature flags:

```go
type SchemaOptions struct {
    NamingConvention string           `json:"namingConvention"`
    StrictOperators  bool             `json:"strictOperators"`
    StrictFieldNames bool             `json:"strictFieldNames"`
    DefaultField     string           `json:"defaultField"`
    EnabledFeatures  EnabledFeatures  `json:"enabledFeatures"`
}
```

### 3. PostgreSQL Translator

Added `translateFuzzyQuery` method to `/internal/translator/postgres.go`:

```go
func (p *PostgresTranslator) translateFuzzyQuery(fq *FuzzyQuery, schema *schema.Schema) (string, error) {
    // Check if fuzzy search is enabled
    if !schema.Options.EnabledFeatures.Fuzzy {
        return "", fmt.Errorf("fuzzy search requires pg_trgm extension...")
    }

    // Validate field exists
    columnName, field, err := schema.ResolveField(fq.Field)
    if err != nil {
        return "", fmt.Errorf("field %s not found in schema %s", fq.Field, schema.Name)
    }

    // Generate SQL: levenshtein(field, $1) <= $2
    p.paramCount++
    p.params = append(p.params, fq.Term)
    p.paramTypes = append(p.paramTypes, string(field.Type))
    termParam := p.paramCount

    p.paramCount++
    p.params = append(p.params, fq.Distance)
    p.paramTypes = append(p.paramTypes, "integer")
    distanceParam := p.paramCount

    return fmt.Sprintf("levenshtein(%s, $%d) <= $%d", columnName, termParam, distanceParam), nil
}
```

## PostgreSQL Translation

### Examples

**Auto fuzzy (distance 2):**
```
Query: name:widget~
SQL:   levenshtein(name, $1) <= $2
Params: ["widget", 2]
```

**Custom distance:**
```
Query: description:fuzzy~1
SQL:   levenshtein(description, $1) <= $2
Params: ["fuzzy", 1]
```

**Combined with other queries:**
```
Query: name:widget~ AND region:ca
SQL:   levenshtein(name, $1) <= $2 AND region = $3
Params: ["widget", 2, "ca"]
```

**With naming conventions:**
```
Schema: namingConvention: "snake_case"
Query: productName:widget~
SQL:   levenshtein(product_name, $1) <= $2
Params: ["widget", 2]
```

## Feature Flag Behavior

### Enabled (fuzzy: true)
```json
{
  "options": {
    "enabledFeatures": {
      "fuzzy": true
    }
  }
}
```

Fuzzy queries translate to Levenshtein distance SQL.

### Disabled (fuzzy: false, default)
```json
{
  "options": {
    "enabledFeatures": {
      "fuzzy": false
    }
  }
}
```

Fuzzy queries return error:
```
"fuzzy search requires pg_trgm extension. Enable in schema or use wildcards instead: field:*"
```

## Tests Added

Added comprehensive test coverage in `/internal/translator/postgres_test.go`:

1. **TestPostgresTranslator_FuzzyQuery_AutoDistance** - Tests default distance of 2
2. **TestPostgresTranslator_FuzzyQuery_CustomDistance** - Tests custom distance (1)
3. **TestPostgresTranslator_FuzzyQuery_FeatureDisabled** - Tests error when feature is disabled
4. **TestPostgresTranslator_FuzzyQuery_InvalidField** - Tests error for non-existent fields
5. **TestPostgresTranslator_FuzzyQuery_WithSnakeCase** - Tests with naming convention transformation
6. **TestPostgresTranslator_FuzzyQuery_CombinedWithOtherQueries** - Tests fuzzy + exact match
7. **TestPostgresTranslator_FuzzyQuery_MultipleFuzzyQueries** - Tests multiple fuzzy queries with OR

All tests pass successfully.

## PostgreSQL Setup

To use fuzzy search, the PostgreSQL database must have the pg_trgm extension installed:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

The Levenshtein function is provided by the pg_trgm extension and calculates the edit distance between two strings.

## Parser Integration

The parser (to be implemented in future phase) should:

1. Detect tilde suffix: `term~` or `term~N`
2. Parse the term and optional distance
3. Create FuzzyQuery AST node:
   - Default distance: 2 (if no number specified)
   - Custom distance: parse the number after tilde

Parser examples:
- `name:widget~` → `FuzzyQuery{Field: "name", Term: "widget", Distance: 2}`
- `name:gadget~1` → `FuzzyQuery{Field: "name", Term: "gadget", Distance: 1}`

## Alternative Approaches Considered

### 1. similarity() function
PostgreSQL's similarity() function uses trigram similarity (0.0 to 1.0):
```sql
similarity(field, 'term') > 0.3
```

**Not chosen because:**
- Threshold value (0.3) is arbitrary and may not work for all use cases
- Less intuitive than edit distance
- Harder to tune per-query

### 2. % operator (trigram matching)
```sql
field % 'term'
```

**Not chosen because:**
- Fixed threshold, cannot be adjusted per-query
- Less flexible than Levenshtein distance

### Chosen: levenshtein()
- Clear, intuitive metric (number of edits)
- Configurable per-query (distance parameter)
- Standard definition used across many systems
- Maps directly to OpenSearch/Elasticsearch fuzzy search semantics

## Usage Example

```json
{
  "schema": {
    "name": "products",
    "fields": {
      "name": {"type": "text"},
      "description": {"type": "text"}
    },
    "options": {
      "enabledFeatures": {
        "fuzzy": true
      }
    }
  }
}
```

Query:
```
POST /api/v1/translate
{
  "schema": "products",
  "database": "postgres",
  "query": "name:widget~ OR description:gadget~1"
}
```

Response:
```json
{
  "type": "sql",
  "whereClause": "levenshtein(name, $1) <= $2 OR levenshtein(description, $3) <= $4",
  "parameters": ["widget", 2, "gadget", 1],
  "parameterTypes": ["text", "integer", "text", "integer"]
}
```

## Files Modified

- `/internal/translator/ast_stub.go` - Added FuzzyQuery node
- `/internal/schema/schema.go` - Added EnabledFeatures struct
- `/internal/translator/postgres.go` - Added fuzzy query translation
- `/internal/translator/postgres_test.go` - Added comprehensive tests
- `/cmd/translator_demo/main.go` - Updated to use new schema API
- `/internal/api/translate_handler_test.go` - Updated tests to use new schema API

## Migration Notes

The schema structure was updated during this implementation:
- Old: `map[string]*schema.Field` with `Name` and `Searchable` fields
- New: `map[string]schema.Field` (no pointers) with simplified structure
- Use `schema.NewSchema()` constructor for proper initialization

All existing tests were updated to use the new schema API.

## Performance Considerations

Levenshtein distance calculation can be expensive for:
- Very long strings
- High edit distances
- Large result sets

Recommendations:
1. Use appropriate indexes on searchable fields
2. Limit fuzzy search to shorter text fields
3. Consider adding LIMIT clauses to queries
4. Monitor query performance in production

## Security

- Field names are validated against the schema (whitelisting)
- All values are parameterized (SQL injection prevention)
- Feature must be explicitly enabled (security by default)
- Clear error messages guide users to alternatives

## Next Steps

1. **Parser implementation** - Add fuzzy syntax parsing (`term~` and `term~N`)
2. **Documentation** - Update user-facing API docs with fuzzy examples
3. **Alternative implementations** - Add fuzzy support for other database translators (MySQL, MongoDB)
4. **Performance tuning** - Add query analysis and optimization recommendations

## Conclusion

Fuzzy search implementation is complete and fully tested. The feature is disabled by default for security and performance, requiring explicit opt-in via schema configuration. The implementation follows rsearch design principles of parameterized queries, clear error messages, and database-specific feature flags.
