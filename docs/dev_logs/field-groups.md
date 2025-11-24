# Field Grouping Implementation

**Date**: 2025-11-24
**Author**: Claude (AI Assistant)
**Status**: Completed

## Overview

Implemented field grouping syntax for rsearch query language, allowing multiple values to be specified for a single field using the syntax `field:(value1 OR value2)`.

## Changes

### 1. AST Node Definition

**File**: `/internal/translator/ast_stub.go`

Added `FieldGroupQuery` AST node to represent field group queries:

```go
type FieldGroupQuery struct {
    Field   string
    Queries []Node
}
```

This node represents queries like `name:(blue OR red)` where multiple values are grouped for a single field.

### 2. Parser Implementation

**Files**:
- `/internal/parser/parser.go` (new)
- `/internal/parser/parser_test.go` (new)

Created a complete parser package that handles:

- Simple field queries: `name:blue`
- Field groups: `name:(blue OR red)`, `region:(ca OR ny OR tx)`
- Binary operations: AND, OR
- Range queries: `price:[10 TO 100]`
- Quoted values: `name:("New York" OR "Los Angeles")`
- Complex nested expressions: `name:(blue OR red) AND status:active`

**Parser Features**:
- Recursive descent parsing for expressions
- Support for parentheses grouping
- Whitespace handling
- Operator precedence (implicit through parsing order)
- Error reporting with position information

**Test Coverage**: 18 comprehensive test cases covering:
- Simple and complex field groups
- Combinations with other query types
- Error cases (empty groups, unclosed parentheses, etc.)
- Edge cases (single value in group, quoted values with spaces)

### 3. PostgreSQL Translator

**Files**:
- `/internal/translator/postgres.go` (modified)
- `/internal/translator/postgres_test.go` (modified)

Implemented `translateFieldGroupQuery` method with intelligent SQL generation:

**Optimization Strategy**:
- Multiple simple values with OR: `name:(blue OR red)` → `name IN ($1, $2)`
- Single value: `name:(blue)` → `name = $1`
- Complex expressions: Falls back to expanded `(name = $1 OR name = $2)` if needed

**Benefits**:
- More efficient SQL for common cases (IN clause vs multiple OR conditions)
- Proper parameter numbering and type tracking
- Seamless integration with existing query translation

**Test Coverage**: 10 new test cases for field groups:
- Simple OR groups with IN clause generation
- Multiple values (3+ items)
- Single value optimization
- Integration with AND operations
- Multiple field groups in one query
- Different data types (text, integer)
- Error handling (field not found, empty groups)
- Complex nested queries

### 4. Schema Test Updates

**Files**:
- `/internal/translator/postgres_test.go`
- `/internal/api/translate_handler_test.go`
- `/cmd/translator_demo/main.go`

Updated all test schemas to use the current `schema.Field` structure:
- Changed from old `map[string]*schema.Field` with `Name` and `Searchable` fields
- Updated to `map[string]schema.Field` with `Type` and `Indexed` fields
- Used `schema.NewSchema()` constructor for proper initialization

## Examples

### Basic Field Group
```
Input:  name:(blue OR red)
SQL:    name IN ($1, $2)
Params: ["blue", "red"]
```

### Multiple Values
```
Input:  region:(ca OR ny OR tx)
SQL:    region IN ($1, $2, $3)
Params: ["ca", "ny", "tx"]
```

### Combined with Other Queries
```
Input:  name:(blue OR red) AND status:active
SQL:    name IN ($1, $2) AND status = $3
Params: ["blue", "red", "active"]
```

### Complex Nested Expression
```
Input:  (name:(blue OR red) AND region:ca) OR status:active
SQL:    (name IN ($1, $2) AND region = $3) OR status = $4
Params: ["blue", "red", "ca", "active"]
```

## Test Results

All tests passing:
```
ok  github.com/infiniv/rsearch/internal/parser      0.006s (18 tests)
ok  github.com/infiniv/rsearch/internal/translator  0.006s (25 tests)
ok  github.com/infiniv/rsearch/internal/api         0.009s (all tests)
```

## Architecture Notes

### Parser Design
- **Recursive descent**: Natural fit for nested expressions
- **Single-pass**: Efficient parsing without backtracking
- **Error recovery**: Clear error messages with position information
- **Extensibility**: Easy to add new syntax elements

### Translator Optimization
The field group translator includes intelligent optimization:

1. **Detection**: Checks if all queries in group are simple FieldQuery nodes
2. **IN clause generation**: For 2+ simple values, generates efficient `IN ($1, $2, ...)`
3. **Fallback**: Complex expressions fall back to expanded OR clauses
4. **Type safety**: Preserves parameter types for each value

### Integration Points
- Parser produces AST nodes compatible with existing translator infrastructure
- No changes required to schema validation or registry systems
- Backward compatible with existing query patterns
- Ready for HTTP API integration (parser can be plugged into translate handler)

## Future Enhancements

1. **Parser improvements**:
   - Support for NOT operator within field groups
   - Wildcard values: `name:(blue* OR *red)`
   - Numeric range shortcuts: `quantity:(10..20)`

2. **Translator optimizations**:
   - AND groups could optimize to array contains for PostgreSQL
   - Consider JSONB operators for nested field groups
   - Generate query execution hints for large IN clauses

3. **API integration**:
   - Wire up parser to translate endpoint
   - Add query validation before parsing
   - Implement query rewriting/normalization

## References

- **Design Doc**: `/docs/plans/2024-11-24-rsearch-design.md`
- **Parser Package**: `/internal/parser/`
- **Translator Updates**: `/internal/translator/postgres.go`
- **Test Files**:
  - `/internal/parser/parser_test.go`
  - `/internal/translator/postgres_test.go`
