package wallet

import "github.com/google/uuid"

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type Wallet struct {
	ID      uuid.UUID
	Balance int64
}