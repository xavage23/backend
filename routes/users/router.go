package users

import (
	"xavagebb/api"
	"xavagebb/routes/users/endpoints/check_auth"
	"xavagebb/routes/users/endpoints/get_user"
	"xavagebb/routes/users/endpoints/login_user"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Users"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to users"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/users",
		OpId:    "login_user",
		Method:  uapi.PUT,
		Docs:    login_user.Docs,
		Handler: login_user.Route,
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}",
		OpId:    "get_user",
		Method:  uapi.GET,
		Docs:    get_user.Docs,
		Handler: get_user.Route,
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}/check_auth",
		OpId:    "check_auth",
		Method:  uapi.POST,
		Docs:    check_auth.Docs,
		Handler: check_auth.Route,
		Auth: []uapi.AuthType{
			{
				URLVar:       "userId",
				Type:         api.TargetTypeUser,
				AllowedScope: "notingame",
			},
		},
	}.Route(r)
}
