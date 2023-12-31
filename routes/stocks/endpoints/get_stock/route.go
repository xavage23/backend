package get_stock

import (
	"errors"
	"net/http"
	"strings"

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
	stocksColsArr = db.GetCols(types.Stock{})
	stocksCols    = strings.Join(stocksColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Stock",
		Description: "Gets a stock based on its ID returning *all* fields.",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "stockId",
				In:          "path",
				Description: "The stock ID or ticker",
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
		Resp: types.StockList{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")
	stockId := chi.URLParam(r, "stockId")
	gameId, ok := d.Auth.Data["gameId"].(string)

	if !ok {
		state.Logger.Error("gameId not found in auth data", zap.Any("data", d.Auth.Data))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if stockId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	var rows pgx.Rows
	var err error

	if transact.IsValidUUID(stockId) {
		rows, err = state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE id = $1 AND game_id = $2 ORDER BY created_at DESC", stockId, gameId)
	} else {
		rows, err = state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE ticker = $1 AND game_id = $2 ORDER BY created_at DESC", stockId, gameId)
	}

	if err != nil {
		state.Logger.Error("Error fetching stocks [db query]", zap.Error(err), zap.String("stockId", stockId), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	stock, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Stock])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json: types.ApiError{
				Message: "Stock not found",
			},
		}
	}

	if err != nil {
		state.Logger.Error("Error fetching stocks [collect one row]", zap.Error(err), zap.String("stockId", stockId), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error("Error fetching current price index", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	parsedStock := transact.ParseStock(d.Context, &stock, currentPriceIndex)

	parsedStock, err = transact.FillStock(d.Context, parsedStock, currentPriceIndex, []string{"prior_prices", "known_ratios", "prior_ratios"})

	if err != nil {
		state.Logger.Error("Error parsing stock", zap.Error(err), zap.String("gameId", gameId), zap.String("stockId", stockId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: parsedStock,
	}
}
