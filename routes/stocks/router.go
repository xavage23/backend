package stocks

import (
	"xavagebb/api"
	"xavagebb/routes/stocks/endpoints/get_stock"
	"xavagebb/routes/stocks/endpoints/get_stock_list"
	"xavagebb/routes/stocks/endpoints/get_stock_news"

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

	uapi.Route{
		Pattern: "/users/{userId}/stocks/{stockId}",
		OpId:    "get_stock",
		Method:  uapi.GET,
		Docs:    get_stock.Docs,
		Handler: get_stock.Route,
		Auth: []uapi.AuthType{
			{
				URLVar: "userId",
				Type:   api.TargetTypeUser,
			},
		},
	}.Route(r)

	uapi.Route{
		Pattern: "/users/{userId}/news",
		OpId:    "get_stock_news",
		Method:  uapi.GET,
		Docs:    get_stock_news.Docs,
		Handler: get_stock_news.Route,
		Auth: []uapi.AuthType{
			{
				URLVar: "userId",
				Type:   api.TargetTypeUser,
			},
		},
	}.Route(r)
}
