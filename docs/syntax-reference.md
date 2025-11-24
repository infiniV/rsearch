# rsearch Query Syntax Reference

*Auto-generated from test suite - Last updated: 2025-11-25*

This reference documents all supported query syntax patterns in rsearch, with examples showing the OpenSearch query and the resulting PostgreSQL translation.

## Table of Contents

- [Boolean Operators](#boolean-operators)
- [Boost Queries](#boost-queries)
- [Complex Queries](#complex-queries)
- [Exists Queries](#exists-queries)
- [Field Queries](#field-queries)
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

