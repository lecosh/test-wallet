-- +goose Up

ALTER TABLE Wallet
    ALTER COLUMN amount TYPE BIGINT,
    ALTER COLUMN amount SET DEFAULT 0,
    ALTER COLUMN amount SET NOT NULL;

ALTER TABLE Wallet
    ADD CONSTRAINT wallet_amount_nonnegative CHECK (amount >= 0);

-- +goose Down

ALTER TABLE Wallet
    DROP CONSTRAINT wallet_amount_nonnegative;

ALTER TABLE Wallet
    ALTER COLUMN amount DROP DEFAULT,
    ALTER COLUMN amount DROP NOT NULL,
    ALTER COLUMN amount TYPE INTEGER;
