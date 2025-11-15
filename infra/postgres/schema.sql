-- Database schema for recommendation engine

-- Items table
CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    sku TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    category TEXT,
    price BIGINT DEFAULT 0,
    stock INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_items_category ON items(category);
CREATE INDEX idx_items_sku ON items(sku);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    external_id TEXT UNIQUE,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_users_external_id ON users(external_id);

-- Events table (append-only for raw events)
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    item_id BIGINT,
    event_type TEXT NOT NULL, -- VIEW, CLICK, CART, PURCHASE
    session_id TEXT,
    metadata JSONB,
    timestamp TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_item_id ON events(item_id);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
CREATE INDEX idx_events_session_id ON events(session_id);
CREATE INDEX idx_events_type ON events(event_type);

-- Models metadata table
CREATE TABLE IF NOT EXISTS models (
    id BIGSERIAL PRIMARY KEY,
    model_name TEXT NOT NULL,
    version TEXT NOT NULL,
    model_type TEXT, -- item2vec, als, neural, etc.
    metrics JSONB,
    config JSONB,
    created_at TIMESTAMP DEFAULT now(),
    UNIQUE(model_name, version)
);

CREATE INDEX idx_models_name ON models(model_name);

-- Item embeddings table (optional, if not using external vector DB)
CREATE TABLE IF NOT EXISTS item_embeddings (
    item_id BIGINT PRIMARY KEY REFERENCES items(id),
    model_id BIGINT REFERENCES models(id),
    embedding FLOAT8[],
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_item_embeddings_model ON item_embeddings(model_id);

-- Seed some sample data
INSERT INTO items (sku, title, category, price, stock) VALUES
('SKU001', 'Laptop Gaming ASUS ROG', 'electronics', 15000000, 10),
('SKU002', 'Mouse Wireless Logitech', 'electronics', 250000, 50),
('SKU003', 'Keyboard Mechanical', 'electronics', 800000, 30),
('SKU004', 'Monitor 27 inch', 'electronics', 3000000, 15),
('SKU005', 'Headset Gaming', 'electronics', 500000, 25),
('SKU006', 'Smartphone Samsung', 'electronics', 5000000, 20),
('SKU007', 'Smartwatch', 'electronics', 2000000, 12),
('SKU008', 'Tablet iPad', 'electronics', 8000000, 8),
('SKU009', 'Camera DSLR', 'electronics', 12000000, 5),
('SKU010', 'Speaker Bluetooth', 'electronics', 300000, 40)
ON CONFLICT (sku) DO NOTHING;

INSERT INTO users (external_id) VALUES
('user_001'),
('user_002'),
('user_003'),
('user_004'),
('user_005')
ON CONFLICT (external_id) DO NOTHING;
