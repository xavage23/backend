package get_user_portfolio

import (
	"net/http"
	"xavagebb/state"
	"xavagebb/transact"
	"xavagebb/types"

	"github.com/go-chi/chi/v5"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Get User Portfolio",
		Description: "Returns the portfolio of the authenticated user as a map of the stock id to its portfolio in random order.",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
			{
				Name:        "gameId",
				In:          "path",
				Description: "The ID of the game",
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
		Resp: map[string]types.Portfolio{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	userId := chi.URLParam(r, "userId")
	gameId, ok := d.Auth.Data["gameId"].(string)

	if !ok {
		state.Logger.Error("gameId not found in auth data", d.Auth.Data)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if userId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	if gameId == "" {
		return uapi.DefaultResponse(http.StatusBadRequest)
	}

	var uts []types.UserTransaction
	var err error

	currentPriceIndex, err := transact.GetCurrentPriceIndex(d.Context, gameId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	uts, err = transact.GetUserTransactions(d.Context, userId, gameId, currentPriceIndex)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	var portfolio = map[string]*types.Portfolio{}

	for i := range uts {
		_, ok := portfolio[uts[i].StockID]

		if !ok {
			stock, err := transact.GetStock(d.Context, uts[i].StockID, currentPriceIndex)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			pp, err := transact.GetPriorStockPrices(d.Context, gameId, uts[i].Stock.Ticker)

			if err != nil {
				state.Logger.Error(err)
				return uapi.DefaultResponse(http.StatusInternalServerError)
			}

			stock.PriorPrices = pp
			stock.Includes = []string{"prior_prices"}

			portfolio[uts[i].StockID] = &types.Portfolio{
				Stock:   stock,
				Amounts: map[int]types.PortfolioAmount{},
			}
		}

		pa, ok := portfolio[uts[i].StockID].Amounts[uts[i].PriceIndex]

		if !ok {
			pa = types.PortfolioAmount{
				SalePrice: uts[i].SalePrice,
			}
			portfolio[uts[i].StockID].Amounts[uts[i].PriceIndex] = pa
		}

		switch uts[i].Action {
		case "buy":
			pa.Amount += uts[i].Amount
		case "sell":
			pa.Amount -= uts[i].Amount
		}

		portfolio[uts[i].StockID].Amounts[uts[i].PriceIndex] = pa
	}

	return uapi.HttpResponse{
		Json: portfolio,
	}
}
