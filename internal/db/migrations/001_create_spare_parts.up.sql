CREATE TABLE IF NOT EXISTS spare_parts (
    id TEXT PRIMARY KEY,
    reference TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    brand TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL
);
