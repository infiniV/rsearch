# Wildcard and Regex Query Implementation

**Date**: 2025-11-24
**Author**: Claude
**Status**: Implemented

## Overview

This document describes the implementation of wildcard and regex pattern matching for rsearch queries. These features allow users to perform flexible text searches using familiar wildcard syntax (like shell globs) and powerful regular expressions.

## Features Implemented

### 1. Wildcard Queries

Wildcard queries support two special characters:

- `*` - Matches zero or more characters
- `?` - Matches exactly one character

#### Examples

| Query | PostgreSQL Translation | Description |
|-------|----------------------|-------------|
| `name:wid*` | `name LIKE 'wid%'` | Prefix match |
| `name:*get` | `name LIKE '%get'` | Suffix match |
| `name:*widget*` | `name LIKE '%widget%'` | Contains match |
| `code:wi?get` | `code LIKE 'wi_get'` | Single character wildcard |
| `name:w?d*t` | `name LIKE 'w_d%t'` | Mixed wildcards |

### 2. Regex Queries

Regex queries support full PostgreSQL regex syntax using the `~` operator.

#### Examples

| Query | PostgreSQL Translation | Description |
|-------|----------------------|-------------|
| `name:/wi[dg]get/` | `name ~ 'wi[dg]get'` | Character class |
| `email:/^[a-z]+@[a-z]+\\.com$/` | `email ~ '^[a-z]+@[a-z]+\\.com$'` | Email validation |
| `name:/(?i)widget/` | `name ~ '(?i)widget'` | Case-insensitive |

### 3. Special Character Escaping

The implementation properly escapes special characters to prevent SQL injection and ensure correct pattern matching:

#### Wildcard Escaping

- Literal `%` → `\%` (PostgreSQL LIKE wildcard)
- Literal `_` → `\_` (PostgreSQL LIKE wildcard)
- Literal `\` → `\\` (Escape character)

#### Examples

| Input Pattern | Output LIKE Pattern | Description |
|--------------|-------------------|-------------|
| `50%*` | `50\%%` | Literal percent sign |
| `test_*` | `test\_%` | Literal underscore |
| `C:\*` | `C:\\%` | Literal backslash |

## Implementation Details

### AST Node Types

Two new AST node types were added to `internal/translator/ast_stub.go`:

```go
// WildcardQuery represents a wildcard query like field:wid* or field:wi?get
type WildcardQuery struct {
    Field   string
    Pattern string
}

// RegexQuery represents a regex query like field:/pattern/
type RegexQuery struct {
    Field   string
    Pattern string
}
```

### PostgreSQL Translator

The `PostgresTranslator` was extended with two new translation methods:

#### 1. `translateWildcardQuery()`

- Validates field exists in schema
- Converts wildcard pattern to PostgreSQL LIKE pattern using `convertWildcardToLike()`
- Generates parameterized query with LIKE operator

#### 2. `translateRegexQuery()`

- Validates field exists in schema
- Uses PostgreSQL regex operator (`~`)
- Generates parameterized query

#### 3. `convertWildcardToLike()` Helper

This helper function converts wildcard patterns to PostgreSQL LIKE patterns:

```go
func convertWildcardToLike(pattern string) string {
    // * → %
    // ? → _
    // %, _, \ → escaped
}
```

### Integration with Existing Code

The new query types integrate seamlessly with existing features:

- **Binary Operations**: Wildcards and regex work with AND/OR operators
- **Schema Validation**: Field existence is validated before translation
- **Parameterized Queries**: All patterns use PostgreSQL parameters to prevent SQL injection
- **Parameter Tracking**: Proper parameter numbering across mixed query types

## Test Coverage

Comprehensive test suite added to `internal/translator/postgres_test.go`:

### Wildcard Tests (11 tests)

1. **Prefix Match** - `wid*` → `wid%`
2. **Suffix Match** - `*get` → `%get`
3. **Contains** - `*widget*` → `%widget%`
4. **Single Character** - `wi?get` → `wi_get`
5. **Mixed Wildcards** - `w?d*t` → `w_d%t`
6. **Escape Percent** - `50%*` → `50\%%`
7. **Escape Underscore** - `test_*` → `test\_%`
8. **Escape Backslash** - `C:\*` → `C:\\%`
9. **Field Not Found** - Error handling
10. **With Binary Op** - Integration with AND/OR
11. **Parameter Numbering** - Multiple wildcards in one query

### Regex Tests (5 tests)

1. **Simple Pattern** - `wi[dg]get` → `name ~ 'wi[dg]get'`
2. **Complex Pattern** - Email validation regex
3. **Case Insensitive** - `(?i)widget`
4. **Field Not Found** - Error handling
5. **With Binary Op** - Integration with AND/OR
6. **Mixed with Wildcard** - Wildcard and regex in same query

All tests verify:
- Correct SQL generation
- Parameter values and types
- Error handling for invalid fields
- Integration with binary operations

## Future Enhancements

### Parser Integration

When the parser is implemented, it should:

1. Detect `*` and `?` in term values → Create `WildcardQuery` nodes
2. Detect `/pattern/` syntax → Create `RegexQuery` nodes
3. Handle proper escaping of wildcard characters
4. Support regex delimiter escaping

### Potential Extensions

1. **Case-Insensitive Wildcards**: Add ILIKE support for case-insensitive wildcard matching
2. **Case-Insensitive Regex**: Detect case-insensitive flag and use `~*` operator
3. **NOT Regex**: Support `!~` operator for negated regex matches
4. **Wildcard Negation**: Support NOT LIKE for negated wildcard queries
5. **Performance Hints**: Add index usage hints for pattern queries

## SQL Injection Prevention

All implementations use parameterized queries:

```go
// Parameters are passed separately, never concatenated
p.params = append(p.params, likePattern)
sql := fmt.Sprintf("%s LIKE $%d", columnName, p.paramCount)
```

This ensures:
- No SQL injection vulnerabilities
- Proper escaping of user input
- Database driver handles parameter type conversion

## Examples in Context

### Simple Wildcard Query
```
Input:  name:wid*
AST:    WildcardQuery{Field: "name", Pattern: "wid*"}
SQL:    name LIKE $1
Param:  ["wid%"]
```

### Complex Mixed Query
```
Input:  name:wid* AND (region:ca OR description:/test.*/)
SQL:    name LIKE $1 AND (region = $2 OR description ~ $3)
Params: ["wid%", "ca", "test.*"]
```

## Verification

Run tests with:
```bash
cd /home/raw/rsearch/.worktrees/wildcards
source /home/raw/rsearch/.go-env
go test ./internal/translator -v -run "Wildcard|Regex"
```

All 16 new tests pass successfully.

## Conclusion

The wildcard and regex implementation provides:
- Intuitive wildcard syntax (`*` and `?`)
- Powerful regex pattern matching
- Proper special character escaping
- Full integration with existing query features
- Comprehensive test coverage
- SQL injection protection

The implementation is production-ready and awaits parser integration to provide end-to-end functionality.
