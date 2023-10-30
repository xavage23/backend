package get_game

import (
	"errors"
	"net/http"
	"strings"
	"xavagebb/db"
	"xavagebb/state"

	"xavagebb/types"

	"github.com/go-chi/chi/v5"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	gameColsArr = db.GetCols(types.Game{})
	gameCols    = strings.Join(gameColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Game",
		Description: "Returns a game object given a game id",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user. Only used for authentication and ratelimiting purposes",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "gameId",
				In:          "path",
				Description: "The ID or code of the game",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Resp: types.Game{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	gameId := chi.URLParam(r, "gameId")

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	row, err := state.Pool.Query(d.Context, "SELECT "+gameCols+" FROM games WHERE id = $1", gameId)

	if err != nil {
		state.Logger.Error("Failed to fetch game [db fetch]", zap.Error(err), zap.String("game_id", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	game, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.Game])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.DefaultResponse(http.StatusNotFound)
	}

	if err != nil {
		state.Logger.Error("Failed to fetch game [collect]", zap.Error(err), zap.String("game_id", gameId))
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: game,
	}
}
