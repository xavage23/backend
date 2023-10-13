package types

import "time"

type User struct {
	ID        string    `db:"id" json:"id" description:"The ID of the user"`
	Username  string    `db:"username" json:"username" description:"The username of the user"`
	Enabled   bool      `db:"enabled" json:"enabled" description:"Whether the user is enabled"`
	Root      bool      `db:"root" json:"root" description:"Whether the user is a root user"`
	CreatedAt time.Time `db:"created_at" json:"created_at" description:"The time the user was created"`
}

type GameJoinRequest struct {
	GameCode   string `json:"game_code" description:"The code of the game to join"`
	Passphrase string `json:"passphrase" description:"The passphrase of the game to join"`
}

type GameJoinResponse struct {
	ID  string `json:"id" description:"The ID of the game join"`
	New bool   `json:"new" description:"Whether the game join is new"`
}

type Game struct {
	ID             string    `db:"id" json:"id" description:"The ID of the game"`
	Code           string    `db:"code" json:"code" description:"The code of the game"`
	Enabled        bool      `db:"enabled" json:"enabled" description:"Whether the game is enabled"`
	Description    string    `db:"description" json:"description" description:"The description of the game"`
	CreatedAt      time.Time `db:"created_at" json:"created_at" description:"The time the game was created"`
	CurrentPrice   string    `db:"current_price" json:"current_price" description:"The current price of the game"`
	InitialBalance int64     `db:"initial_balance" json:"initial_balance" description:"The initial balance of the game"`
}

type GameUser struct {
	ID        string    `db:"id" json:"id" description:"The ID of the game join"`
	UserID    string    `db:"user_id" json:"user_id" description:"The ID of the user"`
	GameID    string    `db:"game_id" json:"game_id" description:"The ID of the game"`
	Game      Game      `db:"-" json:"game" description:"The game object"`
	Balance   int64     `db:"balance" json:"balance" description:"The current balance of the user in the game"`
	CreatedAt time.Time `db:"created_at" json:"created_at" description:"The time the game join was created"`
}
