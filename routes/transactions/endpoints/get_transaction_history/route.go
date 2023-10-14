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
		Description: "Returns the transaction history.",
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
				Name:        "only_me",
				In:          "query",
				Description: "Whether to only include transactions that the user is involved in. ",
				Required:    false,
				Schema:      docs.IdSchema,
			},
		},
		Resp: []types.UserTransaction{},
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

	var uts []types.UserTransaction
	var err error

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if r.URL.Query().Get("only_me") == "true" {
		uts, err = transact.GetUserTransactions(d.Context, userId, gameId, currentPriceIndex)
	} else {
		uts, err = transact.GetAllTransactions(d.Context, gameId, currentPriceIndex)
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

	return uapi.HttpResponse{
		Json: uts,
	}
}
