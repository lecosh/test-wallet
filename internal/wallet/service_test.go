package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type repositoryStub struct {
	get    func(context.Context, uuid.UUID) (int64, error)
	change func(context.Context, uuid.UUID, int64) (int64, error)
	create func(context.Context, uuid.UUID) (Wallet, error)
	list   func(context.Context, int, int) ([]Wallet, error)
}

func (r repositoryStub) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.get(ctx, id)
}
func (r repositoryStub) ChangeBalance(ctx context.Context, id uuid.UUID, delta int64) (int64, error) {
	return r.change(ctx, id, delta)
}
func (r repositoryStub) Create(ctx context.Context, id uuid.UUID) (Wallet, error) {
	return r.create(ctx, id)
}
func (r repositoryStub) List(ctx context.Context, limit, offset int) ([]Wallet, error) {
	return r.list(ctx, limit, offset)
}

var updateBalanceTests = []struct {
	name      string
	operation OperationType
	amount    int64
	delta     int64
	wantErr   error
}{
	{"deposit", Deposit, 100, 100, nil},
	{"withdraw", Withdraw, 100, -100, nil},
	{"zero amount", Deposit, 0, 0, ErrInvalidAmount},
	{"negative amount", Withdraw, -1, 0, ErrInvalidAmount},
	{"unknown operation", "TRANSFER", 100, 0, ErrInvalidOperation},
}

func TestUpdateBalance(t *testing.T) {
	for _, tt := range updateBalanceTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			id := uuid.New()
			calls := 0

			s := NewService(repositoryStub{change: func(gotCtx context.Context, gotID uuid.UUID, delta int64) (int64, error) {
				calls++
				if gotCtx != ctx || gotID != id || delta != tt.delta {
					t.Errorf("unexpected repository arguments: ID=%s delta=%d", gotID, delta)
				}
				return 900, nil
			}})

			balance, err := s.UpdateBalance(ctx, id, tt.operation, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if calls != 0 {
					t.Fatal("invalid operation reached repository")
				}
			} else if calls != 1 || balance != 900 {
				t.Fatalf("calls=%d balance=%d", calls, balance)
			}
		})
	}
}
