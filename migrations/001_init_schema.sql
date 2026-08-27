CREATE TABLE IF NOT EXISTS restaurants (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    city VARCHAR(64) NOT NULL,
    cuisine VARCHAR(64) NOT NULL,
    delivery_eta_min INTEGER NOT NULL,
    commission_rate NUMERIC(5,2) NOT NULL DEFAULT 0.10,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS restaurant_items (
    id SERIAL PRIMARY KEY,
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    category VARCHAR(64) NOT NULL,
    price_cents INTEGER NOT NULL CHECK (price_cents > 0),
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (restaurant_id, name)
);

CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(64) UNIQUE,
    display_name VARCHAR(128) NOT NULL,
    phone VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    restaurant_id INTEGER NOT NULL REFERENCES restaurants(id),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    total_cents INTEGER NOT NULL CHECK (total_cents >= 0),
    delivery_address TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES restaurant_items(id),
    name_snapshot VARCHAR(128) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents INTEGER NOT NULL CHECK (unit_price_cents > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_restaurant_items_restaurant_id
    ON restaurant_items (restaurant_id);

CREATE INDEX IF NOT EXISTS idx_orders_restaurant_id
    ON orders (restaurant_id);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id
    ON orders (customer_id);
