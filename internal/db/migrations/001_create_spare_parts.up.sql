CREATE TABLE IF NOT EXISTS spare_parts (
    id TEXT PRIMARY KEY,
    reference TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    brand TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL
);

INSERT INTO spare_parts (id, reference, label, brand, category, description)
VALUES
    (
        'sp-001',
        'BRK-PAD-4521',
        'Front Brake Pads',
        'Brembo',
        'BRAKING',
        'High performance front brake pads for urban and highway use'
    ),
    (
        'sp-002',
        'ENG-FLT-7823',
        'Oil Filter',
        'Mann',
        'FILTERS',
        'Standard oil filter compatible with most 4-cylinder engines'
    ),
    (
        'sp-003',
        'SUS-SPR-3341',
        'Rear Coil Spring',
        'KYB',
        'SUSPENSION',
        'OEM replacement rear coil spring for medium sedans'
    )
ON CONFLICT (id) DO UPDATE SET
    reference = EXCLUDED.reference,
    label = EXCLUDED.label,
    brand = EXCLUDED.brand,
    category = EXCLUDED.category,
    description = EXCLUDED.description;
