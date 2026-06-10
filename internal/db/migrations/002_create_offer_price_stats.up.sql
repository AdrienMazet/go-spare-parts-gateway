CREATE TABLE IF NOT EXISTS offer_price_stats (
    reference TEXT NOT NULL,
    currency TEXT NOT NULL,
    observed_count INTEGER NOT NULL,
    min_price INTEGER NOT NULL,
    max_price INTEGER NOT NULL,
    latest_price INTEGER NOT NULL,
    average_price NUMERIC(12, 2) NOT NULL,
    last_observed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (reference, currency)
);
