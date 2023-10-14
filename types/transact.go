package types

import "time"

type UserTransaction struct {
	ID         string    `db:"id" json:"id" description:"The ID of the transaction"`
	UserID     string    `db:"user_id" json:"user_id" description:"The ID of the user"`
	GameID     string    `db:"game_id" json:"game_id" description:"The ID of the game"`
	User       *User     `db:"-" json:"user" description:"The user object of the transaction"`
	StockID    string    `db:"stock_id" json:"stock_id" description:"The ID of the stock"`
	Stock      *Stock    `db:"-" json:"stock" description:"The stock object of the transaction, may not always be present"`
	StockPrice int64     `db:"stock_price" json:"stock_price" description:"The price of the stock at the time of the transaction"`
	Amount     int64     `db:"amount" json:"amount" description:"The amount of the stock traded"`
	Action     string    `db:"action" json:"action" description:"The type of the transaction (buy or sell)"`
	CreatedAt  time.Time `db:"created_at" json:"created_at" description:"The time the transaction was created"`
}