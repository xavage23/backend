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
