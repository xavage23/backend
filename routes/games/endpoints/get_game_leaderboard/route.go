package get_game_leaderboard

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"xavagebb/db"
	"xavagebb/state"
	"xavagebb/transact"
	"xavagebb/types"

	"github.com/go-chi/chi/v5"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	userColsArr = db.GetCols(types.User{})
	userCols    = strings.Join(userColsArr, ", ")
)

type userData struct {
	user             *types.User
	initialBalance   int64
	currentBalance   int64
	partialPortfolio map[string]*types.Portfolio
}

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Game Leaderboard",
		Description: "Returns the game leaderboard as a map of the user id to the leaderboard user in random order.",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "gameId",
				In:          "path",
				Description: "The ID of the game",
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
		Resp: map[string]types.Leaderboard{},
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

	// Check cache
	if cachedData, err := state.Redis.Get(d.Context, "game_leaderboard:"+gameId).Result(); err == nil && len(cachedData) > 0 {
		return uapi.HttpResponse{
			Data:    cachedData,
			Headers: map[string]string{"X-Cached": "true"},
		}
	}

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error("Failed to fetch current price index", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var uts []types.UserTransaction

	uts, err = transact.GetAllTransactions(d.Context, gameId)

	if err != nil {
		state.Logger.Error("failed to get all transactions", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var userMap = map[string]*userData{}
	for i := range uts {
		data, ok := userMap[uts[i].UserID]

		if !ok {
			row, err := state.Pool.Query(d.Context, "SELECT "+userCols+" FROM users WHERE id = $1", uts[i].UserID)

			if err != nil {
				state.Logger.Error("failed to fetch user [db fetch]", zap.Error(err), zap.String("userId", uts[i].UserID))
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.User])

			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}

			if err != nil {
				state.Logger.Error("failed to fetch user [collect]", zap.Error(err), zap.String("userId", uts[i].UserID))
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			// Fetch the initial balance of the current game. Other games don't matter here and is a chance of improvement
			var initialBalance int64

			err = state.Pool.QueryRow(d.Context, "SELECT initial_balance FROM game_users WHERE user_id = $1 AND game_id = $2", uts[i].UserID, gameId).Scan(&initialBalance)

			if err != nil {
				state.Logger.Error("failed to fetch initial balance", zap.Error(err), zap.String("userId", uts[i].UserID), zap.String("gameId", gameId))
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			userMap[uts[i].UserID] = &userData{
				user:             &user,
				initialBalance:   initialBalance,
				currentBalance:   initialBalance,
				partialPortfolio: map[string]*types.Portfolio{},
			}

			data = userMap[uts[i].UserID]
		}

		if _, ok := data.partialPortfolio[uts[i].StockID]; !ok {
			stock, err := transact.GetStock(d.Context, uts[i].StockID, currentPriceIndex)

			if err != nil {
				if err != nil {
					state.Logger.Error("Failed to get stock", zap.Error(err), zap.String("gameId", gameId), zap.String("stockId", uts[i].StockID))
					return uapi.DefaultResponse(http.StatusInternalServerError)
				}
			}

			userMap[uts[i].UserID].partialPortfolio[uts[i].StockID] = &types.Portfolio{
				Stock:   stock,
				Amounts: map[int64]types.PortfolioAmount{},
			}
		}

		pa, ok := data.partialPortfolio[uts[i].StockID].Amounts[uts[i].SalePrice]

		if !ok {
			pa = types.PortfolioAmount{}
			data.partialPortfolio[uts[i].StockID].Amounts[uts[i].SalePrice] = pa
		}

		switch uts[i].Action {
		case "buy":
			pa.Amount += uts[i].Amount
			data.currentBalance -= uts[i].SalePrice * uts[i].Amount
		case "sell":
			pa.Amount -= uts[i].Amount
			data.currentBalance += uts[i].SalePrice * uts[i].Amount
		}

		data.partialPortfolio[uts[i].StockID].Amounts[uts[i].SalePrice] = pa

		// Update the user map
		userMap[uts[i].UserID] = data
	}

	var leaderboardUsers = map[string]*types.Leaderboard{}

	for k, v := range userMap {
		// Use the partialPortfolio to calculate the current value of the portfolio if it was all sold at current price
		var portfolioValue int64

		for _, v := range v.partialPortfolio {
			for _, amt := range v.Amounts {
				portfolioValue += amt.Amount * v.Stock.CurrentPrice
			}
		}

		leaderboardUsers[k] = &types.Leaderboard{
			User:           v.user,
			InitialBalance: v.initialBalance,
			CurrentBalance: v.currentBalance,
			PortfolioValue: portfolioValue,
		}
	}

	return uapi.HttpResponse{
		Json:      leaderboardUsers,
		CacheKey:  "game_leaderboard:" + gameId,
		CacheTime: 1 * time.Minute,
	}
}
