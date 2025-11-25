# rsearch Query Syntax Reference

*Auto-generated from test suite - Last updated: 2025-11-25*

This reference documents all supported query syntax patterns in rsearch, with examples showing the OpenSearch query and the resulting PostgreSQL translation.

## Table of Contents

- [Boolean Operators](#boolean-operators)
- [Boost Queries](#boost-queries)
- [Complex Queries](#complex-queries)
- [Exists Queries](#exists-queries)
- [Field Queries](#field-queries)
- [Fuzzy Search](#fuzzy-search)
- [Grouping](#grouping)
- [Proximity Search](#proximity-search)
- [Range Queries](#range-queries)
- [Regex](#regex)
- [Wildcards](#wildcards)

---

## Boolean Operators

### AND with two fields

**Query:**
```
productCode:13w42 AND region:ca
```

**PostgreSQL Translation:**
```sql
product_code = $1 AND region = $2
```

**Parameters:**
```json
[
  "13w42",
  "ca"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### OR with two fields

**Query:**
```
region:ca OR region:ny
```

**PostgreSQL Translation:**
```sql
region = $1 OR region = $2
```

**Parameters:**
```json
[
  "ca",
  "ny"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### NOT operator

**Query:**
```
region:ca AND NOT status:discontinued
```

**PostgreSQL Translation:**
```sql
region = $1 AND NOT status = $2
```

**Parameters:**
```json
[
  "ca",
  "discontinued"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### Symbol && operator

**Query:**
```
productCode:13w42 && region:ca
```

**PostgreSQL Translation:**
```sql
product_code = $1 AND region = $2
```

**Parameters:**
```json
[
  "13w42",
  "ca"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### Symbol || operator

**Query:**
```
region:ca || region:ny
```

**PostgreSQL Translation:**
```sql
region = $1 OR region = $2
```

**Parameters:**
```json
[
  "ca",
  "ny"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### Required operator (+)

**Query:**
```
+name:widget AND region:ca
```

**PostgreSQL Translation:**
```sql
name = $1 AND region = $2
```

**Parameters:**
```json
[
  "widget",
  "ca"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

### Prohibited operator (-)

**Query:**
```
region:ca AND -status:discontinued
```

**PostgreSQL Translation:**
```sql
region = $1 AND NOT status = $2
```

**Parameters:**
```json
[
  "ca",
  "discontinued"
]
```

**Parameter Types:**
```json
[
  "text",
  "text"
]
```

---

## Boost Queries

### Boost field query

**Query:**
```
name:laptop^2
```

**PostgreSQL Translation:**
```sql
name = $1
```

**Parameters:**
```json
[
  "laptop"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

**Metadata:**
```json
{
  "boosts": [
    {
      "boost": 2,
      "query": "field_query"
    }
  ]
}
```

---

## Complex Queries

### Nested boolean with ranges

**Query:**
```
(productCode:13w42 AND region:ca) OR rodLength:[50 TO 500]
```

**PostgreSQL Translation:**
```sql
(product_code = $1 AND region = $2) OR rod_length BETWEEN $3 AND $4
```

**Parameters:**
```json
[
  "13w42",
  "ca",
  "50",
  "500"
]
```

**Parameter Types:**
```json
[
  "text",
  "text",
  "integer",
  "integer"
]
```

---

### Multiple conditions with wildcards

**Query:**
```
name:widget* AND price:>=50 AND region:ca
```

**PostgreSQL Translation:**
```sql
name LIKE $1 AND price >= $2 AND region = $3
```

**Parameters:**
```json
[
  "widget%",
  "50",
  "ca"
]
```

**Parameter Types:**
```json
[
  "text",
  "float",
  "text"
]
```

---

### Range wildcard and exists combined

**Query:**
```
productCode:13w* AND price:[10 TO 500] AND _exists_:description
```

**PostgreSQL Translation:**
```sql
product_code LIKE $1 AND price BETWEEN $2 AND $3 AND description IS NOT NULL
```

**Parameters:**
```json
[
  "13w%",
  "10",
  "500"
]
```

**Parameter Types:**
```json
[
  "text",
  "float",
  "float"
]
```

---

## Exists Queries

### Field exists

**Query:**
```
_exists_:description
```

**PostgreSQL Translation:**
```sql
description IS NOT NULL
```

---

### Field does not exist

**Query:**
```
NOT _exists_:description
```

**PostgreSQL Translation:**
```sql
NOT description IS NOT NULL
```

---

## Field Queries

### Simple field match

**Query:**
```
productCode:13w42
```

**PostgreSQL Translation:**
```sql
product_code = $1
```

**Parameters:**
```json
[
  "13w42"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

### Numeric field match

**Query:**
```
rodLength:150
```

**PostgreSQL Translation:**
```sql
rod_length = $1
```

**Parameters:**
```json
[
  "150"
]
```

**Parameter Types:**
```json
[
  "integer"
]
```

---

### Phrase match with quotes

**Query:**
```
name:"Widget Pro"
```

**PostgreSQL Translation:**
```sql
name = $1
```

**Parameters:**
```json
[
  "Widget Pro"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

## Fuzzy Search

### Fuzzy match with edit distance

**Query:**
```
title:widget~2
```

**PostgreSQL Translation:**
```sql
levenshtein(title, $1) <= $2
```

**Parameters:**
```json
[
  "widget",
  2
]
```

**Parameter Types:**
```json
[
  "text",
  "integer"
]
```

---

## Grouping

### Parenthesized OR with AND

**Query:**
```
(region:ca OR region:ny) AND price:<150
```

**PostgreSQL Translation:**
```sql
(region = $1 OR region = $2) AND price < $3
```

**Parameters:**
```json
[
  "ca",
  "ny",
  "150"
]
```

**Parameter Types:**
```json
[
  "text",
  "text",
  "float"
]
```

---

### Nested groups

**Query:**
```
(name:Widget AND price:[50 TO 200]) OR (region:ca AND rodLength:>=100)
```

**PostgreSQL Translation:**
```sql
((name = $1 AND rod_length BETWEEN $2 AND $3)) OR ((region = $4 AND rod_length >= $5))
```

**Parameters:**
```json
[
  "Widget",
  "50",
  "200",
  "ca",
  "100"
]
```

**Parameter Types:**
```json
[
  "text",
  "integer",
  "integer",
  "text",
  "integer"
]
```

---

### Field group with multiple values

**Query:**
```
region:(ca OR ny OR tx)
```

**PostgreSQL Translation:**
```sql
(region = $1 OR region = $2 OR region = $3)
```

**Parameters:**
```json
[
  "ca",
  "ny",
  "tx"
]
```

**Parameter Types:**
```json
[
  "text",
  "text",
  "text"
]
```

---

## Proximity Search

### Proximity search within distance

**Query:**
```
content:"quick brown fox"~5
```

**PostgreSQL Translation:**
```sql
to_tsvector('english', content) @@ phraseto_tsquery('english', $1)
```

**Parameters:**
```json
[
  "quick brown fox"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

## Range Queries

### Inclusive range with brackets

**Query:**
```
rodLength:[50 TO 500]
```

**PostgreSQL Translation:**
```sql
rod_length BETWEEN $1 AND $2
```

**Parameters:**
```json
[
  "50",
  "500"
]
```

**Parameter Types:**
```json
[
  "integer",
  "integer"
]
```

---

### Exclusive range with braces

**Query:**
```
price:{10 TO 20}
```

**PostgreSQL Translation:**
```sql
price > $1 AND price < $2
```

**Parameters:**
```json
[
  "10",
  "20"
]
```

**Parameter Types:**
```json
[
  "float",
  "float"
]
```

---

### Mixed range inclusive start exclusive end

**Query:**
```
price:[50 TO 100}
```

**PostgreSQL Translation:**
```sql
price >= $1 AND price < $2
```

**Parameters:**
```json
[
  "50",
  "100"
]
```

**Parameter Types:**
```json
[
  "float",
  "float"
]
```

---

### Greater than or equal comparison

**Query:**
```
price:>=100
```

**PostgreSQL Translation:**
```sql
price >= $1
```

**Parameters:**
```json
[
  "100"
]
```

**Parameter Types:**
```json
[
  "float"
]
```

---

### Less than comparison

**Query:**
```
rodLength:<200
```

**PostgreSQL Translation:**
```sql
rod_length < $1
```

**Parameters:**
```json
[
  "200"
]
```

**Parameter Types:**
```json
[
  "integer"
]
```

---

### Unbounded range with wildcard end

**Query:**
```
price:[100 TO *]
```

**PostgreSQL Translation:**
```sql
price >= $1
```

**Parameters:**
```json
[
  "100"
]
```

**Parameter Types:**
```json
[
  "float"
]
```

---

## Regex

### Regex pattern

**Query:**
```
name:/wi[dg]get/
```

**PostgreSQL Translation:**
```sql
name ~ $1
```

**Parameters:**
```json
[
  "wi[dg]get"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

## Wildcards

### Wildcard suffix

**Query:**
```
name:widget*
```

**PostgreSQL Translation:**
```sql
name LIKE $1
```

**Parameters:**
```json
[
  "widget%"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

### Wildcard prefix

**Query:**
```
name:*widget
```

**PostgreSQL Translation:**
```sql
name LIKE $1
```

**Parameters:**
```json
[
  "%widget"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

### Single character wildcard

**Query:**
```
productCode:13w4?
```

**PostgreSQL Translation:**
```sql
product_code LIKE $1
```

**Parameters:**
```json
[
  "13w4_"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

### Contains wildcard

**Query:**
```
name:*idg*
```

**PostgreSQL Translation:**
```sql
name LIKE $1
```

**Parameters:**
```json
[
  "%idg%"
]
```

**Parameter Types:**
```json
[
  "text"
]
```

---

## Additional Information

### Operator Precedence

From highest to lowest:
1. Boost (`^`)
2. Field queries (`:`)
3. Required/Prohibited (`+`, `-`)
4. NOT (`NOT`, `!`)
5. AND (`AND`, `&&`)
6. OR (`OR`, `||`)

### Operator Normalization

The parser normalizes operators to keyword form:
- `&&` → `AND`
- `||` → `OR`
- `!` → `NOT`

### Parameter Safety

All queries use parameterized statements with numbered placeholders (`$1`, `$2`, etc.) to prevent SQL injection attacks. Parameters are typed according to the schema field definitions.

### Schema Configuration

Field names are resolved according to the schema's naming convention (e.g., `snake_case`). Queries use the friendly field names (e.g., `productCode`) which are automatically mapped to database columns (e.g., `product_code`).

