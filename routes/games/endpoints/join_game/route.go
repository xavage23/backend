package join_game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"xavagebb/state"

	"xavagebb/types"

	"github.com/go-playground/validator/v10"
	docs "github.com/infinitybotlist/eureka/doclib"
	"github.com/infinitybotlist/eureka/uapi"
	"github.com/infinitybotlist/eureka/uapi/ratelimit"
	"github.com/jackc/pgx/v5"
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

	err = state.Pool.QueryRow(d.Context, "SELECT id, initial_balance, enabled, old_stocks_carry_over FROM games WHERE code = $1", req.GameCode).Scan(&gameId, &initialBalance, &enabled, &oldStocksCarryOver)

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
	err = migrateUserTransactions(d.Context, tx, d.Auth.ID, gameId, oldStocksCarryOver)

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

func migrateUserTransactions(ctx context.Context, tx pgx.Tx, userId string, gameId string, oldStocksCarryOver bool) error {
	// Copy transactions to new game
	var gameNumber int
	err := tx.QueryRow(ctx, "SELECT game_number FROM games WHERE id = $1", gameId).Scan(&gameNumber)

	if err != nil {
		return fmt.Errorf("could not fetch game_number %s", err)
	}

	oldGames, err := tx.Query(ctx, "SELECT id FROM games WHERE game_number < $1", gameNumber)

	if err != nil {
		return fmt.Errorf("could not fetch old games %s", err)
	}

	var gameIds []string

	for oldGames.Next() {
		// Fetch all transactions from the old game FOR THE user ID and copy them
		var oldGameId string
		err = oldGames.Scan(&oldGameId)

		if err != nil {
			return fmt.Errorf("couldnt scan game id %s", err)
		}

		gameIds = append(gameIds, oldGameId)
	}

	oldGames.Close()

	type utid struct {
		ID      string
		StockId string
	}

	for _, oldGameId := range gameIds {
		rows, err := tx.Query(ctx, "INSERT INTO user_transactions (user_id, game_id, origin_game_id, stock_id, price_index, amount, action, sale_price) SELECT $1, $2, origin_game_id, stock_id, price_index, amount, action, sale_price FROM user_transactions WHERE game_id = $3 AND user_id = $4 RETURNING id, stock_id", userId, gameId, oldGameId, userId)

		if err != nil {
			return fmt.Errorf("couldnt add new uts %s", err)
		}

		var utids []utid
		for rows.Next() {
			var id string
			var stockId string
			err = rows.Scan(&id, &stockId)

			if err != nil {
				return fmt.Errorf("utid update error %s", err)
			}

			utids = append(utids, utid{
				ID:      id,
				StockId: stockId,
			})
		}

		rows.Close()

		for _, utid := range utids {
			// Get the corresponding stock ID with the same ticker as the stock ID in the old game
			var stockId string

			err = tx.QueryRow(ctx, "SELECT id FROM stocks WHERE game_id = $1 AND ticker = (SELECT ticker FROM stocks WHERE id = $2)", gameId, utid.StockId).Scan(&stockId)

			if errors.Is(err, pgx.ErrNoRows) {
				if oldStocksCarryOver {
					return fmt.Errorf("this game has old stocks carry over enabled but the transaction on stock with id %s does not have an equivalent ticker in the new game", utid.StockId)
				}

				continue
			}

			if err != nil {
				return fmt.Errorf("utid id select %s %s", err, utid)
			}

			// Update the stock ID in the new game
			_, err = tx.Exec(ctx, "UPDATE user_transactions SET stock_id = $1 WHERE id = $2", stockId, utid.ID)

			if err != nil {
				return fmt.Errorf("utid transact update error %s", err)
			}
		}
	}

	return nil
}
