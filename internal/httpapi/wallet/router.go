package wallet_api

import "github.com/labstack/echo/v5"

type WalletRouter struct {
	handler *WalletHandler
}

func NewWalletRouter(handler *WalletHandler) *WalletRouter {
	return &WalletRouter{
		handler: handler,
	}
}

func (r *WalletRouter) Mount(group *echo.Group) {
	wallets := group.Group("/wallets")

	wallets.POST("/create", r.handler.CreateWallet)
	wallets.GET("", r.handler.ListWallets)
	wallets.GET("/:walletID", r.handler.GetBalance)
	wallets.POST("", r.handler.UpdateBalance)
}
