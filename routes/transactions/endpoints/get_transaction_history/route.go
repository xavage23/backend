package get_transaction_history

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
	userColsArr = db.GetCols(types.User{})
	userCols    = strings.Join(userColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Transaction History",
		Description: "Returns the transaction history. Note that user id here is only used for authentication.",
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
				Name:        "include_users",
				In:          "query",
				Description: "Whether to include the user object in each transaction. ",
				Required:    false,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "include_stocks",
				In:          "query",
				Description: "Whether to include the stock object in each transaction. ",
				Required:    false,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "only_me",
				In:          "query",
				Description: "Whether to only include transactions that the user is involved in. ",
				Required:    false,
				Schema:      docs.IdSchema,
			},
		},
		Resp: types.User{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")
	gameId, ok := d.Auth.Data["gameId"].(string)

	if !ok {
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	var uts []types.UserTransaction
	var err error

	if r.URL.Query().Get("only_me") == "true" {
		uts, err = transact.GetUserTransactions(d.Context, userId, gameId)
	} else {
		uts, err = transact.GetAllTransactions(d.Context, gameId)
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	// Fill in user
	if r.URL.Query().Get("include_users") == "true" {
		var users = make(map[string]*types.User)
		for i := range uts {
			cachedUser, ok := users[uts[i].UserID]

			if ok {
				uts[i].User = cachedUser
				continue
			}

			row, err := state.Pool.Query(d.Context, "SELECT "+userCols+" FROM users WHERE id = $1", uts[i].UserID)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.User])

			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			users[uts[i].UserID] = &user
			uts[i].User = &user
		}
	}

	// Fill in stocks
	if r.URL.Query().Get("include_stocks") == "true" {
		var gameCurrentPrice string

		err = state.Pool.QueryRow(d.Context, "SELECT current_price FROM games WHERE id = $1", gameId).Scan(&gameCurrentPrice)

		if err != nil {
			state.Logger.Error(err)
			return uapi.DefaultResponse(http.StatusInternalServerError)
		}

		var cachedStocks = make(map[string]*types.Stock)
		for i := range uts {
			cachedStock, ok := cachedStocks[uts[i].StockID]

			if ok {
				uts[i].Stock = cachedStock
				continue
			}

			stock, err := transact.GetStock(d.Context, uts[i].StockID, gameCurrentPrice)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			cachedStocks[uts[i].StockID] = stock
			uts[i].Stock = stock
		}
	}

	return uapi.HttpResponse{
		Json: uts,
	}
}