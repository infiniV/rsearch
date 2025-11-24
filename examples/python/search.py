#!/usr/bin/env python3
"""
rsearch Integration Example - Python
Demonstrates how to use rsearch with PostgreSQL in a Python application
"""

import os
import psycopg2
import requests
from typing import List, Dict, Any

RSEARCH_URL = os.getenv('RSEARCH_URL', 'http://localhost:8080')

# PostgreSQL connection parameters
DB_CONFIG = {
    'host': 'localhost',
    'port': 5432,
    'database': 'rsearch_test',
    'user': 'rsearch',
    'password': 'rsearch123'
}


def search_products(user_query: str) -> List[Dict[str, Any]]:
    """
    Translate user query using rsearch and execute against PostgreSQL

    Args:
        user_query: OpenSearch-style query from user

    Returns:
        List of product records matching the query
    """
    try:
        # 1. Call rsearch to translate the query
        response = requests.post(
            f'{RSEARCH_URL}/api/v1/translate',
            json={
                'schema': 'products',
                'database': 'postgres',
                'query': user_query
            }
        )
        response.raise_for_status()
        translation = response.json()

        print(f'Query: {user_query}')
        print(f'SQL: {translation["whereClause"]}')
        print(f'Params: {translation["parameters"]}')
        print('---')

        # 2. Execute the translated query against PostgreSQL
        conn = psycopg2.connect(**DB_CONFIG)
        cur = conn.cursor()

        sql = f"SELECT * FROM products WHERE {translation['whereClause']}"
        cur.execute(sql, translation['parameters'])

        # Fetch results with column names
        columns = [desc[0] for desc in cur.description]
        results = [dict(zip(columns, row)) for row in cur.fetchall()]

        cur.close()
        conn.close()

        return results

    except requests.exceptions.RequestException as e:
        print(f'rsearch error: {e}')
        raise
    except psycopg2.Error as e:
        print(f'Database error: {e}')
        raise


def main():
    """Example usage of rsearch with various query patterns"""

    examples = [
        ('Simple field search', 'productCode:13w42'),
        ('Boolean AND', 'productCode:13w42 AND region:ca'),
        ('Range query', 'rodLength:[50 TO 200]'),
        ('Comparison operator', 'price:>=100'),
        ('Complex query', '(region:ca OR region:ny) AND price:<150'),
        ('Wildcard search', 'name:Widget*'),
    ]

    for title, query in examples:
        print(f'\n=== Example: {title} ===')
        try:
            results = search_products(query)
            print(f'Found {len(results)} products')

            # Display first result if any
            if results:
                product = results[0]
                print(f'  - {product["name"]} ({product["product_code"]}) - ${product["price"]}')

        except Exception as e:
            print(f'Error: {e}')


if __name__ == '__main__':
    main()
