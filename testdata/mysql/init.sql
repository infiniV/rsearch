-- rsearch Test Database - MySQL
-- Placeholder for future MySQL translator development

CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_code VARCHAR(50),
    name VARCHAR(255),
    description TEXT,
    rod_length INT,
    price DECIMAL(10,2),
    region VARCHAR(50),
    in_stock BOOLEAN DEFAULT true,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sample data (same as PostgreSQL for consistency)
INSERT INTO products (product_code, name, description, rod_length, price, region, in_stock, status) VALUES
('13w42', 'Widget Pro', 'High quality professional widget', 150, 99.99, 'ca', true, 'active'),
('13w43', 'Widget Basic', 'Standard everyday widget', 200, 49.99, 'ny', true, 'active'),
('15x20', 'Gadget One', 'Multi-purpose gadget for home use', 75, 79.99, 'cb', true, 'active');
