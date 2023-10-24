package get_game_leaderboard

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

	uts, err = transact.GetAllTransactions(d.Context, gameId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var leaderboardUsers = map[string]*types.Leaderboard{}

	// Fill in user
	for i := range uts {
		_, ok := leaderboardUsers[uts[i].UserID]

		if !ok {
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

			// Fetch the initial balance of the current game. Other games don't matter here and is a chance of improvement
			var initialBalance int64

			err = state.Pool.QueryRow(d.Context, "SELECT initial_balance FROM game_users WHERE user_id = $1 AND game_id = $2", uts[i].UserID, gameId).Scan(&initialBalance)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			leaderboardUsers[uts[i].UserID] = &types.Leaderboard{
				User:           &user,
				InitialBalance: initialBalance,
				CurrentBalance: initialBalance,
			}
		}

		switch uts[i].Action {
		case "buy":
			if uts[i].Amount < 0 {
				leaderboardUsers[uts[i].UserID].ShortAmount += -1 * uts[i].Amount * uts[i].SalePrice
			}

			leaderboardUsers[uts[i].UserID].CurrentBalance -= uts[i].SalePrice * uts[i].Amount
		case "sell":
			if uts[i].Amount < 0 {
				leaderboardUsers[uts[i].UserID].ShortAmount += -1 * uts[i].Amount * uts[i].SalePrice
			}

			leaderboardUsers[uts[i].UserID].CurrentBalance += uts[i].SalePrice * uts[i].Amount
		}
	}

	return uapi.HttpResponse{
		Json: leaderboardUsers,
	}
}
