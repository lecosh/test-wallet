package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

// Compile check
var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	const query = `
		SELECT amount
		FROM wallet
		WHERE id = $1
	`
	var balance int64

	err := r.pool.QueryRow(ctx, query, walletID).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get wallet balance: %w", err)
	}

	return balance, nil
}

func (r *PostgresRepository) ChangeBalance(
	ctx context.Context,
	walletID uuid.UUID,
	delta int64,
) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin balance transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	const selectQuery = `
		SELECT amount
		FROM wallet
		WHERE id = $1
		FOR UPDATE
	`

	var balance int64

	err = tx.QueryRow(ctx, selectQuery, walletID).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lock wallet: %w", err)
	}

	if delta < 0 && delta < -balance {
		return 0, ErrInsufficientFunds
	}

	if delta > 0 && balance > math.MaxInt64-delta {
		return 0, ErrBalanceOverflow
	}

	const updateQuery = `
		UPDATE wallet
		SET amount = amount + $2
		WHERE id = $1
		RETURNING amount
	`

	var newBalance int64

	err = tx.QueryRow(
		ctx,
		updateQuery,
		walletID,
		delta,
	).Scan(&newBalance)
	if err != nil {
		return 0, fmt.Errorf("update wallet balance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit balance transaction: %w", err)
	}

	return newBalance, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	walletID uuid.UUID,
) (Wallet, error) {
	const query = `
		INSERT INTO wallet (id, amount)
		VALUES ($1, 0)
		RETURNING id, amount
	`

	var result Wallet

	err := r.pool.QueryRow(ctx, query, walletID).Scan(
		&result.ID,
		&result.Balance,
	)
	if err != nil {
		return Wallet{}, fmt.Errorf("create wallet: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) List(
	ctx context.Context,
	limit, offset int,
) ([]Wallet, error) {
	const query = `
		SELECT id, amount
		FROM wallet
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list wallets: %w", err)
	}
	defer rows.Close()

	wallets := make([]Wallet, 0)

	for rows.Next() {
		var item Wallet

		if err := rows.Scan(&item.ID, &item.Balance); err != nil {
			return nil, fmt.Errorf("scan wallet: %w", err)
		}

		wallets = append(wallets, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read wallets: %w", err)
	}

	return wallets, nil
}
