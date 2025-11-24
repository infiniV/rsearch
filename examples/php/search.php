<?php
/**
 * rsearch Integration Example - PHP
 * Demonstrates how to use rsearch with PostgreSQL in a PHP application
 */

// Configuration
$rsearchURL = getenv('RSEARCH_URL') ?: 'http://localhost:8080';
$dbConfig = [
    'host' => 'localhost',
    'port' => '5432',
    'dbname' => 'rsearch_test',
    'user' => 'rsearch',
    'password' => 'rsearch123'
];

/**
 * Translate user query using rsearch and execute against PostgreSQL
 *
 * @param PDO $pdo Database connection
 * @param string $userQuery OpenSearch-style query from user
 * @return array Query results
 */
function searchProducts($pdo, $userQuery) {
    global $rsearchURL;

    // 1. Call rsearch to translate the query
    $requestData = json_encode([
        'schema' => 'products',
        'database' => 'postgres',
        'query' => $userQuery
    ]);

    $context = stream_context_create([
        'http' => [
            'method' => 'POST',
            'header' => 'Content-Type: application/json',
            'content' => $requestData
        ]
    ]);

    $response = file_get_contents("$rsearchURL/api/v1/translate", false, $context);
    if ($response === false) {
        throw new Exception('rsearch request failed');
    }

    $translation = json_decode($response, true);
    if (json_last_error() !== JSON_ERROR_NONE) {
        throw new Exception('Failed to decode rsearch response');
    }

    echo "Query: $userQuery\n";
    echo "SQL: {$translation['whereClause']}\n";
    echo "Params: " . json_encode($translation['parameters']) . "\n";
    echo "---\n";

    // 2. Execute the translated query against PostgreSQL
    $sql = "SELECT * FROM products WHERE {$translation['whereClause']}";
    $stmt = $pdo->prepare($sql);
    $stmt->execute($translation['parameters']);

    return $stmt->fetchAll(PDO::FETCH_ASSOC);
}

// Main execution
try {
    // Connect to PostgreSQL
    $dsn = sprintf(
        'pgsql:host=%s;port=%s;dbname=%s',
        $dbConfig['host'],
        $dbConfig['port'],
        $dbConfig['dbname']
    );
    $pdo = new PDO($dsn, $dbConfig['user'], $dbConfig['password']);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    echo "Connected to PostgreSQL\n\n";

    // Example queries
    $examples = [
        ['title' => 'Simple field search', 'query' => 'productCode:13w42'],
        ['title' => 'Boolean AND', 'query' => 'productCode:13w42 AND region:ca'],
        ['title' => 'Range query', 'query' => 'rodLength:[50 TO 200]'],
        ['title' => 'Comparison operator', 'query' => 'price:>=100'],
        ['title' => 'Complex query', 'query' => '(region:ca OR region:ny) AND price:<150'],
        ['title' => 'Wildcard search', 'query' => 'name:Widget*']
    ];

    foreach ($examples as $example) {
        echo "\n=== Example: {$example['title']} ===\n";

        try {
            $results = searchProducts($pdo, $example['query']);
            echo "Found " . count($results) . " products\n";

            if (!empty($results)) {
                $product = $results[0];
                echo "  - {$product['name']} ({$product['product_code']}) - \${$product['price']}\n";
            }
        } catch (Exception $e) {
            echo "Error: {$e->getMessage()}\n";
        }
    }

} catch (PDOException $e) {
    echo "Database error: {$e->getMessage()}\n";
    exit(1);
} catch (Exception $e) {
    echo "Error: {$e->getMessage()}\n";
    exit(1);
}
