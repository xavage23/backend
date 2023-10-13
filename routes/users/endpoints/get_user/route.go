package get_user

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
)

var (
	userColsArr = db.GetCols(types.User{})
	userCols    = strings.Join(userColsArr, ", ")
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get User",
		Description: "Returns a user object given a user id",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Resp: types.User{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	row, err := state.Pool.Query(d.Context, "SELECT "+userCols+" FROM users WHERE id = $1", userId)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.DefaultResponse(http.StatusNotFound)
	}

	user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.User])

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: user,
	}
}
