const express = require('express');
const { Pool } = require('pg');
const mysql = require('mysql2/promise');
const { MongoClient } = require('mongodb');
const cors = require('cors');
const fs = require('fs');
const path = require('path');

const app = express();
app.use(cors());
app.use(express.json());
app.use(express.static(__dirname));

const RSEARCH_API = process.env.RSEARCH_API || 'http://localhost:8080';

// PostgreSQL connection
const pgPool = new Pool({
    host: process.env.PG_HOST || 'localhost',
    port: parseInt(process.env.PG_PORT || '5432'),
    database: process.env.PG_DATABASE || 'rsearch_test',
    user: process.env.PG_USER || 'rsearch',
    password: process.env.PG_PASSWORD || 'rsearch123',
});

// MySQL connection pool
const mysqlPool = mysql.createPool({
    host: process.env.MYSQL_HOST || 'localhost',
    port: parseInt(process.env.MYSQL_PORT || '3306'),
    database: process.env.MYSQL_DATABASE || 'rsearch_test',
    user: process.env.MYSQL_USER || 'rsearch',
    password: process.env.MYSQL_PASSWORD || 'rsearch123',
    waitForConnections: true,
    connectionLimit: 10,
});

// MongoDB connection
const mongoUrl = process.env.MONGO_URL || 'mongodb://localhost:27017';
const mongoClient = new MongoClient(mongoUrl);
let mongoDb = null;

async function connectMongo() {
    try {
        await mongoClient.connect();
        mongoDb = mongoClient.db('rsearch_test');
        console.log('Connected to MongoDB');
    } catch (err) {
        console.error('MongoDB connection error:', err.message);
    }
}
connectMongo();

// Health check
app.get('/health', (req, res) => {
    res.json({ status: 'ok' });
});

// Execute query against specific database
app.post('/api/execute', async (req, res) => {
    const { schema, database, query } = req.body;

    if (!query) {
        return res.status(400).json({ error: 'Query is required' });
    }

    const db = database || 'postgres';

    try {
        // First translate the query using rsearch API
        const translateResponse = await fetch(`${RSEARCH_API}/api/v1/translate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ schema: schema || 'products', database: db, query })
        });

        if (!translateResponse.ok) {
            const err = await translateResponse.json();
            return res.status(400).json({ error: err.error || err.message || 'Translation failed' });
        }

        const translation = await translateResponse.json();
        const startTime = Date.now();
        let result;

        if (db === 'postgres') {
            const sql = `SELECT * FROM products WHERE ${translation.whereClause} LIMIT 100`;
            const pgResult = await pgPool.query(sql, translation.parameters);
            result = {
                rows: pgResult.rows,
                count: pgResult.rowCount,
                sql: sql,
            };
        } else if (db === 'mysql') {
            const sql = `SELECT * FROM products WHERE ${translation.whereClause} LIMIT 100`;
            const [rows] = await mysqlPool.execute(sql, translation.parameters);
            result = {
                rows: rows,
                count: rows.length,
                sql: sql,
            };
        } else if (db === 'sqlite') {
            // SQLite not available in this demo (would need better-sqlite3)
            return res.status(400).json({ error: 'SQLite execution not available in demo' });
        } else if (db === 'mongodb') {
            if (!mongoDb) {
                return res.status(500).json({ error: 'MongoDB not connected' });
            }
            const filter = translation.filter || {};
            const cursor = mongoDb.collection('products').find(filter).limit(100);
            const rows = await cursor.toArray();
            result = {
                rows: rows,
                count: rows.length,
                filter: JSON.stringify(filter),
            };
        } else {
            return res.status(400).json({ error: `Unknown database: ${db}` });
        }

        const duration = Date.now() - startTime;

        res.json({
            database: db,
            rows: result.rows,
            count: result.count,
            duration: duration,
            sql: result.sql,
            filter: result.filter,
            whereClause: translation.whereClause,
            parameters: translation.parameters
        });

    } catch (err) {
        console.error('Execute error:', err);
        res.status(500).json({ error: err.message });
    }
});

// Execute against all databases
app.post('/api/execute-all', async (req, res) => {
    const { schema, query } = req.body;

    if (!query) {
        return res.status(400).json({ error: 'Query is required' });
    }

    const databases = ['postgres', 'mysql', 'mongodb'];
    const results = {};

    for (const db of databases) {
        try {
            const translateResponse = await fetch(`${RSEARCH_API}/api/v1/translate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ schema: schema || 'products', database: db, query })
            });

            if (!translateResponse.ok) {
                results[db] = { error: 'Translation failed' };
                continue;
            }

            const translation = await translateResponse.json();
            const startTime = Date.now();

            if (db === 'postgres') {
                const sql = `SELECT * FROM products WHERE ${translation.whereClause} LIMIT 100`;
                const pgResult = await pgPool.query(sql, translation.parameters);
                results[db] = {
                    rows: pgResult.rows,
                    count: pgResult.rowCount,
                    duration: Date.now() - startTime,
                    sql: sql,
                };
            } else if (db === 'mysql') {
                const sql = `SELECT * FROM products WHERE ${translation.whereClause} LIMIT 100`;
                const [rows] = await mysqlPool.execute(sql, translation.parameters);
                results[db] = {
                    rows: rows,
                    count: rows.length,
                    duration: Date.now() - startTime,
                    sql: sql,
                };
            } else if (db === 'mongodb') {
                if (mongoDb) {
                    const filter = translation.filter || {};
                    const cursor = mongoDb.collection('products').find(filter).limit(100);
                    const rows = await cursor.toArray();
                    results[db] = {
                        rows: rows,
                        count: rows.length,
                        duration: Date.now() - startTime,
                        filter: JSON.stringify(filter),
                    };
                } else {
                    results[db] = { error: 'MongoDB not connected' };
                }
            }
        } catch (err) {
            results[db] = { error: err.message };
        }
    }

    res.json(results);
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
        const result = await pgPool.query(`
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
        const result = await pgPool.query('SELECT * FROM products LIMIT 10');
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
