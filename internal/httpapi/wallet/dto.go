package wallet_api

type UpdateBalanceRequest struct {
	WalletID      string `json:"walletId" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000" binding:"required"`
	OperationType string `json:"operationType" enums:"DEPOSIT,WITHDRAW" example:"DEPOSIT" binding:"required"`
	Amount        int64  `json:"amount" minimum:"1" example:"1000" binding:"required"`
} // @name UpdateBalanceRequest

type WalletResponse struct {
	WalletID string `json:"walletId" format:"uuid"`
	Balance  int64  `json:"balance" example:"0"`
} // @name WalletResponse

type WalletListResponse struct {
	Wallets []WalletResponse `json:"wallets"`
} // @name WalletListResponse

type ErrorResponse struct {
	Message string `json:"message" example:"Invalid request body"`
} // @name ErrorResponse
