#!/bin/bash

# rsearch Schema System Demo
# This script demonstrates the complete schema management workflow

set -e

BASE_URL="http://localhost:8080"

echo "================================"
echo "rsearch Schema System Demo"
echo "================================"
echo ""

# Check if server is running
echo "1. Checking server health..."
HEALTH=$(curl -s $BASE_URL/health)
echo "Response: $HEALTH"
echo ""

# Register a schema
echo "2. Registering product schema..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d @product_schema.json)
echo "Response: $REGISTER_RESPONSE" | jq '.'
echo ""

# Get the schema
echo "3. Retrieving product schema..."
GET_RESPONSE=$(curl -s $BASE_URL/api/v1/schemas/products)
echo "Response: $GET_RESPONSE" | jq '.'
echo ""

# List all schemas
echo "4. Listing all schemas..."
LIST_RESPONSE=$(curl -s $BASE_URL/api/v1/schemas)
echo "Response: $LIST_RESPONSE" | jq '.'
echo ""

# Register another schema
echo "5. Registering users schema..."
USERS_SCHEMA='{
  "name": "users",
  "fields": {
    "userId": {"type": "integer"},
    "userName": {"type": "text", "aliases": ["user", "login"]},
    "email": {"type": "text"},
    "createdAt": {"type": "datetime"}
  },
  "options": {
    "namingConvention": "snake_case",
    "strictFieldNames": false
  }
}'
USERS_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/schemas \
  -H "Content-Type: application/json" \
  -d "$USERS_SCHEMA")
echo "Response: $USERS_RESPONSE" | jq '.'
echo ""

# List all schemas again
echo "6. Listing all schemas (should have 2)..."
LIST_RESPONSE=$(curl -s $BASE_URL/api/v1/schemas)
echo "Response: $LIST_RESPONSE" | jq '.data | length'
echo ""

# Delete a schema
echo "7. Deleting users schema..."
DELETE_RESPONSE=$(curl -s -X DELETE -w "\nStatus: %{http_code}" $BASE_URL/api/v1/schemas/users)
echo "Response: $DELETE_RESPONSE"
echo ""

# Verify deletion
echo "8. Trying to get deleted schema (should return 404)..."
VERIFY_RESPONSE=$(curl -s -w "\nStatus: %{http_code}" $BASE_URL/api/v1/schemas/users)
echo "Response: $VERIFY_RESPONSE"
echo ""

# Final schema list
echo "9. Final schema list (should have 1)..."
LIST_RESPONSE=$(curl -s $BASE_URL/api/v1/schemas)
echo "Response: $LIST_RESPONSE" | jq '.data | length'
echo ""

echo "================================"
echo "Demo completed successfully!"
echo "================================"
