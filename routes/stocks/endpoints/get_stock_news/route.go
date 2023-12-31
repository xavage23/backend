package get_stock_news

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
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

var (
	newsColsArr = db.GetCols(types.News{})
	newsCols    = strings.Join(newsColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Stock News",
		Description: "Gets the list of news of the stocks with their current snapshot prices.",
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
				Name:        "with_stocks",
				In:          "query",
				Description: "Whether to include stock objects in the response.",
				Required:    false,
				Schema:      docs.IdSchema,
			},
		},
		Resp: []types.News{},
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

	gameEnabledAt, ok := d.Auth.Data["gameEnabledAt"].(pgtype.Timestamptz)

	if !ok {
		state.Logger.Error("gameEnabledAt not found in auth data", zap.Any("data", d.Auth.Data))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	withStocks := r.URL.Query().Get("with_stocks")

	rows, err := state.Pool.Query(d.Context, "SELECT "+newsCols+" FROM news WHERE game_id = $1 AND published = true AND NOW() - $2 > show_at ORDER BY show_at DESC", gameId, gameEnabledAt)

	if err != nil {
		state.Logger.Error("Failed to fetch news [db fetch]", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	news, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.News])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Json: []types.News{},
		}
	}

	if err != nil {
		state.Logger.Error("Failed to fetch news [collect]", zap.Error(err), zap.String("gameId", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if withStocks == "true" {
		currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

		if err != nil {
			state.Logger.Error("Failed to fetch news [get current price index]", zap.Error(err), zap.String("gameId", gameId))
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}

		var cachedStocks = make(map[[16]byte]*types.Stock)
		for i := range news {
			news[i].ShowAtParsed = int64(news[i].ShowAt.Microseconds/(1000000)) + int64(news[i].ShowAt.Days*86400) + int64(news[i].ShowAt.Months*2592000)

			if !news[i].AffectedStockID.Valid {
				continue
			}

			cachedStock, ok := cachedStocks[news[i].AffectedStockID.Bytes]

			if ok {
				news[i].AffectedStock = cachedStock
			} else {
				stock, err := transact.GetStock(d.Context, transact.ConvertUUIDToString(news[i].AffectedStockID.Bytes), currentPriceIndex)

				if err != nil {
					state.Logger.Error("Failed to fetch stock", zap.Error(err), zap.String("gameId", gameId), zap.String("stockId", transact.ConvertUUIDToString(news[i].AffectedStockID.Bytes)))
					return uapi.DefaultResponse(http.StatusInternalServerError)
				}

				stock, err = transact.FillStock(d.Context, stock, currentPriceIndex, []string{"prior_prices"})

				if err != nil {
					state.Logger.Error("Failed to fill prior prices", zap.Error(err), zap.String("gameId", gameId), zap.String("stockId", transact.ConvertUUIDToString(news[i].AffectedStockID.Bytes)))
					return uapi.DefaultResponse(http.StatusInternalServerError)
				}

				news[i].AffectedStock = stock
				cachedStocks[news[i].AffectedStockID.Bytes] = stock
			}
		}
	}

	return uapi.HttpResponse{
		Json: news,
	}
}
