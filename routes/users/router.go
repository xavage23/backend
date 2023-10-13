package users

import (
	"xavagebb/routes/users/endpoints/login_user"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Users"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to users on IBL"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/users",
		OpId:    "login_user",
		Method:  uapi.PUT,
		Docs:    login_user.Docs,
		Handler: login_user.Route,
	}.Route(r)
}
