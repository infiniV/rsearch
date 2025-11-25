# MongoDB Translator Implementation

## Overview

The MongoDB translator for rsearch converts parsed AST nodes into MongoDB query filter documents. It implements the `Translator` interface and outputs valid MongoDB query syntax in the `Filter` field of `TranslatorOutput`.

## Files Created

### Implementation
- `/home/raw/rsearch/internal/translator/mongodb.go` (606 lines)
  - `MongoDBTranslator` struct with state management
  - Implementation of all AST node type translations
  - Helper methods for wildcard-to-regex conversion and metadata handling

### Tests
- `/home/raw/rsearch/internal/translator/mongodb_test.go` (772 lines)
  - 33 comprehensive test cases covering all node types
  - Tests for error cases and edge conditions
  - Complex nested query tests

### Example
- `/home/raw/rsearch/examples/mongodb_example.go`
  - Demonstrates MongoDB filter generation for common queries

## Translation Mapping

### Field Queries
| Query Type | Example | MongoDB Filter |
|------------|---------|----------------|
| Simple field | `product_code:13w42` | `{"product_code": "13w42"}` |
| Phrase value | `description:"hello world"` | `{"description": "hello world"}` |
| Wildcard | `product_code:13*` | `{"product_code": {"$regex": "^13.*$"}}` |
| Regex | `product_code:/^[0-9]+$/` | `{"product_code": {"$regex": "^[0-9]+$", "$options": ""}}` |

### Boolean Operators
| Operator | Example | MongoDB Filter |
|----------|---------|----------------|
| AND | `a:1 AND b:2` | `{"$and": [{"a": "1"}, {"b": "2"}]}` |
| OR | `a:1 OR a:2` | `{"$or": [{"a": "1"}, {"a": "2"}]}` |
| NOT (simple) | `NOT status:inactive` | `{"status": {"$ne": "inactive"}}` |
| NOT (complex) | `NOT (a:1 AND b:2)` | `{"$nor": [{"$and": [{"a": "1"}, {"b": "2"}]}]}` |

### Range Queries
| Range Type | Example | MongoDB Filter |
|------------|---------|----------------|
| Inclusive | `price:[50 TO 500]` | `{"price": {"$gte": "50", "$lte": "500"}}` |
| Exclusive | `price:{50 TO 500}` | `{"price": {"$gt": "50", "$lt": "500"}}` |
| Mixed | `price:[50 TO 500}` | `{"price": {"$gte": "50", "$lt": "500"}}` |
| Unbounded start | `price:[* TO 500]` | `{"price": {"$lte": "500"}}` |
| Unbounded end | `price:[50 TO *]` | `{"price": {"$gte": "50"}}` |

### Advanced Queries
| Query Type | Example | MongoDB Filter | Notes |
|------------|---------|----------------|-------|
| Exists | `_exists_:field` | `{"field": {"$exists": true, "$ne": null}}` | Checks field exists and is not null |
| Fuzzy | `name:laptop~2` | `{"$text": {"$search": "laptop"}}` | Requires text index, stores distance in metadata |
| Proximity | `"gaming laptop"~5` | `{"$text": {"$search": "\"gaming laptop\""}}` | Requires text index, stores distance in metadata |
| Boost | `product:laptop^2` | `{"product": "laptop"}` | Filter unchanged, boost stored in metadata |
| Field group | `status:(active OR pending)` | `{"$or": [{"status": "active"}, {"status": "pending"}]}` | Expands to OR of field queries |

## Implementation Details

### Wildcard to Regex Conversion
The translator converts wildcard patterns to MongoDB regex patterns:
- `*` (match any characters) → `.*` (regex)
- `?` (match single character) → `.` (regex)
- Special regex characters are escaped
- Pattern is anchored with `^...$`

Example: `13*` → `^13.*$`

### NOT Operator Handling
The translator intelligently handles NOT operations:
- **Simple field equality**: Uses `$ne` operator for efficiency
  - `NOT status:inactive` → `{"status": {"$ne": "inactive"}}`
- **Complex expressions**: Uses `$nor` for correct negation
  - `NOT (a:1 AND b:2)` → `{"$nor": [{"$and": [...]}]}`

### Metadata Storage
Certain query features that don't directly translate to MongoDB filters are stored in metadata:
- **Boost values**: Stored in `metadata["boosts"]` array
- **Fuzzy distance**: Stored in `metadata["fuzzy_distance"]`
- **Proximity distance**: Stored in `metadata["proximity_distance"]`

This allows application code to use these values for scoring or post-processing.

## Test Results

