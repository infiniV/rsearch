// rsearch Integration Example - Go
// Demonstrates how to use rsearch with PostgreSQL in a Go application

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var rsearchURL = getEnv("RSEARCH_URL", "http://localhost:8080")

type TranslateRequest struct {
	Schema   string `json:"schema"`
	Database string `json:"database"`
	Query    string `json:"query"`
}

type TranslateResponse struct {
	Type           string        `json:"type"`
	WhereClause    string        `json:"whereClause"`
	Parameters     []interface{} `json:"parameters"`
	ParameterTypes []string      `json:"parameterTypes"`
}

type Product struct {
	ID          int
	ProductCode string
	Name        string
	Description string
	RodLength   int
	Price       float64
	Region      string
	InStock     bool
	Status      string
}

// SearchProducts translates user query using rsearch and executes against PostgreSQL
func SearchProducts(db *sql.DB, userQuery string) ([]Product, error) {
	// 1. Call rsearch to translate the query
	reqBody, _ := json.Marshal(TranslateRequest{
		Schema:   "products",
		Database: "postgres",
		Query:    userQuery,
	})

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/translate", rsearchURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("rsearch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rsearch returned status %d", resp.StatusCode)
	}

	var translation TranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&translation); err != nil {
		return nil, fmt.Errorf("failed to decode rsearch response: %w", err)
	}

	fmt.Printf("Query: %s\n", userQuery)
	fmt.Printf("SQL: %s\n", translation.WhereClause)
	fmt.Printf("Params: %v\n", translation.Parameters)
	fmt.Println("---")

	// 2. Execute the translated query against PostgreSQL
	query := fmt.Sprintf("SELECT id, product_code, name, description, rod_length, price, region, in_stock, status FROM products WHERE %s", translation.WhereClause)
	rows, err := db.Query(query, translation.Parameters...)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.ProductCode, &p.Name, &p.Description, &p.RodLength, &p.Price, &p.Region, &p.InStock, &p.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}

func main() {
	// Connect to PostgreSQL
	connStr := "host=localhost port=5432 user=rsearch password=rsearch123 dbname=rsearch_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to PostgreSQL\n")

	// Example queries
	examples := []struct {
		title string
		query string
	}{
		{"Simple field search", "productCode:13w42"},
		{"Boolean AND", "productCode:13w42 AND region:ca"},
		{"Range query", "rodLength:[50 TO 200]"},
		{"Comparison operator", "price:>=100"},
		{"Complex query", "(region:ca OR region:ny) AND price:<150"},
		{"Wildcard search", "name:Widget*"},
	}

	for _, example := range examples {
		fmt.Printf("\n=== Example: %s ===\n", example.title)

		products, err := SearchProducts(db, example.query)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Found %d products\n", len(products))
		if len(products) > 0 {
			p := products[0]
			fmt.Printf("  - %s (%s) - $%.2f\n", p.Name, p.ProductCode, p.Price)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
