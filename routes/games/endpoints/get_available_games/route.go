package get_available_games

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"xavagebb/db"
	"xavagebb/state"

	"xavagebb/types"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
)

var (
	gameColsArr = db.GetCols(types.Game{})
	gameCols    = strings.Join(gameColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get Available Games",
		Description: "Returns the list of all available games.",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user.",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Resp: []types.AvailableGame{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	row, err := state.Pool.Query(d.Context, "SELECT "+gameCols+" FROM games WHERE publicly_listed = true")

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	games, err := pgx.CollectRows(row, pgx.RowToStructByName[types.Game])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Json: []types.AvailableGame{},
		}
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var availableGameList = []types.AvailableGame{}

	for _, game := range games {
		var canJoin bool

		var gacCount int64

		err = state.Pool.QueryRow(state.Context, "SELECT COUNT(*) FROM game_allowed_users WHERE user_id = $1 AND game_id = $2", d.Auth.ID, game.ID).Scan(&gacCount)

		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return uapi.HttpResponse{
					Status: http.StatusForbidden,
					Json:   types.ApiError{Message: "Failed to fetch selected game: " + err.Error()},
				}
			}
		}

		if gacCount > 0 {
			canJoin = true
		}

		availableGameList = append(availableGameList, types.AvailableGame{
			Game:      game,
			CanJoin:   canJoin,
			IsEnabled: game.Enabled.Valid && game.Enabled.Time.Before(time.Now()),
		})
	}

	return uapi.HttpResponse{
		Json: availableGameList,
	}
}
