// rsearch Integration Example - Node.js/JavaScript
// Demonstrates how to use rsearch with PostgreSQL in a Node.js application

const { Client } = require('pg');

const RSEARCH_URL = process.env.RSEARCH_URL || 'http://localhost:8080';

// PostgreSQL connection
const db = new Client({
  host: 'localhost',
  port: 5432,
  database: 'rsearch_test',
  user: 'rsearch',
  password: 'rsearch123'
});

/**
 * Translate user query using rsearch and execute against PostgreSQL
 * @param {string} userQuery - OpenSearch-style query from user
 * @returns {Promise<Array>} - Query results
 */
async function searchProducts(userQuery) {
  try {
    // 1. Call rsearch to translate the query
    const response = await fetch(`${RSEARCH_URL}/api/v1/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        schema: 'products',
        database: 'postgres',
        query: userQuery
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`rsearch error: ${error.error}`);
    }

    const translation = await response.json();

    console.log('Query:', userQuery);
    console.log('SQL:', translation.whereClause);
    console.log('Params:', translation.parameters);
    console.log('---');

    // 2. Execute the translated query against PostgreSQL
    const sql = `SELECT * FROM products WHERE ${translation.whereClause}`;
    const result = await db.query(sql, translation.parameters);

    return result.rows;
  } catch (error) {
    console.error('Search error:', error.message);
    throw error;
  }
}

// Example usage
async function main() {
  try {
    await db.connect();
    console.log('Connected to PostgreSQL\n');

    // Example 1: Simple field search
    console.log('=== Example 1: Simple field search ===');
    let results = await searchProducts('productCode:13w42');
    console.log(`Found ${results.length} products\n`);

    // Example 2: Boolean AND
    console.log('=== Example 2: Boolean AND ===');
    results = await searchProducts('productCode:13w42 AND region:ca');
    console.log(`Found ${results.length} products\n`);

    // Example 3: Range query
    console.log('=== Example 3: Range query ===');
    results = await searchProducts('rodLength:[50 TO 200]');
    console.log(`Found ${results.length} products\n`);

    // Example 4: Comparison operator
    console.log('=== Example 4: Comparison operator ===');
    results = await searchProducts('price:>=100');
    console.log(`Found ${results.length} products\n`);

    // Example 5: Complex query
    console.log('=== Example 5: Complex query ===');
    results = await searchProducts('(region:ca OR region:ny) AND price:<150');
    console.log(`Found ${results.length} products\n`);

    // Example 6: Wildcard search
    console.log('=== Example 6: Wildcard search ===');
    results = await searchProducts('name:Widget*');
    console.log(`Found ${results.length} products\n`);

  } catch (error) {
    console.error('Error:', error);
    process.exit(1);
  } finally {
    await db.end();
  }
}

// Run if executed directly
if (require.main === module) {
  main();
}

module.exports = { searchProducts };
