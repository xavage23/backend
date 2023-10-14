package transactions

import (
	"xavagebb/api"
	"xavagebb/routes/transactions/endpoints/get_transaction_history"

	"github.com/go-chi/chi/v5"
	"github.com/infinitybotlist/eureka/uapi"
)

const tagName = "Transactions"

type Router struct{}

func (b Router) Tag() (string, string) {
	return tagName, "These API endpoints are related to transactions"
}

func (b Router) Routes(r *chi.Mux) {
	uapi.Route{
		Pattern: "/users/{userId}/transactions",
		OpId:    "get_transaction_history",
		Method:  uapi.GET,
		Docs:    get_transaction_history.Docs,
		Handler: get_transaction_history.Route,
		Auth: []uapi.AuthType{
			{
				URLVar: "userId",
				Type:   api.TargetTypeUser,
			},
		},
	}.Route(r)
}
