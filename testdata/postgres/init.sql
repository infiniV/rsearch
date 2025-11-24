-- rsearch Test Database - PostgreSQL
-- Comprehensive product data for demo and testing

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

DROP TABLE IF EXISTS products CASCADE;

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    product_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
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

-- Comprehensive sample data covering all query types
INSERT INTO products (product_code, name, description, rod_length, price, region, in_stock, status, tags, metadata) VALUES
-- Basic products (Widget series)
('13w42', 'Widget Pro', 'High quality professional widget for experts', 150, 99.99, 'ca', true, 'active', ARRAY['premium', 'bestseller'], '{"warranty": 24, "rating": 4.8}'),
('13w43', 'Widget Basic', 'Standard everyday widget for beginners', 200, 49.99, 'ny', true, 'active', ARRAY['basic', 'starter'], '{"warranty": 12, "rating": 4.2}'),
('13w44', 'Widget Premium', 'Premium quality widget with extended warranty', 450, 199.99, 'ca', false, 'discontinued', ARRAY['premium', 'legacy'], '{"warranty": 36, "rating": 4.9}'),
('13w45', 'Widget Mini', 'Compact widget for small spaces', 75, 29.99, 'ny', true, 'active', ARRAY['compact', 'portable'], '{"warranty": 6, "rating": 4.0}'),
('13w46', 'Widget Ultra', 'Ultra-performance widget for professionals', 350, 299.99, 'ca', true, 'active', ARRAY['premium', 'professional', 'new'], '{"warranty": 48, "rating": 4.95}'),

-- Gadget series
('15x20', 'Gadget One', 'Multi-purpose gadget for home use', 75, 79.99, 'cb', true, 'active', ARRAY['gadget', 'home'], '{"warranty": 12, "rating": 4.3}'),
('15x21', 'Gadget Two', 'Advanced gadget with smart features', 125, 129.99, 'ca', true, 'active', ARRAY['gadget', 'smart', 'advanced'], '{"warranty": 18, "rating": 4.6}'),
('15x22', 'Gadget Pro', 'Professional-grade gadget with extended features', 180, 189.99, 'ny', true, 'active', ARRAY['gadget', 'professional'], '{"warranty": 24, "rating": 4.7}'),
('15x23', 'Gadget Lite', 'Lightweight gadget for everyday use', 60, 59.99, 'cb', true, 'active', ARRAY['gadget', 'portable', 'light'], '{"warranty": 12, "rating": 4.1}'),

-- Gizmo series
('16z10', 'Gizmo Alpha', 'Compact gizmo for professionals', 90, 89.99, 'ny', true, 'active', ARRAY['professional', 'compact'], '{"warranty": 12, "rating": 4.4}'),
('16z11', 'Gizmo Beta', 'Full-featured gizmo with premium build', 180, 149.99, 'ca', true, 'active', ARRAY['professional', 'premium', 'featured'], '{"warranty": 24, "rating": 4.6}'),
('16z12', 'Gizmo Gamma', 'Next-gen gizmo with AI capabilities', 220, 249.99, 'ny', true, 'preorder', ARRAY['professional', 'ai', 'new'], '{"warranty": 24, "rating": null}'),
('16z13', 'Gizmo Delta', NULL, 150, 119.99, 'ca', true, 'active', ARRAY['standard'], '{"warranty": 12, "rating": 4.2}'),

-- Tool sets
('17a05', 'Tool Set A', 'Complete tool set for beginners', 300, 299.99, 'ca', true, 'active', ARRAY['tools', 'complete', 'beginner'], '{"warranty": 36, "rating": 4.5}'),
('17a06', 'Tool Set B', 'Professional grade tool set with premium case', 450, 499.99, 'ny', true, 'active', ARRAY['tools', 'professional', 'premium'], '{"warranty": 60, "rating": 4.8}'),
('17a07', 'Tool Set C', 'Compact portable tool kit', 100, 149.99, 'cb', true, 'active', ARRAY['tools', 'portable', 'compact'], '{"warranty": 24, "rating": 4.3}'),

-- Device series
('18b20', 'Device X', 'Experimental device with new technology', 60, 599.99, 'cb', false, 'preorder', ARRAY['experimental', 'new', 'tech'], '{"warranty": 12, "rating": null}'),
('18b21', 'Device Y', 'Standard device for everyday computing', 80, 399.99, 'ca', true, 'active', ARRAY['device', 'standard'], '{"warranty": 24, "rating": 4.4}'),
('18b22', 'Device Z', 'High-performance device for gaming', 120, 799.99, 'ny', true, 'active', ARRAY['device', 'gaming', 'premium'], '{"warranty": 36, "rating": 4.7}'),

-- Mixed regions for testing
('19c01', 'Accessor Pro', 'Professional accessory kit', 25, 49.99, 'ca', true, 'active', ARRAY['accessory', 'kit'], '{"warranty": 12, "rating": 4.1}'),
('19c02', 'Accessor Basic', 'Basic accessory bundle', 20, 24.99, 'ny', true, 'active', ARRAY['accessory', 'basic'], '{"warranty": 6, "rating": 3.9}'),
('19c03', 'Accessor Plus', 'Enhanced accessory package', 30, 79.99, 'cb', true, 'active', ARRAY['accessory', 'enhanced'], '{"warranty": 18, "rating": 4.3}'),

-- Edge cases for testing
('20d01', 'Special Item', 'Item with special characters: & < > " in name', 50, 19.99, 'ca', true, 'active', ARRAY['special'], '{}'),
('20d02', 'Zero Price Item', 'Free promotional item', 10, 0.00, 'ny', true, 'active', ARRAY['free', 'promo'], '{}'),
('20d03', 'Max Length Rod', 'Item with maximum rod length', 999, 149.99, 'ca', true, 'active', ARRAY['large'], '{}'),
('20d04', 'Min Length Rod', 'Item with minimum rod length', 1, 9.99, 'ny', true, 'active', ARRAY['tiny'], '{}');

-- Create indexes for optimized queries
CREATE INDEX idx_products_code ON products(product_code);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_region ON products(region);
CREATE INDEX idx_products_rod_length ON products(rod_length);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_in_stock ON products(in_stock);

-- Full-text search index (for future use)
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
CREATE INDEX idx_products_description_trgm ON products USING gin (description gin_trgm_ops);

-- Comment for documentation
COMMENT ON TABLE products IS 'Sample product catalog for rsearch query translation demo - 25 products covering all query patterns';
