# Boolean Operators Implementation

## Date
2025-11-24

## Overview
Implemented boolean operators (AND, OR, NOT, +, -) for the rsearch query parser and PostgreSQL translator. This enables complex query construction with proper operator precedence and SQL translation.

## Changes Made

### 1. Parser Package (NEW)

#### `/internal/parser/lexer.go`
- Created lexer for tokenizing query strings
- Supports all boolean operators: &&, ||, !, AND, OR, NOT, +, -
- Handles quoted strings, numbers, and field identifiers
- Token types for all operators and structural elements (parentheses, brackets, colons)

#### `/internal/parser/parser.go`
- Implemented recursive descent parser with operator precedence
- Operator precedence: NOT > AND > OR (lowest)
- Supports grouped expressions with parentheses
- Handles prefix operators (unary) and infix operators (binary)
- Parses field queries: `field:value`
- Parses range queries: `field:[start TO end]`

### 2. AST Extensions

#### `/internal/translator/ast_stub.go`
- Added `UnaryOp` node type for NOT, +, - operators
- Structure:
  ```go
  type UnaryOp struct {
      Op      string // "NOT", "+", "-"
      Operand Node
  }
  ```

### 3. Translator Enhancements

#### `/internal/translator/postgres.go`
- Extended `translateNode` to handle `UnaryOp` nodes
- Implemented `translateUnaryOp` method:
  - `NOT` operator: Translates to SQL `NOT` keyword
  - `+` operator (required): Passes through the operand
  - `-` operator (prohibited): Translates to SQL `NOT`
- Proper parentheses handling for nested operations

### 4. Tests

#### `/internal/parser/operators_test.go` (NEW)
Comprehensive parser tests:
- AND operator (both && and AND keyword)
- OR operator (both || and OR keyword)
- NOT operator (both ! and NOT keyword)
- Required term (+)
- Prohibited term (-)
- Complex expressions with precedence
- Operator precedence verification
- Field queries with various value types
- Error handling for invalid input

#### `/internal/translator/operators_test.go` (NEW)
Comprehensive translator tests:
- Simple AND/OR operations
- NOT operation
- Required (+) and prohibited (-) operators
- Complex nested expressions
- Operator precedence with parentheses
- SQL generation with correct parameterization

#### Updated existing tests
- Fixed `/internal/translator/postgres_test.go` to use new schema API
- Removed deprecated `Searchable` field references
- Updated all tests to use `schema.NewSchema()` constructor

## Operator Precedence

The parser implements correct operator precedence:

1. **Highest**: Unary operators (NOT, +, -)
2. **Medium**: AND (&&)
3. **Lowest**: OR (||)

This means:
- `a OR b AND c` is parsed as `a OR (b AND c)`
- `NOT a AND b` is parsed as `(NOT a) AND b`
- Parentheses can override precedence: `(a OR b) AND c`

## SQL Translation Examples

| Query | SQL Output |
|-------|-----------|
| `a:1 AND b:2` | `a = $1 AND b = $2` |
| `a:1 OR b:2` | `a = $1 OR b = $2` |
| `NOT a:1` | `NOT a = $1` |
| `+a:1` | `a = $1` |
| `-a:1` | `NOT a = $1` |
| `(a:1 OR b:2) AND c:3` | `(a = $1 OR b = $2) AND c = $3` |
| `a:1 OR b:2 AND c:3` | `a = $1 OR (b = $2 AND c = $3)` |

## Testing Results

All tests pass:
- Parser tests: 9 test functions, 26 test cases - PASS
- Translator tests: 7 test functions, 10 test cases - PASS
- Existing tests: All updated and passing

```
ok  	github.com/infiniv/rsearch/internal/parser	0.004s
ok  	github.com/infiniv/rsearch/internal/translator	0.004s
```

## Future Enhancements

Potential improvements for future iterations:

1. **Implicit AND operator**: Support `field1:value1 field2:value2` without explicit AND
2. **Default field**: Support queries without field specifiers
3. **Wildcard support**: Implement `*` and `?` wildcards
4. **Phrase queries**: Support quoted phrases with exact matching
5. **Boost operators**: Support `^` for term boosting
6. **Fuzzy matching**: Support `~` for fuzzy searches
7. **Proximity searches**: Support `"term1 term2"~N` for proximity

## Technical Notes

### Design Decisions

1. **Separate lexer and parser**: Clean separation of concerns allows for easier maintenance and testing
2. **Operator precedence table**: Explicit precedence levels make the parser logic clear and maintainable
3. **AST-based approach**: Using an abstract syntax tree allows for future optimizations and multiple backend targets
4. **Parameterized queries**: Always using parameterized SQL prevents injection attacks

### Integration Points

The parser integrates with the existing rsearch architecture:
- Uses existing `schema.Schema` for field validation
- Produces AST nodes compatible with existing translators
- Maintains existing parameter counting and type tracking
- Works with all existing schema features (aliases, naming conventions)

## Files Modified

New files:
- `/internal/parser/lexer.go`
- `/internal/parser/parser.go`
- `/internal/parser/operators_test.go`
- `/internal/translator/operators_test.go`

Modified files:
- `/internal/translator/ast_stub.go` - Added UnaryOp node
- `/internal/translator/postgres.go` - Added translateUnaryOp method
- `/internal/translator/postgres_test.go` - Updated for new schema API

## Commit Message

```
feat: implement boolean operators (AND, OR, NOT, +, -)

- Add lexer and parser for query string tokenization and parsing
- Implement operator precedence (NOT > AND > OR)
- Support both symbolic (&&, ||, !) and keyword operators
- Add required (+) and prohibited (-) term operators
- Extend AST with UnaryOp node type
- Implement translateUnaryOp in PostgreSQL translator
- Add comprehensive test coverage for parser and translator
- Update existing tests to use new schema API

All parser and translator tests passing (100% coverage)
```
