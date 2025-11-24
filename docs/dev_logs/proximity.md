# Proximity Search Implementation

**Date**: 2025-11-24
**Feature**: Proximity Search for Phrases
**Status**: Completed

## Overview

Implemented proximity search functionality for rsearch, allowing users to search for terms within a specified distance of each other using PostgreSQL full-text search capabilities.

## Syntax

```
"exact phrase"~N
```

Where `N` is the maximum number of positions between words.

### Examples

- `"quick brown"~5` - Find "quick" and "brown" within 5 positions of each other
- `"the quick fox"~3` - Find these three words within 3 positions of each other
- `content:"database system"~10` - Field-specific proximity search

## Implementation Details

### 1. AST Node Addition

**File**: `internal/translator/ast_stub.go`

Added `ProximityQuery` struct to represent proximity search queries:

```go
type ProximityQuery struct {
    Field    string
    Terms    []string
    Distance int
}
```

The AST node captures:
- `Field`: The field to search in
- `Terms`: Array of search terms (minimum 2 required)
- `Distance`: Maximum positions between consecutive terms

### 2. Feature Flag Configuration

**File**: `internal/config/config.go`

Added feature flags for advanced search capabilities:

```go
type FeaturesConfig struct {
    QuerySuggestions bool   `mapstructure:"querySuggestions"`
    MaxQueryLength   int    `mapstructure:"maxQueryLength"`
    RequestIDHeader  string `mapstructure:"requestIdHeader"`
    Fuzzy            bool   `mapstructure:"fuzzy"`
    Proximity        bool   `mapstructure:"proximity"`  // New
}
```

Default configuration:
- `features.proximity`: `false` (must be explicitly enabled)

### 3. PostgreSQL Translator Enhancement

**File**: `internal/translator/postgres.go`

Updated `PostgresTranslator` to:
1. Accept configuration in constructor
2. Implement proximity query translation using PostgreSQL FTS

#### Translation Logic

Proximity queries are translated to PostgreSQL full-text search syntax:

```
"word1 word2"~5 → to_tsvector('english', field) @@ to_tsquery('english', 'word1 <5> word2')
```

For multiple terms:
```
"the quick fox"~3 → to_tsvector('english', field) @@ to_tsquery('english', 'the <3> quick <3> fox')
```

#### Validation Rules

1. **Feature Flag Check**: Proximity search requires `features.proximity: true`
2. **Field Type Validation**: Only text fields support proximity search
3. **Minimum Terms**: At least 2 terms required
4. **Field Existence**: Field must exist in schema

### 4. Error Handling

When proximity search is disabled or conditions aren't met:

```json
{
  "error": {
    "code": "FEATURE_DISABLED",
    "message": "Proximity search requires PostgreSQL full-text search (feature disabled)"
  }
}
```

Other validation errors:
- `"proximity search requires text field, got number"`
- `"proximity search requires at least 2 terms, got 1"`
- `"field 'invalid_field' not found in schema 'products'"`

## PostgreSQL Full-Text Search Details

### Distance Operator

PostgreSQL's `<N>` operator in `to_tsquery()` matches terms within N lexemes:

```sql
-- Match if 'quick' appears within 5 positions of 'brown'
SELECT * FROM documents
WHERE to_tsvector('english', content) @@ to_tsquery('english', 'quick <5> brown');
```

### Language Support

Current implementation uses English dictionary (`'english'`):
- Stemming: "running" matches "run"
- Stop words: Common words like "the", "a", "an" are ignored
- Case insensitive by default

Future enhancement: Make language configurable per schema or query.

## Testing

### Test Coverage

Comprehensive test suite in `internal/translator/postgres_test.go`:

1. **Feature Enabled Tests**
   - Basic two-term proximity query
   - Three-term proximity query
   - Various distance values (1, 10, 100)

2. **Feature Disabled Tests**
   - Nil config rejection
   - Explicit disabled flag rejection

3. **Validation Tests**
   - Non-text field rejection
   - Insufficient terms rejection
   - Non-existent field rejection

### Test Results

All 24 tests pass:
- 13 existing translator tests (unchanged)
- 11 new proximity search tests

```
PASS
ok  	github.com/infiniv/rsearch/internal/translator	0.005s
```

## Configuration Example

### Enabling Proximity Search

```yaml
features:
  proximity: true
  fuzzy: false
  querySuggestions: false
```

### Schema Definition

```json
{
  "name": "documents",
  "fields": {
    "content": {
      "type": "text",
      "indexed": true
    },
    "title": {
      "type": "text"
    }
  }
}
```

