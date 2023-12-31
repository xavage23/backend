package get_current_game_user

import (
	"errors"
	"net/http"
	"strings"
	"xavagebb/db"
	"xavagebb/state"
	"xavagebb/transact"

	"xavagebb/types"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	gameUserColsArr = db.GetCols(types.GameUser{})
	gameUserCols    = strings.Join(gameUserColsArr, ", ")

	gameColsArr = db.GetCols(types.Game{})
	gameCols    = strings.Join(gameColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Current Game User",
		Description: "Returns a game user object given a user id. `X-GameUser-ID` must be set to the game user ID.",
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
		Resp: types.GameUser{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	row, err := state.Pool.Query(d.Context, "SELECT "+gameUserCols+" FROM game_users WHERE id = $1", r.Header.Get("X-GameUser-ID"))

	if err != nil {
		state.Logger.Error("Failed to fetch game user [db fetch]", zap.Error(err), zap.String("game_user_id", r.Header.Get("X-GameUser-ID")))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	gu, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.GameUser])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.DefaultResponse(http.StatusNotFound)
	}

	if err != nil {
		state.Logger.Error("Failed to fetch game user [collect]", zap.Error(err), zap.String("game_user_id", r.Header.Get("X-GameUser-ID")))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	gameRow, err := state.Pool.Query(d.Context, "SELECT "+gameCols+" FROM games WHERE id = $1", gu.GameID)

	if err != nil {
		state.Logger.Error("Failed to fetch game [db fetch]", zap.Error(err), zap.String("game_id", gu.GameID))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	game, err := pgx.CollectOneRow(gameRow, pgx.RowToStructByName[types.Game])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.DefaultResponse(http.StatusNotFound)
	}

	if err != nil {
		state.Logger.Error("Failed to fetch game [collect]", zap.Error(err), zap.String("game_id", gu.GameID))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	gu.Game = game

	uts, err := transact.GetUserTransactions(d.Context, gu.UserID, gu.GameID)

	if err != nil {
		state.Logger.Error("Failed to fetch user transactions", zap.Error(err), zap.String("game_user_id", r.Header.Get("X-GameUser-ID")), zap.String("user_id", gu.UserID), zap.String("game_id", gu.GameID))
		return uapi.HttpResponse{
			Status: http.StatusInternalServerError,
			Json:   types.ApiError{Message: "An error occurred while fetching user transactions: " + err.Error()},
		}
	}

	gu.CurrentBalance = transact.GetUserCurrentBalance(gu.InitialBalance, uts)

	return uapi.HttpResponse{
		Json: gu,
	}
}
