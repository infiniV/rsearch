const express = require('express');
const { Pool } = require('pg');
const cors = require('cors');
const fs = require('fs');
const path = require('path');

const app = express();
app.use(cors());
app.use(express.json());
app.use(express.static(__dirname));

const RSEARCH_API = process.env.RSEARCH_API || 'http://localhost:8080';

const pool = new Pool({
    host: process.env.PG_HOST || 'localhost',
    port: parseInt(process.env.PG_PORT || '5432'),
    database: process.env.PG_DATABASE || 'rsearch_test',
    user: process.env.PG_USER || 'rsearch',
    password: process.env.PG_PASSWORD || 'rsearch123',
});

// Health check
app.get('/health', (req, res) => {
    res.json({ status: 'ok' });
});

// Execute query against database
app.post('/api/execute', async (req, res) => {
    const { schema, database, query } = req.body;

    if (!query) {
        return res.status(400).json({ error: 'Query is required' });
    }

    try {
        // First translate the query using rsearch API
        const translateResponse = await fetch(`${RSEARCH_API}/api/v1/translate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ schema: schema || 'products', database: database || 'postgres', query })
        });

        if (!translateResponse.ok) {
            const err = await translateResponse.json();
            return res.status(400).json({ error: err.error || 'Translation failed' });
        }

        const translation = await translateResponse.json();

        // Build full SQL query
        const sql = `SELECT * FROM products WHERE ${translation.whereClause} LIMIT 100`;

        // Execute against PostgreSQL
        const startTime = Date.now();
        const result = await pool.query(sql, translation.parameters);
        const duration = Date.now() - startTime;

        res.json({
            rows: result.rows,
            count: result.rowCount,
            duration: duration,
            sql: sql,
            whereClause: translation.whereClause,
            parameters: translation.parameters
        });

    } catch (err) {
        console.error('Execute error:', err);
        res.status(500).json({ error: err.message });
    }
});

// Serve syntax reference markdown
app.get('/api/docs/syntax', (req, res) => {
    const docsPath = path.join(__dirname, '..', 'docs', 'syntax-reference.md');

    if (fs.existsSync(docsPath)) {
        const content = fs.readFileSync(docsPath, 'utf8');
        res.type('text/markdown').send(content);
    } else {
        res.status(404).json({ error: 'Documentation not found' });
    }
});

// Get table schema info
app.get('/api/schema/products', async (req, res) => {
    try {
        const result = await pool.query(`
            SELECT column_name, data_type, is_nullable
            FROM information_schema.columns
            WHERE table_name = 'products'
            ORDER BY ordinal_position
        `);
        res.json({ columns: result.rows });
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

// Get sample data
app.get('/api/data/sample', async (req, res) => {
    try {
        const result = await pool.query('SELECT * FROM products LIMIT 10');
        res.json({ rows: result.rows, count: result.rowCount });
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

const PORT = process.env.PORT || 3000;

app.listen(PORT, () => {
    console.log(`Demo server running at http://localhost:${PORT}`);
    console.log(`Open http://localhost:${PORT}/demo.html in your browser`);
});
