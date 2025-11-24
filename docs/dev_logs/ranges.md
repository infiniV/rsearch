# Range Query Implementation

**Date:** 2025-11-24
**Branch:** ranges
**Worktree:** /home/raw/rsearch/.worktrees/ranges

## Overview

Implemented comprehensive range query support for rsearch, including parsing and PostgreSQL translation. Range queries support both bracket and brace notation for inclusive/exclusive bounds, as well as comparison operators.

## Implementation Summary

### Parser Package Integration

Copied the complete parser package from the parser worktree to the ranges worktree. The parser already had full range query support implemented:

**Location:** `/home/raw/rsearch/.worktrees/ranges/internal/parser/`

**Files:**
- `ast.go` - AST node definitions including `RangeQuery`
- `parser.go` - Parsing logic for range queries
- `lexer.go` - Tokenization support for range operators
- `errors.go` - Parse error handling
- `escape.go` - String escaping utilities

### Range Query Syntax Supported

#### Bracket Notation (Inclusive)
```
field:[start TO end]    # Both sides inclusive
age:[18 TO 65]          # SQL: age BETWEEN $1 AND $2
```

#### Brace Notation (Exclusive)
```
field:{start TO end}    # Both sides exclusive
price:{100 TO 1000}     # SQL: price > $1 AND price < $2
```

#### Mixed Notation
```
field:[start TO end}    # Inclusive start, exclusive end
score:[50 TO 100}       # SQL: score >= $1 AND score < $2

field:{start TO end]    # Exclusive start, inclusive end
rating:{0 TO 5]         # SQL: rating > $1 AND rating <= $2
```

#### Comparison Operators
```
field:>=value           # Greater than or equal
age:>=18                # SQL: age >= $1

field:>value            # Greater than
price:>100              # SQL: price > $1

field:<=value           # Less than or equal
age:<=65                # SQL: age <= $1

field:<value            # Less than
score:<100              # SQL: score < $1
```

#### Unbounded Ranges
```
field:[* TO value]      # Open start
age:[* TO 18]           # SQL: age <= $1

field:[value TO *]      # Open end
price:[100 TO *]        # SQL: price >= $1
```

### Translator Updates

**Location:** `/home/raw/rsearch/.worktrees/ranges/internal/translator/`

**Modified Files:**
- `translator.go` - Updated interface to use `parser.Node`
- `postgres.go` - Enhanced `translateRangeQuery()` method

**Key Features:**
- Supports all range notations (inclusive, exclusive, mixed)
- Handles unbounded ranges with wildcard values
- Generates optimized SQL with BETWEEN for fully inclusive ranges
- Uses comparison operators for exclusive and mixed ranges
- Proper parameter binding and type tracking
- Field name to column name mapping via schema

### Translation Examples

```
Query: age:[18 TO 65]
SQL:   age BETWEEN $1 AND $2
Params: ["18", "65"]

Query: price:{100 TO 1000}
SQL:   price > $1 AND price < $2
Params: ["100", "1000"]

Query: score:[50 TO 100}
SQL:   score >= $1 AND score < $2
Params: ["50", "100"]

Query: age:>=18
SQL:   age >= $1
Params: ["18"]

Query: price:[100 TO *]
SQL:   price >= $1
Params: ["100"]

Query: age:[* TO 18]
SQL:   age <= $1
Params: ["18"]
```

## Test Coverage

### Parser Tests

**File:** `/home/raw/rsearch/.worktrees/ranges/internal/parser/range_test.go`

**Test Suites:**
1. `TestRangeQueryParsing` - 22 test cases covering:
   - Inclusive/exclusive/mixed bracket notation
   - Comparison operators (>=, >, <=, <)
   - Date ranges
   - Decimal numbers
   - Negative numbers (quoted)
   - String ranges (alphabetical)
   - Quoted string ranges
   - Unbounded ranges (open start/end)
   - Large numbers
   - Scientific notation
   - Timestamps with microseconds
   - IP addresses (quoted)
   - Zero value ranges
   - Single character ranges

2. `TestRangeQueryInComplexQueries` - 4 test cases:
   - Range AND field query
   - Range OR range query
   - Multiple range queries with AND
   - Range with comparison operator combined

3. `TestRangeQueryEdgeCases` - 4 test cases:
   - Same start and end values
   - Whitespace handling
   - Mixed quotes in range values

**Total Parser Tests:** 30 test cases
**Status:** All tests passing

### Translator Tests

**File:** `/home/raw/rsearch/.worktrees/ranges/internal/translator/range_test.go`

**Test Suites:**
1. `TestRangeQueryTranslation` - 20 test cases covering:
   - Inclusive both sides (BETWEEN)
   - Exclusive both sides
   - Mixed inclusiveness
   - Comparison operators
   - Date ranges
   - Negative numbers
   - String ranges
   - Unbounded ranges
   - Zero values
   - Decimal values
   - Large numbers
   - Same start and end

2. `TestRangeQueryWithBooleanOperators` - 6 test cases:
   - Range AND field query
   - Range OR range
   - Multiple ranges with AND
   - Comparison with field query
   - Mixed range notations with OR
   - Exclusive and inclusive ranges combined

3. `TestRangeQueryParameterTypes` - 5 test cases:
   - Integer range type checking
   - Float range type checking
   - String range type checking
   - Comparison operator type checking