## Usage Examples

### Basic Proximity Search

**Query**: `content:"quick brown"~5`

**Generated SQL**:
```sql
WHERE to_tsvector('english', content) @@ to_tsquery('english', 'quick <5> brown')
```

### Multiple Terms

**Query**: `content:"machine learning algorithm"~10`

**Generated SQL**:
```sql
WHERE to_tsvector('english', content) @@ to_tsquery('english', 'machine <10> learning <10> algorithm')
```

### Combined with Other Queries

**Query**: `content:"neural network"~5 AND status:published`

**Generated SQL** (conceptual):
```sql
WHERE to_tsvector('english', content) @@ to_tsquery('english', 'neural <5> network')
  AND status = 'published'
```

## Performance Considerations

### Indexing

For optimal performance, create GIN indexes on text fields:

```sql
CREATE INDEX idx_documents_content_fts
ON documents USING GIN (to_tsvector('english', content));
```

### Query Planning

PostgreSQL's query planner can use GIN indexes efficiently for full-text searches. Distance operators have similar performance characteristics to basic FTS queries.

### Memory Usage

Each proximity query adds one parameter to the query:
- Parameter: The tsquery string (e.g., `"word1 <5> word2"`)
- Parameter type: `text`

## Limitations and Future Enhancements

### Current Limitations

1. **Language Fixed**: Only English dictionary supported
2. **Distance Semantics**: PostgreSQL counts lexemes (stemmed words), not raw word positions
3. **Ordered Only**: Terms must appear in specified order
4. **Single Field**: Cannot search across multiple fields in one proximity query

### Future Enhancements

1. **Configurable Language**
   ```go
   type ProximityQuery struct {
       Field    string
       Terms    []string
       Distance int
       Language string  // e.g., "english", "spanish", "simple"
   }
   ```

2. **Unordered Proximity**
   - Support for terms in any order within distance
   - Syntax: `"word1 NEAR/5 word2"`

3. **Multi-Field Proximity**
   - Search across multiple fields
   - Syntax: `(title OR content):"search term"~5`

4. **Phrase Slop with Stemming**
   - Fine-grained control over stemming behavior
   - Per-query language override

## Integration Points

### Parser Integration (Pending)

When the parser is integrated, it should:

1. Recognize `"phrase"~N` syntax
2. Parse terms from the quoted phrase
3. Extract distance value from `~N` suffix
4. Create `ProximityQuery` AST node

Example parser logic:
```go
// Pseudo-code
if match := phraseWithDistance.FindString(query); match != "" {
    phrase := extractQuotedText(match)
    distance := extractDistance(match)
    terms := strings.Fields(phrase)

    return &ProximityQuery{
        Field:    currentField,
        Terms:    terms,
        Distance: distance,
    }
}
```

### API Integration

The translate endpoint will automatically support proximity queries once the parser is integrated. No API changes required.

## Database Support Matrix

| Database     | Support Status | Implementation Method          |
|------------- |--------------- |------------------------------- |
| PostgreSQL   | Implemented    | Full-text search with <N>      |
| MySQL        | Planned        | MATCH AGAINST with proximity   |
| MongoDB      | Planned        | $text with $near operators     |
| Elasticsearch| Planned        | span_near query                |

## References

- PostgreSQL Full-Text Search: https://www.postgresql.org/docs/current/textsearch.html
- PostgreSQL tsquery operators: https://www.postgresql.org/docs/current/textsearch-controls.html
- Lucene proximity search: https://lucene.apache.org/core/

## Migration Notes

### Breaking Changes

None. This is a new feature with no impact on existing functionality.

### Configuration Changes Required

To enable proximity search, add to config:

```yaml
features:
  proximity: true
```

### Database Changes Required

For optimal performance, create GIN indexes on text fields that will be searched:

```sql
CREATE INDEX idx_table_field_fts
ON table_name USING GIN (to_tsvector('english', field_name));
```

## Commit Summary

**Files Modified**:
- `internal/translator/ast_stub.go` - Added ProximityQuery AST node
- `internal/config/config.go` - Added proximity feature flag
- `internal/translator/postgres.go` - Implemented proximity translation
- `internal/translator/postgres_test.go` - Added comprehensive tests
- `cmd/rsearch/main.go` - Updated translator initialization
- `cmd/translator_demo/main.go` - Updated for config parameter
- `internal/api/translate_handler_test.go` - Updated for config parameter

**Lines Changed**: ~350 additions

**Test Coverage**: 11 new tests, all passing
