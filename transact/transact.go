package transact

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"xavagebb/db"
	"xavagebb/state"
	"xavagebb/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

var (
	json                   = jsoniter.ConfigFastest
	userTransactionColsArr = db.GetCols(types.UserTransaction{})
	userTransactionCols    = strings.Join(userTransactionColsArr, ", ")

	stockColsArr = db.GetCols(types.Stock{})
	stockCols    = strings.Join(stockColsArr, ", ")

	stockRatioColsArr = db.GetCols(types.StockRatio{})
	stockRatioCols    = strings.Join(stockRatioColsArr, ", ")

	gameColsArr = db.GetCols(types.Game{})
	gameCols    = strings.Join(gameColsArr, ", ")

	ppCacheTime    = 3 * time.Minute // Prior prices cache time
	ratioCacheTime = 3 * time.Minute // Stock ratios cache time
)

func ConvertUUIDToString(uuid [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

func GetCurrentPriceIndex(ctx context.Context, gameId string) (int, error) {
	var gameCurrentPriceIndex int

	err := state.Pool.QueryRow(ctx, "SELECT current_price_index FROM games WHERE id = $1", gameId).Scan(&gameCurrentPriceIndex)

	if err != nil {
		return 0, err
	}

	return gameCurrentPriceIndex, nil
}

func GetAllTransactionsUnparsed(ctx context.Context, gameId string) ([]types.UserTransaction, error) {
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

func GetUserTransactionsUnparsed(ctx context.Context, userId string, gameId string) ([]types.UserTransaction, error) {
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

func FillStock(ctx context.Context, stock *types.Stock, currentPriceIndex int, includes []string) (*types.Stock, error) {
	for _, include := range includes {
		switch include {
		case "prior_prices":
			pp, err := GetPriorStockPrices(ctx, stock.GameID, stock.Ticker)

			if err != nil {
				return nil, fmt.Errorf("failed to get prior stock prices: %w", err)
			}

			stock.PriorPrices = pp
		case "known_ratios":
			ratios, err := GetKnownStockRatios(ctx, stock.GameID, stock.ID, currentPriceIndex)

			if err != nil {
				return nil, fmt.Errorf("failed to get known stock ratios: %w", err)
			}

			stock.KnownRatios = ratios
		case "prior_ratios":
			ratios, err := GetPriorStockRatios(ctx, stock.GameID, stock.Ticker, currentPriceIndex)

			if err != nil {
				return nil, fmt.Errorf("failed to get prior stock ratios: %w", err)
			}

			stock.PriorRatios = ratios
		default:
			return nil, fmt.Errorf("unknown include: %s", include)
		}
	}

	stock.Includes = includes

	return stock, nil
}

// Parses a list of user transactions to find the current balance of the user
func GetUserCurrentBalance(initialBalance int64, uts []types.UserTransaction) int64 {
	var currentBalance = initialBalance

	for _, ut := range uts {
		switch ut.Action {
		case "buy":
			currentBalance -= ut.SalePrice * ut.Amount
		case "sell":
			currentBalance += ut.SalePrice * ut.Amount
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

	return ParseStock(ctx, &stock, currentPriceIndex), nil
}

func ParseStock(ctx context.Context, stock *types.Stock, currentPriceIndex int) *types.Stock {
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

	return stock
}

func GetPriorStockPrices(ctx context.Context, gameId, ticker string) ([]types.PriorPricePoint, error) {
	// Check cache first for prior stock prices
	cachedData := state.Redis.Get(ctx, "prior_stock_prices:"+gameId+":"+ticker)

	if cachedData != nil {
		val, err := cachedData.Result()

		if err == nil {
			if val != "" {
				var allPrices []types.PriorPricePoint

				err = json.Unmarshal([]byte(val), &allPrices)

				if err != nil {
					return nil, err
				}

				return allPrices, nil
			} else {
				state.Logger.Debug("Failed to get prior stock prices from cache due to empty cache", zap.String("game_id", gameId), zap.String("ticker", ticker))
			}
		} else {
			state.Logger.Debug("Failed to get prior stock prices from cache", zap.Error(err), zap.String("game_id", gameId), zap.String("ticker", ticker))
		}
	}

	// Get game number of current game
	var gameNumber int

	err := state.Pool.QueryRow(ctx, "SELECT game_number FROM games WHERE id = $1", gameId).Scan(&gameNumber)

	if err != nil {
		return nil, err
	}

	gameRows, err := state.Pool.Query(ctx, "SELECT "+gameCols+" FROM games WHERE game_number < $1 ORDER BY game_number ASC", gameNumber)

	if err != nil {
		return nil, err
	}

	games, err := pgx.CollectRows(gameRows, pgx.RowToStructByName[types.Game])

	if errors.Is(err, pgx.ErrNoRows) {
		err = state.Redis.Set(ctx, "prior_stock_prices:"+gameId+":"+ticker, "[]", ppCacheTime).Err()

		if err != nil {
			return nil, err
		}
		return []types.PriorPricePoint{}, nil
	}

	if err != nil {
		return nil, err
	}

	var allPrices = []types.PriorPricePoint{}
	for _, game := range games {
		// Fetch prices within this game ID
		var prices []int64

		err = state.Pool.QueryRow(ctx, "SELECT prices FROM stocks WHERE game_id = $1 AND ticker = $2", game.ID, ticker).Scan(&prices)

		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		if err != nil {
			return nil, err
		}

		allPrices = append(allPrices, types.PriorPricePoint{
			Game:   game,
			Prices: prices,
		})
	}

	// Cache prior stock prices
	cacheStr, err := json.MarshalToString(allPrices)

	if err != nil {
		return nil, err
	}

	err = state.Redis.Set(ctx, "prior_stock_prices:"+gameId+":"+ticker, cacheStr, ppCacheTime).Err()

	if err != nil {
		return nil, err
	}

	return allPrices, nil
}

func GetKnownStockRatios(ctx context.Context, gameId string, stockId string, currentPriceIndex int) ([]types.KnownRatios, error) {
	// Fetch ratios within this game ID for this stock
	rows, err := state.Pool.Query(ctx, "SELECT "+stockRatioCols+" FROM stock_ratios WHERE stock_id = $1 AND price_index <= $2", stockId, currentPriceIndex)

	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KnownRatios{}, nil
	}

	if err != nil {
		return nil, err
	}

	ratios, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.StockRatio])

	if errors.Is(err, pgx.ErrNoRows) {
		return []types.KnownRatios{}, nil
	}

	if err != nil {
		return nil, err
	}

	// Map of price index to stock ratio
	var ratioMap = make(map[int][]types.StockRatio)

	for _, ratio := range ratios {
		if _, ok := ratioMap[ratio.PriceIndex]; !ok {
			ratioMap[ratio.PriceIndex] = []types.StockRatio{}
		}

		ratioMap[ratio.PriceIndex] = append(ratioMap[ratio.PriceIndex], ratio)
	}

	var knownRatios []types.KnownRatios
	for index, ratios := range ratioMap {
		knownRatios = append(knownRatios, types.KnownRatios{
			PriceIndex: index,
			Ratios:     ratios,
		})
	}

	return knownRatios, nil
}

func GetPriorStockRatios(ctx context.Context, gameId string, ticker string, currentPriceIndex int) ([]types.PriorRatios, error) {
	// Check cache first for prior stock prices
	cachedData := state.Redis.Get(ctx, "stock_ratios:"+ticker)

	if cachedData != nil {
		val, err := cachedData.Result()

		if err == nil {
			if val != "" {
				var krs []types.PriorRatios

				err = json.Unmarshal([]byte(val), &krs)

				if err != nil {
					return nil, err
				}

				if len(krs) == 0 {
					return []types.PriorRatios{}, nil
				}

				return krs, nil
			} else {
				state.Logger.Debug("Failed to get all stock ratios from cache due to empty cache", zap.String("ticker", ticker), zap.String("game_id", gameId))
			}
		} else {
			state.Logger.Debug("Failed to get all stock ratios from cache", zap.Error(err), zap.String("game_id", gameId))
		}
	}

	// Get game number of current game
	var gameNumber int

	err := state.Pool.QueryRow(ctx, "SELECT game_number FROM games WHERE id = $1", gameId).Scan(&gameNumber)

	if err != nil {
		return nil, err
	}

	gameRows, err := state.Pool.Query(ctx, "SELECT "+gameCols+" FROM games WHERE game_number < $1 ORDER BY game_number ASC", gameNumber)

	if err != nil {
		return nil, err
	}

	games, err := pgx.CollectRows(gameRows, pgx.RowToStructByName[types.Game])

	if errors.Is(err, pgx.ErrNoRows) {
		state.Logger.Info("No games in prior stock ratios", zap.String("game_id", gameId), zap.String("ticker", ticker))
		err = state.Redis.Set(ctx, "stock_ratios:"+ticker, "[]", ratioCacheTime).Err()

		if err != nil {
			return nil, err
		}
		return []types.PriorRatios{}, nil
	}

	if err != nil {
		return nil, err
	}

	var priorRatios = []types.PriorRatios{}
	for _, game := range games {
		var maxPriceIndex int

		// Get the maximum price index to get ratios up to
		if gameId == game.ID {
			// This line should never be reached, but just in case
			maxPriceIndex = currentPriceIndex
		} else {
			// Fetch the max price index for this game
			maxPriceIndex, err = GetCurrentPriceIndex(ctx, game.ID)

			if err != nil {
				return nil, err
			}
		}

		state.Logger.Info("Max price index", zap.Int("max_price_index", maxPriceIndex), zap.String("game_id", game.ID), zap.Int("game_number", game.GameNumber), zap.String("stock_ticker", ticker))

		var stockId string

		err = state.Pool.QueryRow(ctx, "SELECT id FROM stocks WHERE game_id = $1 AND ticker = $2", game.ID, ticker).Scan(&stockId)

		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		if err != nil {
			return nil, err
		}

		// Fetch ratios within this game ID for this stock
		rows, err := state.Pool.Query(ctx, "SELECT "+stockRatioCols+" FROM stock_ratios WHERE stock_id = $1 AND price_index <= $2", stockId, maxPriceIndex)

		if errors.Is(err, pgx.ErrNoRows) {
			state.Logger.Info("No rows in prior stock ratios", zap.String("game_id", game.ID), zap.String("ticker", ticker), zap.String("stock_id", stockId))
			continue
		}

		if err != nil {
			return nil, err
		}

		ratios, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.StockRatio])

		if errors.Is(err, pgx.ErrNoRows) {
			state.Logger.Info("No rows in prior stock ratios", zap.String("game_id", game.ID), zap.String("ticker", ticker), zap.String("stock_id", stockId))
			continue
		}

		if err != nil {
			return nil, err
		}

		state.Logger.Info("Found prior stock ratios", zap.String("game_id", game.ID), zap.String("ticker", ticker), zap.String("stock_id", stockId), zap.Int("num_ratios", len(ratios)))

		// Map of price index to stock ratio
		var ratioMap = make(map[int][]types.StockRatio)

		for _, ratio := range ratios {
			if _, ok := ratioMap[ratio.PriceIndex]; !ok {
				ratioMap[ratio.PriceIndex] = []types.StockRatio{}
			}

			ratioMap[ratio.PriceIndex] = append(ratioMap[ratio.PriceIndex], ratio)
		}

		for index, ratios := range ratioMap {
			priorRatios = append(priorRatios, types.PriorRatios{
				Game:       game,
				PriceIndex: index,
				Ratios:     ratios,
			})
		}
	}

	// Cache prior stock prices
	cacheStr, err := json.MarshalToString(priorRatios)

	if err != nil {
		return nil, err
	}

	err = state.Redis.Set(ctx, "stock_ratios:"+ticker, cacheStr, ratioCacheTime).Err()

	if err != nil {
		return nil, err
	}

	return priorRatios, nil
}

func GetTotalStockQuantity(uts []types.UserTransaction, stockId string) int64 {
	var totalQuantity int64
	for _, ut := range uts {
		if ut.StockID != stockId {
			continue
		}

		switch ut.Action {
		case "buy":
			totalQuantity += ut.Amount
		case "sell":
			totalQuantity -= ut.Amount
		}
	}

	return totalQuantity
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
