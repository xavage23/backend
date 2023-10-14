package stocks

import (
	"xavagebb/api"
	"xavagebb/routes/stocks/endpoints/get_stock_list"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Stocks"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to stocks"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/users/{userId}/stocks",
		OpId:    "get_stock_list",
		Method:  uapi.GET,
		Docs:    get_stock_list.Docs,
		Handler: get_stock_list.Route,
		Auth: []uapi.AuthType{
			{
				URLVar: "userId",
				Type:   api.TargetTypeUser,
			},
		},
	}.Route(r)
}
