package wallet_api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"test-wallet/internal/httpapi"
	wallet_api "test-wallet/internal/httpapi/wallet"
	"test-wallet/internal/wallet"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	id      uuid.UUID
	balance int64
	err     error
	called  bool
	delta   int64
	limit   int
	offset  int
}

func (r *repositoryStub) GetBalance(_ context.Context, id uuid.UUID) (int64, error) {
	r.called = true
	if id != r.id {
		return 0, wallet.ErrWalletNotFound
	}
	return r.balance, r.err
}

func (r *repositoryStub) ChangeBalance(_ context.Context, id uuid.UUID, delta int64) (int64, error) {
	r.delta = delta
	return r.GetBalance(context.Background(), id)
}

func (r *repositoryStub) Create(_ context.Context, id uuid.UUID) (wallet.Wallet, error) {
	r.called = true
	return wallet.Wallet{ID: id}, r.err
}

func (r *repositoryStub) List(_ context.Context, limit, offset int) ([]wallet.Wallet, error) {
	r.called, r.limit, r.offset = true, limit, offset
	return []wallet.Wallet{}, r.err
}

var walletHTTPTestID = uuid.New()

func walletHTTPTestBody(operation string, amount string) string {
	return fmt.Sprintf(`{"walletId":%q,"operationType":%q,"amount":%s}`, walletHTTPTestID.String(), operation, amount)
}

var getBalanceTests = []struct {
	name, path string
	repoErr    error
	status     int
	message    string
	called     bool
}{
	{
		name:   "get",
		path:   "/wallets/" + walletHTTPTestID.String(),
		status: 200,
		called: true,
	},
	{
		name:   "bad path UUID",
		path:   "/wallets/bad",
		status: 400,
	},
	{
		name:    "missing wallet",
		path:    "/wallets/" + walletHTTPTestID.String(),
		repoErr: fmt.Errorf("repository: %w", wallet.ErrWalletNotFound),
		status:  404,
		called:  true,
	},
	{
		name:    "get internal error",
		path:    "/wallets/" + walletHTTPTestID.String(),
		repoErr: errors.New("secret database details"),
		status:  500,
		message: "Internal server error",
		called:  true,
	},
}

func TestGetBalance(t *testing.T) {
	for _, tt := range getBalanceTests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{
				id:      walletHTTPTestID,
				balance: 123,
				err:     tt.repoErr,
			}
			rec := serveWalletRequest(repo, http.MethodGet, tt.path, "")

			require.Equal(t, tt.status, rec.Code, "body=%s", rec.Body.String())
			require.Equal(t, tt.called, repo.called)
			if tt.status >= 400 {
				requireErrorResponse(t, rec, tt.message)
				return
			}

			require.JSONEq(t,
				fmt.Sprintf(`{"walletId":%q,"balance":123}`, walletHTTPTestID.String()),
				rec.Body.String(),
			)
		})
	}
}

var updateBalanceTests = []struct {
	name, body string
	repoErr    error
	status     int
	message    string
	called     bool
	delta      int64
}{
	{
		name:   "deposit",
		body:   walletHTTPTestBody("DEPOSIT", "100"),
		status: 200,
		called: true,
		delta:  100,
	},
	{
		name:   "withdraw",
		body:   walletHTTPTestBody("WITHDRAW", "100"),
		status: 200,
		called: true,
		delta:  -100,
	},
	{
		name:   "malformed JSON",
		body:   "{",
		status: 400,
	},
	{
		name:   "fractional amount",
		body:   walletHTTPTestBody("DEPOSIT", "0.5"),
		status: 400,
	},
	{
		name:   "empty body",
		status: 400,
	},
	{
		name:   "invalid body UUID",
		body:   `{"walletId":"bad","operationType":"DEPOSIT","amount":1}`,
		status: 400,
	},
	{
		name:   "zero amount",
		body:   walletHTTPTestBody("DEPOSIT", "0"),
		status: 400,
	},
	{
		name:   "negative amount",
		body:   walletHTTPTestBody("WITHDRAW", "-1"),
		status: 400,
	},
	{
		name:   "invalid operation",
		body:   walletHTTPTestBody("TRANSFER", "1"),
		status: 400,
	},
	{
		name:    "deposit missing wallet",
		body:    walletHTTPTestBody("DEPOSIT", "1"),
		repoErr: wallet.ErrWalletNotFound,
		status:  404,
		called:  true,
		delta:   1,
	},
	{
		name:    "insufficient funds",
		body:    walletHTTPTestBody("WITHDRAW", "100"),
		repoErr: fmt.Errorf("repository: %w", wallet.ErrInsufficientFunds),
		status:  409,
		called:  true,
		delta:   -100,
	},
	{
		name:    "overflow",
		body:    walletHTTPTestBody("DEPOSIT", "100"),
		repoErr: wallet.ErrBalanceOverflow,
		status:  409,
		called:  true,
		delta:   100,
	},
	{
		name:    "internal error",
		body:    walletHTTPTestBody("DEPOSIT", "100"),
		repoErr: errors.New("database details"),
		status:  500,
		message: "Internal server error",
		called:  true,
		delta:   100,
	},
}

