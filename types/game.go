package types

import (
	"time"
)

type Game struct {
	ID                 string    `db:"id" json:"id" description:"The ID of the game"`
	Code               string    `db:"code" json:"code" description:"The code of the game"`
	Enabled            bool      `db:"enabled" json:"enabled" description:"Whether the game is enabled"`
	TradingAllowed     bool      `db:"trading_allowed" json:"trading_enabled" description:"Whether trading is enabled"`
	Description        string    `db:"description" json:"description" description:"The description of the game"`
	CreatedAt          time.Time `db:"created_at" json:"created_at" description:"The time the game was created"`
	CurrentPriceIndex  int       `db:"current_price_index" json:"current_price_index" description:"The current price index of the game"`
	InitialBalance     int64     `db:"initial_balance" json:"initial_balance" description:"The initial balance of the game"`
	GameNumber         int       `db:"game_number" json:"game_number" description:"The number of the game. Higher numbered games will have transactions from lower game numbers migrated to themselves"`
	OldStocksCarryOver bool      `db:"old_stocks_carry_over" json:"old_stocks_carry_over" description:"Whether stocks from previous games carry over to this game"`
}

type GameJoinRequest struct {
	GameCode string `json:"game_code" description:"The code of the game to join"`
}

type GameJoinResponse struct {
	ID  string `json:"id" description:"The ID of the game join"`
	New bool   `json:"new" description:"Whether the game join is new"`
}

type GameUser struct {
	ID             string    `db:"id" json:"id" description:"The ID of the game join"`
	UserID         string    `db:"user_id" json:"user_id" description:"The ID of the user"`
	GameID         string    `db:"game_id" json:"game_id" description:"The ID of the game"`
	Game           Game      `db:"-" json:"game" description:"The game object"`
	InitialBalance int64     `db:"initial_balance" json:"initial_balance" description:"The initial balance of the user in the game. Usually equal to the games initial balance unless the user is sanctioned/penalized"`
	CurrentBalance int64     `db:"-" json:"current_balance" description:"The current balance of the user in the game calculated by processing all trades made by the user"`
	CreatedAt      time.Time `db:"created_at" json:"created_at" description:"The time the game join was created"`
}
