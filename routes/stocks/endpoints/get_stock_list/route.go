package get_stock_list

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
			{
				Name:        "stock_id",
				In:          "query",
				Description: "The ID of the stock to get. If this is not provided, all stocks will be returned.",
			},
		},
		Resp: types.StockList{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")
	gameId, ok := d.Auth.Data["gameId"].(string)

	if !ok {
		state.Logger.Error("gameId not found in auth data", d.Auth.Data)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	withPriorPrices := r.URL.Query().Get("with_prior_prices")
	stockId := r.URL.Query().Get("stock_id")

	if stockId == "" {
		if state.Redis.Exists(d.Context, "stock_list:"+gameId+"?wpp="+withPriorPrices).Val() > 0 {
			data := state.Redis.Get(d.Context, "stock_list:"+gameId+"?wpp="+withPriorPrices).Val()

			if data != "" {
				return uapi.HttpResponse{
					Data: data,
					Headers: map[string]string{
						"X-Cache": "true",
					},
				}
			}
		}
	}

	var rows pgx.Rows
	var err error

	if stockId != "" {
		rows, err = state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE game_id = $1 AND id = $2 ORDER BY created_at DESC", gameId, stockId)
	} else {
		rows, err = state.Pool.Query(d.Context, "SELECT "+stocksCols+" FROM stocks WHERE game_id = $1 ORDER BY created_at DESC", gameId)
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	stocks, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Stock])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Json: []types.Stock{},
		}
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var included []string = []string{}

	if withPriorPrices == "true" {
		included = append(included, "prior_prices")
	}

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var stockList []*types.Stock
	for i := range stocks {
		parsedStock := transact.ParseStock(d.Context, &stocks[i], currentPriceIndex)

		if withPriorPrices == "true" {
			pp, err := transact.GetPriorStockPrices(d.Context, gameId, parsedStock.Ticker)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			parsedStock.PriorPrices = pp
			parsedStock.Includes = included
		}

		stockList = append(stockList, parsedStock)
	}

	return uapi.HttpResponse{
		Json: types.StockList{
			Stocks:     stockList,
			PriceIndex: currentPriceIndex,
		},
		CacheKey: func() string {
			if stockId != "" {
				return ""
			}

			return "stock_list:" + gameId + "?wpp=" + withPriorPrices
		}(),
		CacheTime: 30 * time.Second,
	}
}
