// Binds onto eureka uapi
package api

import (
	"errors"
	"net/http"
	"xavagebb/constants"
	"xavagebb/state"
	"xavagebb/types"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/exp/slices"
)

const (
	TargetTypeUser = "user"
)

type ErrorStructGen struct{}

func (e ErrorStructGen) New(err string, ctx map[string]string) any {
	return types.ApiError{
		Message: err,
		Context: ctx,
	}
}

// Authorizes a request
func Authorize(r uapi.Route, req *http.Request) (uapi.AuthData, uapi.HttpResponse, bool) {
	authHeader := req.Header.Get("Authorization")

	if len(r.Auth) > 0 && authHeader == "" && !r.AuthOptional {
		return uapi.AuthData{}, uapi.DefaultResponse(http.StatusUnauthorized), false
	}

	authData := uapi.AuthData{}

	for _, auth := range r.Auth {
		// There are two cases, one with a URLVar (such as /bots/stats) and one without

		if authData.Authorized {
			break
		}

		if authHeader == "" {
			continue
		}

		var urlIds []string

		switch auth.Type {
		case TargetTypeUser:
			// Check if the user exists with said API token only
			var id pgtype.Text
			var enabled bool

			err := state.Pool.QueryRow(state.Context, "SELECT user_id, enabled FROM users WHERE token = $1", authHeader, "User ").Scan(&id, &enabled)

			if err != nil {
				continue
			}

			if !id.Valid {
				continue
			}

			if req.Header.Get("X-GameUser-ID") != "" {
				var count int64

				err = state.Pool.QueryRow(state.Context, "SELECT COUNT(*) FROM game_users WHERE id = $1", req.Header.Get("X-GameUser-ID")).Scan(&count)

				if errors.Is(err, pgx.ErrNoRows) {
					return uapi.AuthData{}, uapi.HttpResponse{
						Status: http.StatusForbidden,
						Json:   types.ApiError{Message: "This game user ID does not exist [no count rows]!"},
					}, false
				}

				if err != nil {
					return uapi.AuthData{}, uapi.HttpResponse{
						Status: http.StatusForbidden,
						Json:   types.ApiError{Message: "Failed to fetch selected game: " + err.Error()},
					}, false
				}

				if count == 0 {
					return uapi.AuthData{}, uapi.HttpResponse{
						Status: http.StatusForbidden,
						Json:   types.ApiError{Message: "This game user ID does not exist [count = 0]!"},
					}, false
				}
			} else {
				if auth.AllowedScope != "notingame" {
					return uapi.AuthData{}, uapi.HttpResponse{
						Status: http.StatusForbidden,
						Json:   types.ApiError{Message: "You must specify a game user ID to use this endpoint!"},
					}, false
				}
			}

			authData = uapi.AuthData{
				TargetType: TargetTypeUser,
				ID:         id.String,
				Authorized: true,
				Banned:     enabled,
			}
			urlIds = []string{id.String}
		}

		// Now handle the URLVar
		if auth.URLVar != "" {
			state.Logger.Info("URLVar: ", auth.URLVar)
			gotUserId := chi.URLParam(req, auth.URLVar)
			if !slices.Contains(urlIds, gotUserId) {
				authData = uapi.AuthData{} // Remove auth data
			}
		}

		// Banned users cannot use the API at all otherwise if not explicitly scoped to "ban_exempt"
		if authData.Banned && auth.AllowedScope != "ban_exempt" {
			return uapi.AuthData{}, uapi.HttpResponse{
				Status: http.StatusForbidden,
				Json:   types.ApiError{Message: "This user account is not enabled yet!"},
			}, false
		}
	}

	if len(r.Auth) > 0 && !authData.Authorized && !r.AuthOptional {
		return uapi.AuthData{}, uapi.DefaultResponse(http.StatusUnauthorized), false
	}

	return authData, uapi.HttpResponse{}, true
}

func Setup() {
	uapi.SetupState(uapi.UAPIState{
		Logger:    state.Logger,
		Authorize: Authorize,
		AuthTypeMap: map[string]string{
			TargetTypeUser: "user",
		},
		Redis:   state.Redis,
		Context: state.Context,
		Constants: &uapi.UAPIConstants{
			NotFound:         constants.NotFound,
			NotFoundPage:     constants.NotFoundPage,
			BadRequest:       constants.BadRequest,
			Forbidden:        constants.Forbidden,
			Unauthorized:     constants.Unauthorized,
			InternalError:    constants.InternalError,
			MethodNotAllowed: constants.MethodNotAllowed,
			BodyRequired:     constants.BodyRequired,
		},
		UAPIErrorType: ErrorStructGen{},
	})
}
