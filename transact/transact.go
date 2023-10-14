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

	return transactions, nil
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

	return transactions, nil
}

// Parses a list of user transactions to find the current balance of the user
func GetUserCurrentBalance(initialBalance int64, uts []types.UserTransaction) int64 {
	var currentBalance = initialBalance
	for _, ut := range uts {
		switch ut.Action {
		case "buy":
			currentBalance -= ut.StockPrice * ut.Amount
		case "sell":
			currentBalance += ut.StockPrice * ut.Amount
		}
	}

	return currentBalance
}

// Given a stock ID and what price to use, return a Stock object
//
// Current price can be found by fetching the current_price field in DB for the `games` table
//
// This function does not handle ratios, use GetStockRatios (not implemented) for that
func GetStock(ctx context.Context, stockId string, currentPrice string) (*types.Stock, error) {
	row, err := state.Pool.Query(ctx, "SELECT "+stockCols+" FROM stocks WHERE id = $1", stockId)

	if err != nil {
		return nil, err
	}

	stock, err := pgx.CollectOneRow(row, pgx.RowToStructByName[types.Stock])

	if err != nil {
		return nil, err
	}

	if currentPrice == "start" {
		stock.CurrentPrice = stock.StartPrice
		stock.KnownPrices = []int64{stock.StartPrice}
	} else {
		stock.CurrentPrice = stock.EndPrice
		stock.KnownPrices = []int64{stock.StartPrice, stock.EndPrice}
	}

	return &stock, nil
}
