package wallet

import "errors"

var (
	ErrInvalidAmount     = errors.New("Amount must be greater than zero")
	ErrInvalidOperation  = errors.New("Unsupported operation type")
	ErrWalletNotFound    = errors.New("Wallet not found")
	ErrInsufficientFunds = errors.New("Insufficient funds")
	ErrBalanceOverflow   = errors.New("Balance exceeds maximum value")
	ErrInvalidPagination = errors.New(
		"Limit must be between 1 and 100; offset must be non-negative",
	)
)
