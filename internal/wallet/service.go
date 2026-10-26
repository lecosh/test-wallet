package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	GetBalance(
		ctx context.Context, 
		walletID uuid.UUID,
	) (int64, error)

	ChangeBalance(
		ctx context.Context,
		walletID uuid.UUID,
		delta int64,
	) (int64, error)

	Create(
		ctx context.Context, 
		walletID uuid.UUID,
	) (Wallet, error)
	
	List(
		ctx context.Context, 
		limit int, 
		offset int,
	) ([]Wallet, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) UpdateBalance(
	ctx context.Context,
	walletID uuid.UUID,
	operation OperationType,
	amount int64,
) (int64, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	var delta int64

	switch operation {
	case Deposit:
		delta = amount
	case Withdraw:
		delta = -amount
	default:
		return 0, ErrInvalidOperation
	}

	return s.repo.ChangeBalance(ctx, walletID, delta)
}

func (s *Service) GetBalance(
	ctx context.Context,
	walletID uuid.UUID,
) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}

func (s *Service) CreateWallet(ctx context.Context) (Wallet, error) {
	walletID, err := uuid.NewRandom()
	if err != nil {
		return Wallet{}, fmt.Errorf("generate wallet ID: %w", err)
	}

	return s.repo.Create(ctx, walletID)
}

func (s *Service) ListWallets(
	ctx context.Context,
	limit, offset int,
) ([]Wallet, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}

	return s.repo.List(ctx, limit, offset)
}