func TestUpdateBalance(t *testing.T) {
	for _, tt := range updateBalanceTests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{
				id:      walletHTTPTestID,
				balance: 123,
				err:     tt.repoErr,
			}
			rec := serveWalletRequest(repo, http.MethodPost, "/wallets", tt.body)

			require.Equal(t, tt.status, rec.Code, "body=%s", rec.Body.String())
			require.Equal(t, tt.called, repo.called)
			require.Equal(t, tt.delta, repo.delta)
			if tt.status >= 400 {
				requireErrorResponse(t, rec, tt.message)
				return
			}

			require.JSONEq(t,
				fmt.Sprintf(`{"walletId":%q,"balance":123}`, walletHTTPTestID.String()),
				rec.Body.String(),
			)
		})
	}
}

var createWalletTests = []struct {
	name    string
	repoErr error
	status  int
	message string
}{
	{
		name:   "create",
		status: 201,
	},
	{
		name:    "create internal error",
		repoErr: errors.New("database details"),
		status:  500,
		message: "Internal server error",
	},
}

func TestCreateWallet(t *testing.T) {
	for _, tt := range createWalletTests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{err: tt.repoErr}
			rec := serveWalletRequest(repo, http.MethodPost, "/wallets/create", "")

			require.Equal(t, tt.status, rec.Code, "body=%s", rec.Body.String())
			require.True(t, repo.called)
			if tt.status >= 400 {
				requireErrorResponse(t, rec, tt.message)
				return
			}

			var result struct {
				WalletID uuid.UUID `json:"walletId"`
				Balance  int64     `json:"balance"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
			require.NotEqual(t, uuid.Nil, result.WalletID)
			require.Zero(t, result.Balance)
		})
	}
}

var listWalletsTests = []struct {
	name, path string
	repoErr    error
	status     int
	message    string
	called     bool
}{
	{
		name:   "empty list",
		path:   "/wallets",
		status: 200,
		called: true,
	},
	{
		name:    "list internal error",
		path:    "/wallets",
		repoErr: errors.New("secret database details"),
		status:  500,
		message: "Internal server error",
		called:  true,
	},
	{
		name:   "list limit text",
		path:   "/wallets?limit=abc",
		status: 400,
	},
	{
		name:   "list offset text",
		path:   "/wallets?offset=abc",
		status: 400,
	},
	{
		name:   "list invalid limit",
		path:   "/wallets?limit=101",
		status: 400,
	},
	{
		name:   "list negative offset",
		path:   "/wallets?offset=-1",
		status: 400,
	},
}

func TestListWallets(t *testing.T) {
	for _, tt := range listWalletsTests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{err: tt.repoErr}
			rec := serveWalletRequest(repo, http.MethodGet, tt.path, "")

			require.Equal(t, tt.status, rec.Code, "body=%s", rec.Body.String())
			require.Equal(t, tt.called, repo.called)
			if tt.status >= 400 {
				requireErrorResponse(t, rec, tt.message)
				return
			}

			require.JSONEq(t, `{"wallets":[]}`, rec.Body.String())
			require.Equal(t, 20, repo.limit)
			require.Zero(t, repo.offset)
		})
	}
}

func serveWalletRequest(
	repo *repositoryStub,
	method,
	path,
	body string,
) *httptest.ResponseRecorder {
	handler := wallet_api.NewWalletHandler(wallet.NewService(repo))
	router := httpapi.NewRouter(wallet_api.NewWalletRouter(handler))
	router.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	if method == http.MethodPost && body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func requireErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, message string) {
	t.Helper()

	var result struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.NotEmpty(t, result.Message)
	if message != "" {
		require.Equal(t, message, result.Message)
	}
}
