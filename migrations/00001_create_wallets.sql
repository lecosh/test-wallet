-- +goose Up

CREATE TABLE IF NOT EXISTS Wallet (
    id UUID PRIMARY KEY,
    amount INT
);

-- +goose Down

DROP TABLE Wallet;