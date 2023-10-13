package join_game

import (
	"errors"
	"net/http"
	"time"
	"xavagebb/state"

	"xavagebb/types"

	"github.com/go-playground/validator/v10"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/infinitybotlist/eureka/uapi/ratelimit"
	"github.com/jackc/pgx/v5"
)

var (
	compiledMessages = uapi.CompileValidationErrors(types.GameJoinRequest{})
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Login User",
		Description: "Takes in a ``code`` query parameter and returns a user ``token``. **Cannot be used outside of the site for security reasons but documented in case we wish to allow its use in the future.**",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Req:  types.GameJoinRequest{},
		Resp: types.GameJoinResponse{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	limit, err := ratelimit.Ratelimit{
		Expiry:      1 * time.Second,
		MaxRequests: 10,
		Bucket:      "login",
	}.Limit(d.Context, r)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if limit.Exceeded {
		return uapi.HttpResponse{
			Json: types.ApiError{
				Message: "You are being ratelimited. Please try again in " + limit.TimeToReset.String(),
			},
			Headers: limit.Headers(),
			Status:  http.StatusTooManyRequests,
		}
	}

	var req types.GameJoinRequest

	hresp, ok := uapi.MarshalReqWithHeaders(r, &req, limit.Headers())

	if !ok {
		return hresp
	}

	// Validate the payload
	err = state.Validator.Struct(req)

	if err != nil {
		errors := err.(validator.ValidationErrors)
		return uapi.ValidatorErrorResponse(compiledMessages, errors)
	}

	// Check that the game exists with the current code and passphrase
	var gameId string
	var initialBalance int

	err = state.Pool.QueryRow(d.Context, "SELECT id, initial_balance FROM games WHERE code = $1 AND passphrase = $2", req.GameCode, req.Passphrase).Scan(&gameId, &initialBalance)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json:   types.ApiError{Message: "Game not found!"},
		}
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	// Check that the user is allowed to join the game
	//
	// 1. Check game_allowed_users to ensure user is still allowed to use the API
	var gacCount int64

	err = state.Pool.QueryRow(state.Context, "SELECT COUNT(*) FROM game_allowed_users WHERE user_id = $1 AND game_id = $2", d.Auth.ID, gameId).Scan(&gacCount)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "This user does not have permission to play this game [count]!"},
		}
	}

	if err != nil {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "Failed to fetch selected game: " + err.Error()},
		}
	}

	if gacCount == 0 {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "This user does not have permission to play this game!"},
		}
	}

	// Check that the user is not already in the game
	var gameUserId string

	err = state.Pool.QueryRow(d.Context, "SELECT id FROM game_user WHERE user_id = $1 AND game_id = $2", d.Auth.ID, gameId).Scan(&gameUserId)

	if err == nil {
		return uapi.HttpResponse{
			Json: types.GameJoinResponse{
				ID:  gameUserId,
				New: true,
			},
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	// Create the game join
	err = state.Pool.QueryRow(d.Context, "INSERT INTO game_user (user_id, game_id, balance) VALUES ($1, $2, $3) RETURNING id", d.Auth.ID, gameId, initialBalance).Scan(&gameUserId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: types.GameJoinResponse{
			ID:  gameUserId,
			New: true,
		},
	}
}
