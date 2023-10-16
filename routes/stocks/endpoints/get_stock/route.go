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
				Description: "The stock ID.",
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
		state.Logger.Error("gameId not found in auth data", d.Auth.Data)
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

	rows, err := state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE id = $1 AND (game_id = $2 OR ticker = $2) ORDER BY created_at DESC", stockId, gameId)

	if err != nil {
		state.Logger.Error(err)
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
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if stock.GameID != gameId {
		return uapi.HttpResponse{
			Status: http.StatusBadRequest,
			Json: types.ApiError{
				Message: "This stock is not available in this game",
			},
		}
	}

	included := []string{"prior_prices"}

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	parsedStock := transact.ParseStock(d.Context, &stock, currentPriceIndex)

	pp, err := transact.GetPriorStockPrices(d.Context, gameId, parsedStock.Ticker)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	parsedStock.PriorPrices = pp
	parsedStock.Includes = included

	return uapi.HttpResponse{
		Json: parsedStock,
	}
}
