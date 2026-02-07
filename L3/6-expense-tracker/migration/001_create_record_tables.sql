CREATE TABLE IF NOT EXISTS record (
    id BIGSERIAL PRIMARY KEY,
    type INTEGER NOT NULL,
    category INTEGER NOT NULL,
    amount DECIMAL(15, 2) NOT NULL CHECK (amount >= 0),
    date DATE NOT NULL
);
