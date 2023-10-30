package get_stock_list

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
		Summary:     "Get Stock List",
		Description: "Gets the list of stocks with their current snapshot prices.",
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
			{
				Name:        "with_prior_prices",
				In:          "query",
				Description: "Whether to include the prior prices of stocks. If using this, a stock_id must be provided.",
				Required:    false,
				Schema:      docs.IdSchema,
			},
		},
		Resp: types.StockList{},
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

	includes := []string{}

	if r.URL.Query().Get("with_prior_prices") == "true" {
		includes = append(includes, "prior_prices")
	}

	rows, err := state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE game_id = $1 ORDER BY created_at DESC", gameId)

	if err != nil {
		state.Logger.Error("Failed to fetch stocks [db fetch]", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	stocks, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Stock])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Json: []types.Stock{},
		}
	}

	if err != nil {
		state.Logger.Error("Failed to fetch stock [collect]", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error("Failed to fetch current price index", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var stockList []*types.Stock
	for i := range stocks {
		parsedStock := transact.ParseStock(d.Context, &stocks[i], currentPriceIndex)

		parsedStock, err = transact.FillStock(d.Context, parsedStock, currentPriceIndex, includes)

		if err != nil {
			state.Logger.Error("Failed to fill stock", zap.Error(err), zap.String("gameId", gameId))
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}

		stockList = append(stockList, parsedStock)
	}

	return uapi.HttpResponse{
		Json: types.StockList{
			Stocks:     stockList,
			PriceIndex: currentPriceIndex,
		},
	}
}