4. `TestRangeQueryFieldMapping` - 4 test cases:
   - CamelCase to snake_case conversion
   - Column name mapping
   - Timestamp field mapping
   - Comparison with field mapping

5. `TestRangeQueryErrors` - 2 test cases:
   - Non-existent field error handling
   - Valid query validation

**Total Translator Tests:** 37 test cases
**Status:** All tests passing

## Test Results

```bash
$ go test ./internal/parser/... -v -run TestRangeQuery
PASS
ok      github.com/infiniv/rsearch/internal/parser     0.004s

$ go test ./internal/translator/... -v -run TestRange
PASS
ok      github.com/infiniv/rsearch/internal/translator 0.006s
```

**Summary:**
- Total test cases: 67+
- All range query tests passing
- Coverage includes parsing, translation, error handling, and edge cases

## Files Changed

### New Files
- `/home/raw/rsearch/.worktrees/ranges/internal/parser/` (entire package copied)
  - `ast.go`
  - `parser.go`
  - `lexer.go`
  - `errors.go`
  - `escape.go`
  - `parser_test.go`
  - `lexer_test.go`
  - `range_test.go` (new comprehensive tests)

- `/home/raw/rsearch/.worktrees/ranges/internal/translator/range_test.go` (new)

### Modified Files
- `/home/raw/rsearch/.worktrees/ranges/internal/translator/translator.go`
  - Updated `Translator` interface to use `parser.Node`
  - Added import for parser package

- `/home/raw/rsearch/.worktrees/ranges/internal/translator/postgres.go`
  - Updated to use `parser.Node`, `parser.FieldQuery`, `parser.BinaryOp`, `parser.RangeQuery`
  - Enhanced `translateRangeQuery()` to handle:
    - Unbounded ranges (wildcard values)
    - All combination of inclusive/exclusive bounds
    - Proper SQL generation with BETWEEN or comparison operators
    - Parameter and type tracking

- `/home/raw/rsearch/.worktrees/ranges/internal/translator/translator_test.go`
  - Updated to use `parser.Node` in mock translator

- `/home/raw/rsearch/.worktrees/ranges/internal/translator/postgres_test.go`
  - Updated to use correct schema structure
  - Updated to use `parser` package types

### Deleted Files
- `/home/raw/rsearch/.worktrees/ranges/internal/translator/ast_stub.go` (replaced by full parser)

## Technical Notes

### Range Query AST Structure

```go
type RangeQuery struct {
    Field          string      // Field name
    Start          ValueNode   // Start value (can be wildcard "*")
    End            ValueNode   // End value (can be wildcard "*")
    InclusiveStart bool        // true for [, false for {
    InclusiveEnd   bool        // true for ], false for }
    Pos            Position    // Position in source
}
```

### Value Node Types

Range queries use the following value node types:
- `TermValue` - Simple unquoted terms
- `NumberValue` - Numeric values
- `PhraseValue` - Quoted strings
- `WildcardValue` - Wildcard patterns (for unbounded ranges)

### SQL Translation Logic

1. **Both Inclusive** → `BETWEEN $1 AND $2`
2. **Both Exclusive** → `> $1 AND < $2`
3. **Mixed** → `>= $1 AND < $2` (or appropriate combination)
4. **Unbounded Start** → `<= $1` or `< $1`
5. **Unbounded End** → `>= $1` or `> $1`

### Schema Integration

Range queries properly integrate with the schema system:
- Field name resolution and validation
- Column name mapping (field → database column)
- Type information tracking for parameters
- Support for schema naming conventions

## Edge Cases Handled

1. **Negative Numbers** - Must be quoted: `temperature:["-10" TO "40"]`
2. **Timestamps with Colons** - Must be quoted: `timestamp:["2024-01-01T00:00:00" TO "2024-12-31T23:59:59"]`
3. **IP Addresses** - Must be quoted: `ip:["192.168.0.1" TO "192.168.0.255"]`
4. **Same Start and End** - Valid for both inclusive and exclusive ranges
5. **Whitespace** - Properly handled in range expressions
6. **Wildcard Bounds** - Correctly translate to single-sided comparisons

## Future Enhancements

1. Support for date math in ranges (e.g., `[now-7d TO now]`)
2. Additional database backends (MySQL, MongoDB, Elasticsearch)
3. Range query optimization hints
4. Range validation against field types
5. Support for special values like `NULL`, `INFINITY`

## Commit Message

```
Implement range query parsing and translation

Add comprehensive range query support including:
- Bracket notation for inclusive bounds: field:[start TO end]
- Brace notation for exclusive bounds: field:{start TO end}
- Mixed notation: field:[start TO end}
- Comparison operators: >=, >, <=, <
- Unbounded ranges with wildcards: field:[* TO value]

Parser:
- Copy complete parser package with full range query support
- Support for all range query syntax variants
- 30+ parser test cases covering edge cases

Translator:
- Enhanced PostgreSQL translator for range queries
- Optimized SQL with BETWEEN for inclusive ranges
- Comparison operators for exclusive/mixed ranges
- Wildcard handling for unbounded ranges
- 37+ translator test cases

All tests passing. Ready for integration.

Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

## References

- OpenSearch Query String Syntax
- PostgreSQL BETWEEN operator
- SQL comparison operators
- Schema field resolution system