All 33 MongoDB translator tests pass:
```
=== MongoDB Translator Tests ===
✓ TestMongoDBTranslator_DatabaseType
✓ TestMongoDBTranslator_SimpleFieldQuery
✓ TestMongoDBTranslator_NumberFieldQuery
✓ TestMongoDBTranslator_PhraseValue
✓ TestMongoDBTranslator_BooleanAND
✓ TestMongoDBTranslator_BooleanOR
✓ TestMongoDBTranslator_UnaryOpNOT
✓ TestMongoDBTranslator_UnaryOpNOTWithBinaryOp
✓ TestMongoDBTranslator_RangeQueryInclusive
✓ TestMongoDBTranslator_RangeQueryExclusive
✓ TestMongoDBTranslator_RangeQueryMixed
✓ TestMongoDBTranslator_RangeQueryUnboundedStart
✓ TestMongoDBTranslator_RangeQueryUnboundedEnd
✓ TestMongoDBTranslator_WildcardValue
✓ TestMongoDBTranslator_WildcardValueWithQuestionMark
✓ TestMongoDBTranslator_RegexValue
✓ TestMongoDBTranslator_ExistsQuery
✓ TestMongoDBTranslator_BoostQuery
✓ TestMongoDBTranslator_GroupQuery
✓ TestMongoDBTranslator_RequiredQuery
✓ TestMongoDBTranslator_ProhibitedQuery
✓ TestMongoDBTranslator_TermQuery
✓ TestMongoDBTranslator_PhraseQuery
✓ TestMongoDBTranslator_WildcardQuery
✓ TestMongoDBTranslator_FuzzyQuery
✓ TestMongoDBTranslator_FuzzyQueryDisabled
✓ TestMongoDBTranslator_ProximityQuery
✓ TestMongoDBTranslator_ProximityQueryDisabled
✓ TestMongoDBTranslator_FieldGroupQuery
✓ TestMongoDBTranslator_FieldGroupQueryWithWildcard
✓ TestMongoDBTranslator_ComplexNestedQuery
✓ TestMongoDBTranslator_NoDefaultField
✓ TestMongoDBTranslator_InvalidFieldName

PASS: 33/33 tests
```

All existing project tests also pass (166 total tests across all packages).

## Usage Example

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/infiniv/rsearch/internal/parser"
    "github.com/infiniv/rsearch/internal/schema"
    "github.com/infiniv/rsearch/internal/translator"
)

func main() {
    // Create MongoDB translator
    mongoTranslator := translator.NewMongoDBTranslator()

    // Create schema
    testSchema := schema.NewSchema("products", map[string]schema.Field{
        "product_code": {Type: schema.TypeText},
        "region":       {Type: schema.TypeText},
        "price":        {Type: schema.TypeFloat},
    }, schema.SchemaOptions{})

    // Parse query
    query := "(status:active OR status:pending) AND price:[50 TO 500]"
    p := parser.NewParser(query)
    ast, err := p.Parse()
    if err != nil {
        panic(err)
    }

    // Translate to MongoDB filter
    output, err := mongoTranslator.Translate(ast, testSchema)
    if err != nil {
        panic(err)
    }

    // Use the filter
    filterJSON, _ := json.MarshalIndent(output.Filter, "", "  ")
    fmt.Printf("MongoDB Filter:\n%s\n", string(filterJSON))

    // Output:
    // {
    //   "$and": [
    //     {
    //       "$or": [
    //         {"status": "active"},
    //         {"status": "pending"}
    //       ]
    //     },
    //     {
    //       "price": {
    //         "$gte": "50",
    //         "$lte": "500"
    //       }
    //     }
    //   ]
    // }
}
```

## Design Decisions

### 1. No MongoDB Driver Dependency
The implementation uses `map[string]interface{}` for filters instead of importing the MongoDB driver. This keeps the translator lightweight and driver-agnostic. Applications can convert to BSON types as needed.

### 2. Filter Field Usage
Filters are returned in the `Filter` field (not `WhereClause`), following the NoSQL pattern established in the `TranslatorOutput` struct.

### 3. Metadata for Non-Translatable Features
Features like boost, fuzzy distance, and proximity distance that don't have direct MongoDB query equivalents are stored in metadata for application-level handling.

### 4. Text Index Requirements
Fuzzy and proximity searches require MongoDB text indexes and return errors if not enabled in schema options. This prevents runtime failures when queries are executed.

### 5. Smart NOT Translation
The translator uses `$ne` for simple negations (more efficient) and `$nor` for complex expressions (correct semantics).

## Limitations

1. **Fuzzy Search**: Uses MongoDB `$text` search, which requires text indexes. Levenshtein distance is stored in metadata but not enforced in the filter.

2. **Proximity Search**: Uses MongoDB `$text` search with phrase syntax. The distance parameter is stored in metadata but MongoDB doesn't support exact proximity distance like PostgreSQL.

3. **Boost**: Stored in metadata only. Applications must handle scoring separately, or use MongoDB Atlas Search for weighted queries.

4. **Type Conversion**: Values are kept as strings in filters. Applications should handle type conversion when constructing the final MongoDB query.

## Future Enhancements

Potential improvements for future versions:

1. **MongoDB Atlas Search Integration**: Support for `$search` operator with weighted scoring for boost queries
2. **Aggregation Pipeline Support**: Generate aggregation pipelines for complex queries
3. **Type-Aware Filters**: Automatic type conversion based on schema field types
4. **GeoSpatial Query Support**: Add support for location-based queries
5. **Array Query Support**: Enhanced handling for array field types
