package login_user

import (
	"errors"
	"net/http"
	"time"
	"xavagebb/state"

	"xavagebb/types"

	"github.com/alexedwards/argon2id"
	"github.com/go-playground/validator/v10"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/infinitybotlist/eureka/uapi/ratelimit"
	"github.com/jackc/pgx/v5"
)

var (
	compiledMessages = uapi.CompileValidationErrors(types.UserLogin{})
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Login User",
		Description: "Takes in a ``code`` query parameter and returns a user ``token``. **Cannot be used outside of the site for security reasons but documented in case we wish to allow its use in the future.**",
		Req:         types.UserLogin{},
		Resp:        types.UserLoginResponse{},
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

	var req types.UserLogin

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

	// Fetch hashed password and token given user id
	var userId string
	var hashedPw string
	var token string

	err = state.Pool.QueryRow(d.Context, "SELECT id, password, token FROM users WHERE username = $1", req.Username).Scan(&userId, &hashedPw, &token)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Json: types.ApiError{
				Message: "User ID not found",
			},
			Status: http.StatusNotFound,
		}
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.HttpResponse{
			Json: types.ApiError{
				Message: "An error occurred while fetching the user id",
			},
			Status: http.StatusInternalServerError,
		}
	}

	match, err := argon2id.ComparePasswordAndHash(req.Password, hashedPw)

	if err != nil {
		state.Logger.Error(err)
		return uapi.HttpResponse{
			Json: types.ApiError{
				Message: "An error occurred validating your identity",
			},
			Status: http.StatusInternalServerError,
		}
	}

	if !match {
		return uapi.HttpResponse{
			Json: types.ApiError{
				Message: "Incorrect user id or password",
			},
			Status: http.StatusUnauthorized,
		}
	}

	return uapi.HttpResponse{
		Json: types.UserLoginResponse{
			UserID: userId,
			Token:  token,
		},
		Status: http.StatusOK,
	}
}
