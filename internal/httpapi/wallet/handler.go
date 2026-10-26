package wallet_api

import (
	"errors"
	"net/http"

	"test-wallet/internal/wallet"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type WalletHandler struct {
	service *wallet.Service
}

func NewWalletHandler(service *wallet.Service) *WalletHandler {
	return &WalletHandler{
		service: service,
	}
}

// @Summary Create wallet
// @Tags wallets
// @Produce json
// @Success 201 {object} WalletResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallets/create [post]
func (h *WalletHandler) CreateWallet(c *echo.Context) error {
	result, err := h.service.CreateWallet(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		).Wrap(err)
	}

	return c.JSON(http.StatusCreated, WalletResponse{
		WalletID: result.ID.String(),
		Balance:  result.Balance,
	})
}

// @Summary Get list of wallets
// @Tags wallets
// @Produce json
// @Param limit query int false " " default(20)
// @Param offset query int false " " default(0)
// @Success 200 {object} WalletListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallets [get]
func (h *WalletHandler) ListWallets(c *echo.Context) error {
	limit := 20
	offset := 0

	err := echo.QueryParamsBinder(c).
		Int("limit", &limit).
		Int("offset", &offset).
		BindError()

	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"limit and offset must be integers",
		)
	}

	items, err := h.service.ListWallets(
		c.Request().Context(),
		limit,
		offset,
	)
	if errors.Is(err, wallet.ErrInvalidPagination) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			err.Error(),
		)
	}

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		).Wrap(err)
	}

	response := make([]WalletResponse, 0, len(items))

	for _, item := range items {
		response = append(response, WalletResponse{
			WalletID: item.ID.String(),
			Balance:  item.Balance,
		})
	}

	return c.JSON(http.StatusOK, WalletListResponse{
		Wallets: response,
	})
}

// @Summary Update wallet balance
// @Tags wallets
// @Accept json
// @Produce json
// @Param request body UpdateBalanceRequest true " "
// @Success 200 {object} WalletResponse
// @Failure 400 {object} ErrorResponse "Incorrect request"
// @Failure 404 {object} ErrorResponse "Wallet not found"
// @Failure 409 {object} ErrorResponse "Insufficient funds or balance limit exceeded"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/wallet [post]
func (h *WalletHandler) UpdateBalance(c *echo.Context) error {
	var req UpdateBalanceRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Invalid request body",
		)
	}

	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"walletId must be a valid UUID",
		)
	}

	balance, err := h.service.UpdateBalance(
		c.Request().Context(),
		walletID,
		wallet.OperationType(req.OperationType),
		req.Amount,
	)
	if err != nil {
		return walletHTTPError(c, err)
	}

	return c.JSON(http.StatusOK, WalletResponse{
		WalletID: walletID.String(),
		Balance:  balance,
	})
}

// @Summary Get wallet balance
// @Tags wallets
// @Produce json
// @Param walletID path string true "UUID кошелька" Format(uuid)
// @Success 200 {object} WalletResponse
// @Failure 400 {object} ErrorResponse "Invalid UUID"
// @Failure 404 {object} ErrorResponse "Wallet not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/wallets/{walletID} [get]
func (h *WalletHandler) GetBalance(c *echo.Context) error {
	walletID, err := uuid.Parse(c.Param("walletID"))
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"walletID must be a valid UUID",
		)
	}

	balance, err := h.service.GetBalance(c.Request().Context(), walletID)
	if err != nil {
		return walletHTTPError(c, err)
	}

	return c.JSON(http.StatusOK, WalletResponse{
		WalletID: walletID.String(),
		Balance:  balance,
	})
}

func walletHTTPError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, wallet.ErrInvalidAmount):
		return echo.NewHTTPError(http.StatusBadRequest, wallet.ErrInvalidAmount.Error())
	case errors.Is(err, wallet.ErrInvalidOperation):
		return echo.NewHTTPError(http.StatusBadRequest, wallet.ErrInvalidOperation.Error())
	case errors.Is(err, wallet.ErrWalletNotFound):
		return echo.NewHTTPError(http.StatusNotFound, wallet.ErrWalletNotFound.Error())
	case errors.Is(err, wallet.ErrInsufficientFunds):
		return echo.NewHTTPError(http.StatusConflict, wallet.ErrInsufficientFunds.Error())
	case errors.Is(err, wallet.ErrBalanceOverflow):
		return echo.NewHTTPError(http.StatusConflict, wallet.ErrBalanceOverflow.Error())
	default:
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Internal server error",
		).Wrap(err)
	}
}
