//go:build integration

package integration_test

import (
	"context"
	"math"
	"os"
	"testing"
	"time"

	"test-wallet/internal/postgres"
	"test-wallet/internal/wallet"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T) (*wallet.PostgresRepository, uuid.UUID) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(
		t, dsn, "TEST_DATABASE_URL is required; run make test-integration")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := postgres.NewPool(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := wallet.NewPostgresRepository(pool)
	id := uuid.New()
	created, err := repo.Create(ctx, id)
	require.NoError(t, err)
	require.Equal(t, id, created.ID)
	require.Zero(t, created.Balance)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := pool.Exec(ctx, "DELETE FROM wallet WHERE id = $1", id)
		require.NoError(t, err, "cleanup wallet")
	})
	return repo, id
}

func TestPostgresBalanceErrorsAndRollback(t *testing.T) {
	repo, id := fixture(t)
	ctx := context.Background()
	for _, delta := range []int64{100, -30} {
		_, err := repo.ChangeBalance(ctx, id, delta)
		require.NoError(t, err)
	}

	_, err := repo.ChangeBalance(ctx, id, -71)
	require.ErrorIs(t, err, wallet.ErrInsufficientFunds)
	balance, err := repo.GetBalance(ctx, id)
	require.NoError(t, err)
	require.Equal(t, int64(70), balance, "failed withdrawal changed balance")

	_, err = repo.ChangeBalance(ctx, id, math.MinInt64)
	require.ErrorIs(t, err, wallet.ErrInsufficientFunds)
	_, err = repo.ChangeBalance(ctx, id, math.MaxInt64-70)
	require.NoError(t, err)
	_, err = repo.ChangeBalance(ctx, id, 1)
	require.ErrorIs(t, err, wallet.ErrBalanceOverflow)
	balance, err = repo.GetBalance(ctx, id)
	require.NoError(t, err)
	require.Equal(t, int64(math.MaxInt64), balance, "overflow changed balance")

	balance, err = repo.ChangeBalance(ctx, id, -math.MaxInt64)
	require.NoError(t, err, "transaction lock was not released")
	require.Zero(t, balance)

	missing := uuid.New()
	_, err = repo.GetBalance(ctx, missing)
	require.ErrorIs(t, err, wallet.ErrWalletNotFound)
	_, err = repo.ChangeBalance(ctx, missing, 1)
	require.ErrorIs(t, err, wallet.ErrWalletNotFound)

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = repo.ChangeBalance(canceled, id, 1)
	require.ErrorIs(t, err, context.Canceled)
}
