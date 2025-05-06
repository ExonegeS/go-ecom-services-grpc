CREATE TABLE user_order_statistics (
    user_id UUID PRIMARY KEY,
    total_orders INT DEFAULT 0,
    hourly_distribution JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_statistics (
    user_id UUID PRIMARY KEY,
    total_items_purchased INT DEFAULT 0,
    average_order_value NUMERIC(10,2) DEFAULT 0.00,
    most_purchased_item TEXT,
    total_completed_orders INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);