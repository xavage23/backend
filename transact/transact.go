package transact

import (
	"context"
	"strings"
	"xavagebb/db"
	"xavagebb/state"
	"xavagebb/types"

	"github.com/jackc/pgx/v5"
)

var (
	userTransactionColsArr = db.GetCols(types.UserTransaction{})
	userTransactionCols    = strings.Join(userTransactionColsArr, ", ")

	stockColsArr = db.GetCols(types.Stock{})
	stockCols    = strings.Join(stockColsArr, ", ")
)

func GetAllTransactions(ctx context.Context, gameId string) ([]types.UserTransaction, error) {
	rows, err := state.Pool.Query(ctx, "SELECT "+userTransactionCols+" FROM user_transactions WHERE game_id = $1 ORDER BY created_at DESC", gameId)

	if err != nil {
		return nil, err
	}

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.UserTransaction])

	if err != nil {
		return nil, err
	}

	return parseTrList(ctx, gameId, transactions)
}

func GetUserTransactions(ctx context.Context, userId string, gameId string) ([]types.UserTransaction, error) {
	rows, err := state.Pool.Query(ctx, "SELECT "+userTransactionCols+" FROM user_transactions WHERE user_id = $1 AND game_id = $2 ORDER BY created_at DESC", userId, gameId)

	if err != nil {
		return nil, err
	}

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.UserTransaction])

	if err != nil {
		return nil, err
	}

	return parseTrList(ctx, gameId, transactions)
}

func parseTrList(ctx context.Context, gameId string, transactions []types.UserTransaction) ([]types.UserTransaction, error) {
	var gameCurrentPriceIndex int

	err := state.Pool.QueryRow(ctx, "SELECT current_price_index FROM games WHERE id = $1", gameId).Scan(&gameCurrentPriceIndex)

	if err != nil {
		return nil, err
	}

	var cachedStocks = make(map[string]*types.Stock)
	for i := range transactions {
		cachedStock, ok := cachedStocks[transactions[i].StockID]

		if ok {
			transactions[i].Stock = cachedStock
			continue
		}

		stock, err := GetStock(ctx, transactions[i].StockID, gameCurrentPriceIndex)

		if err != nil {
			return nil, err
		}

		cachedStocks[transactions[i].StockID] = stock
		transactions[i].Stock = stock
	}

	return transactions, nil
}

// Parses a list of user transactions to find the current balance of the user
func GetUserCurrentBalance(initialBalance int64, uts []types.UserTransaction) int64 {
	var currentBalance = initialBalance
	for _, ut := range uts {
		var stockPrice int64

		if ut.Stock == nil {
			panic("Stock is nil")
		}

		if ut.PriceIndex > len(ut.Stock.Prices)-1 {
			stockPrice = ut.Stock.Prices[len(ut.Stock.Prices)-1]
		} else {
			stockPrice = ut.Stock.Prices[ut.PriceIndex]
		}

		switch ut.Action {
		case "buy":
			currentBalance -= stockPrice * ut.Amount
		case "sell":
			currentBalance += stockPrice * ut.Amount
		}
	}

	return currentBalance
}

// Given a stock ID and what price to use, return a Stock object
//
// Current price can be found by fetching the current_price_index field in DB for the `games` table
//
// This function does not handle ratios, use GetStockRatios (not implemented) for that
func GetStock(ctx context.Context, stockId string, currentPriceIndex int) (*types.Stock, error) {
	row, err := state.Pool.Query(ctx, "SELECT "+stockCols+" FROM stocks WHERE id = $1", stockId)

	if err != nil {
		return nil, err
	}

	stock, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.Stock])

	if err != nil {
		return nil, err
	}

	if currentPriceIndex > len(stock.Prices)-1 {
		currentPriceIndex = len(stock.Prices) - 1
	}

	stock.CurrentPrice = stock.Prices[currentPriceIndex]
	// Known prices is all prices until the current price
	for i := range stock.Prices {
		stock.KnownPrices = append(stock.KnownPrices, stock.Prices[i])
		if i >= currentPriceIndex {
			break
		}
	}

	return &stock, nil
}
