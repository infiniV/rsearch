-- rsearch Test Database - PostgreSQL
-- Sample product data for testing query translation

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    product_code VARCHAR(50),
    name VARCHAR(255),
    description TEXT,
    rod_length INTEGER,
    price DECIMAL(10,2),
    region VARCHAR(50),
    in_stock BOOLEAN DEFAULT true,
    status VARCHAR(50) DEFAULT 'active',
    tags TEXT[],
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sample data
INSERT INTO products (product_code, name, description, rod_length, price, region, in_stock, status, tags) VALUES
('13w42', 'Widget Pro', 'High quality professional widget', 150, 99.99, 'ca', true, 'active', ARRAY['premium', 'bestseller']),
('13w43', 'Widget Basic', 'Standard everyday widget', 200, 49.99, 'ny', true, 'active', ARRAY['basic']),
('13w44', 'Widget Premium', 'Premium quality widget with extended warranty', 450, 199.99, 'ca', false, 'discontinued', ARRAY['premium', 'new']),
('15x20', 'Gadget One', 'Multi-purpose gadget for home use', 75, 79.99, 'cb', true, 'active', ARRAY['gadget']),
('15x21', 'Gadget Two', 'Advanced gadget with smart features', 125, 129.99, 'ca', true, 'active', ARRAY['gadget', 'advanced']),
('16z10', 'Gizmo Alpha', 'Compact gizmo for professionals', 90, 89.99, 'ny', true, 'active', ARRAY['professional', 'compact']),
('16z11', 'Gizmo Beta', 'Full-featured gizmo with premium build', 180, 149.99, 'ca', true, 'active', ARRAY['professional', 'premium']),
('17a05', 'Tool Set A', 'Complete tool set for beginners', 300, 299.99, 'ca', true, 'active', ARRAY['tools', 'complete']),
('17a06', 'Tool Set B', 'Professional grade tool set', 450, 499.99, 'ny', true, 'active', ARRAY['tools', 'professional']),
('18b20', 'Device X', 'Experimental device with new technology', 60, 599.99, 'cb', false, 'preorder', ARRAY['experimental', 'new']);

-- Indexes for common queries
CREATE INDEX idx_product_code ON products(product_code);
CREATE INDEX idx_region ON products(region);
CREATE INDEX idx_rod_length ON products(rod_length);
CREATE INDEX idx_price ON products(price);
CREATE INDEX idx_status ON products(status);

-- Test query examples
COMMENT ON TABLE products IS 'Sample product catalog for testing rsearch query translation';
