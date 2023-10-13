package get_current_game_user

import (
	"errors"
	"net/http"
	"strings"
	"xavagebb/db"
	"xavagebb/state"

	"xavagebb/types"

	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
)

var (
	gameUserColsArr = db.GetCols(types.GameUser{})
	gameUserCols    = strings.Join(gameUserColsArr, ", ")
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
	row, err := state.Pool.Query(d.Context, "SELECT "+gameUserCols+" FROM game_user WHERE id = $1", r.Header.Get("X-GameUser-ID"))

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	gu, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.GameUser])

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.DefaultResponse(http.StatusNotFound)
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: gu,
	}
}
