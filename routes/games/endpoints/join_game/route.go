package join_game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"xavagebb/state"
	"xavagebb/transact"

	"xavagebb/types"

	"github.com/go-playground/validator/v10"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/infinitybotlist/eureka/uapi/ratelimit"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slices"
)

type GameMigrationMethod string

const (
	GameMigrationMethodMoveEntireTransactionHistory GameMigrationMethod = "move_entire_transaction_history"
	GameMigrationMethodCondensedMigration           GameMigrationMethod = "condensed_migration"
	GameMigrationMethodNoMigration                  GameMigrationMethod = "no_migration"
)

var (
	compiledMessages = uapi.CompileValidationErrors(types.GameJoinRequest{})
)

func Docs() *docs.Doc {
	return &docs.Doc{
		Summary:     "Login User",
		Description: "Takes in a ``code`` query parameter and returns a user ``token``. **Cannot be used outside of the site for security reasons but documented in case we wish to allow its use in the future.**",
		Params: []docs.Parameter{
			{
				Name:        "userId",
				In:          "path",
				Description: "The ID of the user",
				Required:    true,
				Schema:      docs.IdSchema,
			},
		},
		Req:  types.GameJoinRequest{},
		Resp: types.GameJoinResponse{},
	}
}

func Route(d uapi.RouteData, r *http.Request) uapi.HttpResponse {
	limit, err := ratelimit.Ratelimit{
		Expiry:      1 * time.Second,
		MaxRequests: 10,
		Bucket:      "join",
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

	var req types.GameJoinRequest

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

	// Check that the game exists with the current code and passphrase
	var gameId string
	var initialBalance int
	var enabled bool
	var oldStocksCarryOver bool
	var gameMigrationMethod GameMigrationMethod

	err = state.Pool.QueryRow(d.Context, "SELECT id, initial_balance, enabled, old_stocks_carry_over, game_migration_method FROM games WHERE code = $1", req.GameCode).Scan(&gameId, &initialBalance, &enabled, &oldStocksCarryOver, &gameMigrationMethod)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusNotFound,
			Json:   types.ApiError{Message: "Game not found!"},
		}
	}

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	if !enabled {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "This game is not enabled and thus cannot be joined!"},
		}
	}

	// Check that the user is allowed to join the game
	//
	// 1. Check game_allowed_users to ensure user is still allowed to use the API
	var gacCount int64

	err = state.Pool.QueryRow(state.Context, "SELECT COUNT(*) FROM game_allowed_users WHERE user_id = $1 AND game_id = $2", d.Auth.ID, gameId).Scan(&gacCount)

	if errors.Is(err, pgx.ErrNoRows) {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "This user does not have permission to play this game [count]!"},
		}
	}

	if err != nil {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "Failed to fetch selected game: " + err.Error()},
		}
	}

	if gacCount == 0 {
		return uapi.HttpResponse{
			Status: http.StatusForbidden,
			Json:   types.ApiError{Message: "This user does not have permission to play this game!"},
		}
	}

	// Check that the user is not already in the game
	var gameUserId string

	err = state.Pool.QueryRow(d.Context, "SELECT id FROM game_users WHERE user_id = $1 AND game_id = $2", d.Auth.ID, gameId).Scan(&gameUserId)

	if err == nil {
		return uapi.HttpResponse{
			Json: types.GameJoinResponse{
				ID:  gameUserId,
				New: false,
			},
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	tx, err := state.Pool.Begin(d.Context)

	if err != nil {
		state.Logger.Error("tx create error:", err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	defer tx.Rollback(d.Context)

	// Migrate user transactions
	err = migrateUserTransactions(d.Context, tx, d.Auth.ID, gameId, oldStocksCarryOver, gameMigrationMethod)

	if err != nil {
		state.Logger.Error("migrate error:", err)
		return uapi.HttpResponse{
			Status: http.StatusInternalServerError,
			Json:   types.ApiError{Message: "Failed to migrate user transactions: " + err.Error()},
		}
	}

	// Create the game join
	err = tx.QueryRow(d.Context, "INSERT INTO game_users (user_id, game_id, initial_balance) VALUES ($1, $2, $3) RETURNING id", d.Auth.ID, gameId, initialBalance).Scan(&gameUserId)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	err = tx.Commit(d.Context)

	if err != nil {
		state.Logger.Error(err)
		return uapi.DefaultResponse(http.StatusInternalServerError)
	}

	return uapi.HttpResponse{
		Json: types.GameJoinResponse{
			ID:  gameUserId,
			New: true,
		},
	}
}

func migrateUserTransactions(ctx context.Context, tx pgx.Tx, userId string, gameId string, oldStocksCarryOver bool, gameMigrationMethod GameMigrationMethod) error {
	// Early return
	if gameMigrationMethod == GameMigrationMethodNoMigration {
		return nil
	}

	getOldGameIds := func() ([]string, error) {
		var gameNumber int
		err := tx.QueryRow(ctx, "SELECT game_number FROM games WHERE id = $1", gameId).Scan(&gameNumber)

		if err != nil {
			return []string{}, fmt.Errorf("could not fetch game_number %s", err)
		}

		oldGames, err := tx.Query(ctx, "SELECT id FROM games WHERE game_number < $1", gameNumber)

		if err != nil {
			return []string{}, fmt.Errorf("could not fetch old game rows %s", err)
		}

		var gameIds []string

		for oldGames.Next() {
			// Fetch all transactions from the old game for the user ID and copy them
			var oldGameId string
			err = oldGames.Scan(&oldGameId)

			if err != nil {
				return []string{}, fmt.Errorf("couldnt scan game id %s", err)
			}

			gameIds = append(gameIds, oldGameId)
		}

		oldGames.Close()

		return gameIds, nil
	}

	// Handle old stock IDs
	// This returns a list of stock IDs that were handled and an error (if any)
	handleOldStockIds := func(stockIds []string) ([]string, error) {
		var handledStocks []string
		for _, oldStockId := range stockIds {
			// Get the corresponding stock ID with the same ticker as the stock ID in the old game
			if !slices.Contains(handledStocks, oldStockId) {
				var stockId string
				err := tx.QueryRow(ctx, "SELECT id FROM stocks WHERE game_id = $1 AND ticker = (SELECT ticker FROM stocks WHERE id = $2)", gameId, oldStockId).Scan(&stockId)

				if errors.Is(err, pgx.ErrNoRows) {
					if oldStocksCarryOver {
						return handledStocks, fmt.Errorf("this game has old stocks carry over enabled but the transaction on stock with id %s does not have an equivalent ticker in the new game", oldStockId)
					}

					continue
				}

				if err != nil {
					return handledStocks, fmt.Errorf("utid id select %s %s", err, oldStockId)
				}

				// Update the stock ID in the new game
				_, err = tx.Exec(ctx, "UPDATE user_transactions SET stock_id = $1 WHERE stock_id = $2 AND game_id = $3", stockId, oldStockId, gameId)

				if err != nil {
					return handledStocks, fmt.Errorf("utid transact update error %s", err)
				}

				handledStocks = append(handledStocks, oldStockId)
			}
		}

		return handledStocks, nil
	}

	switch gameMigrationMethod {
	case GameMigrationMethodCondensedMigration:
		gameIds, err := getOldGameIds()

		if err != nil {
			return fmt.Errorf("couldnt get old game ids %s", err)
		}

		type cgmNormalPortfolio struct {
			Amount int64
		}

		type cgmData struct {
			GameID           string
			UserTransactions []types.UserTransaction
			// Map of sale price to a map of stock IDs to a map of price indexes to a cgmNormalPortfolios.
			//
			// If number is negative, then the user is net selling the stock
			//
			// Note: if shorting is added, a new ShortPortfolio may need to be added
			NormalPortfolio map[int64]map[string]map[int]cgmNormalPortfolio
		}

		var stockIds []string
		for _, oldGameId := range gameIds {
			uts, err := transact.GetUserTransactionsUnparsed(ctx, userId, oldGameId)

			if err != nil {
				return fmt.Errorf("couldnt get uts %s", err)
			}

			data := cgmData{
				GameID:           oldGameId,
				UserTransactions: uts,
				NormalPortfolio:  map[int64]map[string]map[int]cgmNormalPortfolio{},
			}

			for _, ut := range uts {
				if _, ok := data.NormalPortfolio[ut.SalePrice]; !ok {
					data.NormalPortfolio[ut.SalePrice] = map[string]map[int]cgmNormalPortfolio{}
				}

				if _, ok := data.NormalPortfolio[ut.SalePrice][ut.StockID]; !ok {
					data.NormalPortfolio[ut.SalePrice][ut.StockID] = map[int]cgmNormalPortfolio{}
				}

				var portfolio cgmNormalPortfolio
				if _, ok := data.NormalPortfolio[ut.SalePrice][ut.StockID][ut.PriceIndex]; !ok {
					portfolio = cgmNormalPortfolio{}
				} else {
					portfolio = data.NormalPortfolio[ut.SalePrice][ut.StockID][ut.PriceIndex]
				}

				switch ut.Action {
				case "buy":
					portfolio.Amount += ut.Amount
				case "sell":
					portfolio.Amount -= ut.Amount
				}

				data.NormalPortfolio[ut.SalePrice][ut.StockID][ut.PriceIndex] = portfolio
			}

			for salePrice, stockMap := range data.NormalPortfolio {
				for stockId, priceIndexMap := range stockMap {
					for priceIndex, portfolio := range priceIndexMap {
						// Create condensed user transactions
						state.Logger.Infof("Creating condensed transaction for game %s, stock %s, price index %d, sale price %d, amount %d", oldGameId, stockId, priceIndex, salePrice, portfolio.Amount)

						if portfolio.Amount > 0 {
							_, err = tx.Exec(ctx, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", userId, gameId, oldGameId, stockId, priceIndex, portfolio.Amount, "buy", salePrice)
						} else {
							_, err = tx.Exec(ctx, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", userId, gameId, oldGameId, stockId, priceIndex, -portfolio.Amount, "sell", salePrice)
						}

						if err != nil {
							return fmt.Errorf("couldnt add new user transaction %s", err)
						}
					}

					stockIds = append(stockIds, stockId)
				}
			}
		}

		handledStocks, err := handleOldStockIds(stockIds)

		if err != nil {
			return fmt.Errorf("couldnt handle old stock ids %s (%d/%d)", err, len(handledStocks), len(stockIds))
		}
	case GameMigrationMethodMoveEntireTransactionHistory:
		var stockIds []string

		gameIds, err := getOldGameIds()

		if err != nil {
			return fmt.Errorf("couldnt get old game ids %s", err)
		}

		// Copy transactions to new game
		for _, oldGameId := range gameIds {
			rows, err := tx.Query(ctx, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) SELECT $1, $2, origin_game_id, stock_id, price_index, amount, action, sale_price FROM user_transactions WHERE game_id = $3 AND user_id = $4 RETURNING stock_id", userId, gameId, oldGameId, userId)

			if err != nil {
				return fmt.Errorf("couldnt add new uts %s", err)
			}

			for rows.Next() {
				var stockId string
				err = rows.Scan(&stockId)

				if err != nil {
					return fmt.Errorf("utid update error %s", err)
				}

				stockIds = append(stockIds, stockId)
			}

			rows.Close()
		}

		// Update stock IDs
		handledStocks, err := handleOldStockIds(stockIds)

		if err != nil {
			return fmt.Errorf("couldnt handle old stock ids %s (%d/%d)", err, len(handledStocks), len(stockIds))
		}
	}

	return nil
}
