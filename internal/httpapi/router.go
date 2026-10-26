package httpapi

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	echoSwagger "github.com/swaggo/echo-swagger/v2"

	_ "test-wallet/docs"

	"test-wallet/internal/httpapi/wallet"
)

func NewRouter(walletRouter *wallet_api.WalletRouter) *echo.Echo {
	router := echo.New()

	router.Use(requestLogger())
	router.Use(middleware.Recover())

	router.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "OK",
		})
	})

	walletRouter.Mount(router.Group("/api/v1"))

	router.GET("/swagger/*", echoSwagger.WrapHandler)

	return router
}
