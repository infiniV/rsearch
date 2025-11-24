# Boost Query Implementation

**Date:** 2025-11-24
**Branch:** boost
**Status:** Complete

## Overview

Implemented query boosting functionality to allow relevance weighting in search queries. Boost syntax follows the standard Lucene/OpenSearch pattern where queries can be suffixed with `^N` to apply a relevance multiplier.

## Implementation Details

### 1. AST Node Definition

Added `BoostQuery` node to `/home/raw/rsearch/.worktrees/boost/internal/translator/ast_stub.go`:

```go
// BoostQuery represents a boosted query (query^boost).
type BoostQuery struct {
    Query Node
    Boost float64
}

func (b *BoostQuery) Type() string {
    return "boost_query"
}
```

### 2. Translator Metadata Support

Extended `TranslatorOutput` structure in `/home/raw/rsearch/.worktrees/boost/internal/translator/translator.go` to include metadata:

```go
type TranslatorOutput struct {
    // ... existing fields ...

    // Metadata contains additional information about the query
    Metadata map[string]interface{}
}
```

### 3. PostgreSQL Translator Implementation

Implemented `translateBoostQuery` in `/home/raw/rsearch/.worktrees/boost/internal/translator/postgres.go`:

- **Behavior**: PostgreSQL doesn't natively support relevance scoring like search engines (Elasticsearch)
- **Strategy**: Translate the inner query normally and store boost values in metadata
- **Metadata Structure**: Boosts are stored as an array of objects with query type and boost value

```go
// Example metadata structure:
{
    "boosts": [
        {
            "query": "field_query",
            "boost": 2.0
        }
    ]
}
```

### 4. Query Syntax Examples

The parser (in separate parser worktree) already supports these boost syntaxes:

```
term^2                    // Boost simple term by 2x
name:widget^4             // Boost field match by 4x
"blue widget"^2           // Boost phrase match by 2x
(name:foo AND status:active)^3  // Boost entire group by 3x
price:[10 TO 100]^1.5     // Boost range query by 1.5x
```

## Test Coverage

Created comprehensive test suite in `/home/raw/rsearch/.worktrees/boost/internal/translator/boost_test.go`:

1. **TestPostgresTranslator_BoostQuery_SimpleField** - Basic field boost (name:widget^2)
2. **TestPostgresTranslator_BoostQuery_HighBoost** - High boost values (^4)
3. **TestPostgresTranslator_BoostQuery_WithBinaryOp** - Boost with AND/OR operations
4. **TestPostgresTranslator_BoostQuery_MultipleBoosts** - Multiple boosted terms in one query
5. **TestPostgresTranslator_BoostQuery_NestedBoost** - Boosting complex nested queries
6. **TestPostgresTranslator_BoostQuery_RangeQuery** - Boosting range queries

All tests pass successfully:

```
PASS: TestPostgresTranslator_BoostQuery_SimpleField (0.00s)
PASS: TestPostgresTranslator_BoostQuery_HighBoost (0.00s)
PASS: TestPostgresTranslator_BoostQuery_WithBinaryOp (0.00s)
PASS: TestPostgresTranslator_BoostQuery_MultipleBoosts (0.00s)
PASS: TestPostgresTranslator_BoostQuery_NestedBoost (0.00s)
PASS: TestPostgresTranslator_BoostQuery_RangeQuery (0.00s)
```

## Design Decisions

### Why Metadata Approach for PostgreSQL?

1. **SQL Limitations**: SQL databases don't have built-in relevance scoring
2. **Translator Purity**: The translator should produce valid SQL that matches the logical query
3. **Application Flexibility**: Applications can use boost metadata to:
   - Implement custom relevance scoring
   - Apply weights in application logic
   - Log/monitor which fields users consider important
   - Pass boost information to caching layers

### Future: Elasticsearch Translator

When the Elasticsearch translator is implemented, it will properly utilize boost values:

```json
{
  "query": {
    "bool": {
      "should": [
        {
          "match": {
            "name": {
              "query": "widget",
              "boost": 2.0
            }
          }
        }
      ]
    }
  }
}
```

## Files Changed

1. `/home/raw/rsearch/.worktrees/boost/internal/translator/ast_stub.go`
   - Added `BoostQuery` AST node

2. `/home/raw/rsearch/.worktrees/boost/internal/translator/translator.go`
   - Added `Metadata` field to `TranslatorOutput`

3. `/home/raw/rsearch/.worktrees/boost/internal/translator/postgres.go`
   - Added `metadata` field to `PostgresTranslator`
   - Implemented `translateBoostQuery` method
   - Updated `Translate` to initialize and return metadata

4. `/home/raw/rsearch/.worktrees/boost/internal/translator/boost_test.go`
   - New file with comprehensive boost test coverage

## Usage Example

```go
// Parse query: name:widget^2
ast := &BoostQuery{
    Query: &FieldQuery{
        Field: "name",
        Value: "widget",
    },
    Boost: 2.0,
}

// Translate to PostgreSQL
translator := NewPostgresTranslator()
output, _ := translator.Translate(ast, schema)

// Result:
// output.WhereClause = "name = $1"
// output.Parameters = ["widget"]
// output.Metadata = {
//     "boosts": [{
//         "query": "field_query",
//         "boost": 2.0
//     }]
// }
```

## Notes

- Parser implementation (in parser worktree) already handles boost parsing
- Boost values are stored as `float64` to support fractional boosts (e.g., `^1.5`)
- Multiple boosts in a single query are preserved in metadata array
- Boost metadata includes the query type to help applications understand context

## Next Steps

1. Integrate with API handlers to return boost metadata in responses
2. Implement Elasticsearch translator that uses boost values natively
3. Add documentation for API consumers on how to use boost metadata
4. Consider adding boost value validation (e.g., must be positive)
