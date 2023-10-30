package create_transaction

import (
	"errors"
	"fmt"
	"net/http"
	"xavagebb/mapofmu"
	"xavagebb/state"
	"xavagebb/transact"
	"xavagebb/types"

	"github.com/go-chi/chi/v5"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var mut = mapofmu.New[string]()

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Create Transaction",
		Description: "Creates a transaction (buy/sell) for a stock",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "X-GameUser-ID",
				In:          "header",
				Description: "The ID of the game user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Req:  types.CreateTransaction{},
		Resp: types.ApiError{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")
	gameId, ok := d.Auth.Data["gameId"].(string)

	if !ok {
		state.Logger.Error("gameId not found in auth data", zap.Any("data", d.Auth.Data))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	// Ensure game allows trading first
	var tradingAllowed bool

	err := state.Pool.QueryRow(d.Context, "SELECT trading_allowed FROM games WHERE id = $1", gameId).Scan(&tradingAllowed)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json: types.ApiError{
				Message: "Game not found",
			},
		}
	}

	if err != nil {
		state.Logger.Error("Failed to fetch game trading allowed status", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if !tradingAllowed {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "Trading is not allowed right now",
			},
		}
	}

	l := mut.Lock(userId)
	defer l.Unlock()

	var req types.CreateTransaction

	hresp, ok := uapi.MarshalReq(r, &req)

	if !ok {
		return hresp
	}

	// Update here, this is a quick short-circuit point
	switch req.Action {
	case "buy":
	case "sell":
	default:
		return uapi.HttpResponse{
			Status: http.StatusNotImplemented,
			Json: types.ApiError{
				Message: "Action must be either buy or sell",
			},
		}
	}

	var uts []types.UserTransaction

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error("Failed to fetch current price index", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	uts, err = transact.GetUserTransactions(d.Context, userId, gameId)

	if err != nil {
		state.Logger.Error("Failed to get user transactions", zap.Error(err), zap.String("gameId", gameId), zap.String("userId", userId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	// We still haven't found the stock, fetch it manually
	stock, err := transact.GetStock(d.Context, req.StockID, currentPriceIndex)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json: types.ApiError{
				Message: "Stock not found",
			},
		}
	}

	if err != nil {
		state.Logger.Error("Failed to fetch stock", zap.Error(err), zap.String("gameId", gameId), zap.String("stockId", req.StockID))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if stock.GameID != gameId {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json: types.ApiError{
				Message: "Stock not available in this game",
			},
		}
	}

	// Get initial balance of the user
	var initialBalance int64

	err = state.Pool.QueryRow(d.Context, "SELECT initial_balance FROM game_users WHERE id = $1", r.Header.Get("X-GameUser-ID")).Scan(&initialBalance)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "Game user not found",
			},
		}
	}

	if err != nil {
		state.Logger.Error("Failed to fetch game user", zap.Error(err), zap.String("gameUserId", r.Header.Get("X-GameUser-ID")))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	currBalance, _ := transact.GetUserCurrentBalance(initialBalance, uts)

	// Clamp/find best pIndex
	pIndex := currentPriceIndex
	if pIndex > len(stock.Prices)-1 {
		pIndex = len(stock.Prices) - 1
	}

	switch req.Action {
	case "buy":
		totalCost := req.Amount * stock.Prices[pIndex]

		if currBalance < totalCost {
			return uapi.HttpResponse{
				Status: http.StatusBadRequest,
				Json: types.ApiError{
					Message: "Insufficient funds/balance",
				},
			}
		}

		// Create transaction
		_, err = state.Pool.Exec(d.Context, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) VALUES ($1, $2, $2, $3, $4, $5, $6, $7)", userId, gameId, req.StockID, currentPriceIndex, req.Amount, req.Action, stock.Prices[pIndex])

		if err != nil {
			state.Logger.Error("Failed to enact final buy", zap.Error(err), zap.String("gameId", gameId), zap.String("userId", userId))
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}
	case "sell":
		// For sale, check that the user has enough total quantity of the stock
		totalAvailableQuantity := transact.GetTotalStockQuantity(uts, req.StockID)

		if totalAvailableQuantity < req.Amount {
			return uapi.HttpResponse{
				Status: http.StatusBadRequest,
				Json: types.ApiError{
					Message: fmt.Sprint("You are trying to sell ", req.Amount, " stocks but you only have ", totalAvailableQuantity, " available"),
				},
			}
		}

		// Create transaction
		_, err = state.Pool.Exec(d.Context, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) VALUES ($1, $2, $2, $3, $4, $5, $6, $7)", userId, gameId, req.StockID, currentPriceIndex, req.Amount, req.Action, stock.Prices[pIndex])

		if err != nil {
			state.Logger.Error("Failed to enact final sell", zap.Error(err), zap.String("gameId", gameId), zap.String("userId", userId))
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}
	default:
		return uapi.HttpResponse{
			Status: http.StatusNotImplemented,
			Json: types.ApiError{
				Message: "Action must be either buy or sell. To short, use a negative amount using the same buy/sell actions",
			},
		}
	}

	return uapi.DefaultResponse(http.StatusNoContent)
}
