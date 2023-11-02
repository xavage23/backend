package bulkstock

import (
	"admintool-cli/common"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

func BulkImportStock(progname string, args []string) {
	dbName := args[0]
	fileName := args[1]

	var allowGameWithExistingStocks bool

	for _, arg := range args {
		if arg == "--allow-game-with-existing-stocks" {
			allowGameWithExistingStocks = true
			break
		}
	}

	f, err := os.Open(fileName)

	if err != nil {
		common.Fatal(err)
	}

	// YAML file decode
	var importFile ImportFile

	err = yaml.NewDecoder(f).Decode(&importFile)

	if err != nil {
		common.Fatal(fmt.Errorf("error decoding YAML file: %w", err))
	}

	// Validate the file
	err = v.Struct(importFile)

	if err != nil {
		common.Fatal(fmt.Errorf("error validating YAML file: %w", err))
	}

	bold("Loaded", len(importFile.Games), "games from", fileName)

	_pool, err := pgxpool.New(common.Ctx, "postgres:///"+dbName)

	if err != nil {
		panic(err)
	}

	sp := common.NewSandboxPool(_pool)

	sp.AllowCommit = true // Allow commits

	tx, err := sp.Begin(common.Ctx)

	if err != nil {
		common.Fatal(err)
	}

	defer tx.Rollback(common.Ctx)

	for i := range importFile.Games {
		// Get game id from identify column name and value
		var gameId string

		err = tx.QueryRow(common.Ctx, "SELECT id FROM games WHERE "+importFile.Games[i].IdentifyColumnName+" = $1", importFile.Games[i].IdentifyColumnValue).Scan(&gameId)

		if errors.Is(err, pgx.ErrNoRows) {
			common.Fatal(fmt.Errorf("game with %s = %s not found", importFile.Games[i].IdentifyColumnName, importFile.Games[i].IdentifyColumnValue))
		}

		if err != nil {
			common.Fatal(fmt.Errorf("error getting game id: %w", err))
		}

		bold("Processing game", gameId)

		// First, set price_times column
		var parsedPriceTimes []time.Time

		for j := range importFile.Games[i].PriceTimes {
			parsedPriceTimes = append(parsedPriceTimes, time.Unix(importFile.Games[i].PriceTimes[j], 0))
		}

		err = tx.Exec(common.Ctx, "UPDATE games SET price_times = $1 WHERE id = $2", parsedPriceTimes, gameId)

		if err != nil {
			common.Fatal(fmt.Errorf("error setting price_times: %w", err))
		}

		// Check for existing stocks
		if !allowGameWithExistingStocks {
			var existingStocksCount int64

			err = tx.QueryRow(common.Ctx, "SELECT COUNT(*) FROM stocks WHERE game_id = $1", gameId).Scan(&existingStocksCount)

			if err != nil {
				common.Fatal(fmt.Errorf("error getting existing stocks count: %w", err))
			}

			if existingStocksCount > 0 {
				common.Fatal(fmt.Errorf("game %s already has stocks", gameId))
			}
		}

		for j := range importFile.Games[i].Stocks {
			// Add stock first
			var stockId string

			// Game ID, ticket, company name, description, prices
			err = tx.QueryRow(
				common.Ctx,
				`INSERT INTO stocks (game_id, ticker, company_name, description, prices) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				gameId,
				importFile.Games[i].Stocks[j].Ticker,
				importFile.Games[i].Stocks[j].CompanyName,
				importFile.Games[i].Stocks[j].Description,
				importFile.Games[i].Stocks[j].Prices,
			).Scan(&stockId)

			if err != nil {
				common.Fatal(fmt.Errorf("error adding stock: %w", err))
			}

			// Then add all the ratios
			for k := range importFile.Games[i].Stocks[j].Ratios {
				// Stock ID, name, value text, value, price index
				err = tx.Exec(
					common.Ctx,
					`INSERT INTO stock_ratios (stock_id, name, value_text, value, price_index) VALUES ($1, $2, $3, $4, $5)`,
					stockId,
					importFile.Games[i].Stocks[j].Ratios[k].Name,
					importFile.Games[i].Stocks[j].Ratios[k].ValueText,
					importFile.Games[i].Stocks[j].Ratios[k].Value,
					importFile.Games[i].Stocks[j].Ratios[k].PriceIndex,
				)

				if err != nil {
					common.Fatal(fmt.Errorf("error adding stock ratio: %w", err))
				}
			}
		}
	}

	err = tx.Commit(common.Ctx)

	if err != nil {
		common.Fatal(fmt.Errorf("error committing transaction: %w", err))
	}
}
