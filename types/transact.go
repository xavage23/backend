package types

import "time"

type CreateTransaction struct {
	StockID string `json:"stock_id" description:"The ID of the stock"`
	Amount  int64  `json:"amount" description:"The amount of the stock to trade"`
	Action  string `json:"action" description:"The type of the transaction (buy or sell)"`
}

type UserTransaction struct {
	ID           string    `db:"id" json:"id" description:"The ID of the transaction"`
	UserID       string    `db:"user_id" json:"user_id" description:"The ID of the user"`
	GameID       string    `db:"game_id" json:"game_id" description:"The ID of the game"`
	OriginGameID string    `db:"origin_game_id" json:"origin_game_id" description:"The ID of the game where the transaction originated from"`
	StockID      string    `db:"stock_id" json:"stock_id" description:"The ID of the stock"`
	PriceIndex   int       `db:"price_index" json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
	SalePrice    int64     `db:"sale_price" json:"sale_price" description:"The price of the stock at the time of the transaction"`
	Amount       int64     `db:"amount" json:"amount" description:"The amount of the stock traded"`
	Action       string    `db:"action" json:"action" description:"The type of the transaction (buy or sell)"`
	CreatedAt    time.Time `db:"created_at" json:"created_at" description:"The time the transaction was created"`
}

type TransactionList struct {
	Transactions []UserTransaction `json:"transactions" description:"The list of transactions"`
	Users        map[string]*User  `json:"users" description:"The list of users"`
	Games        map[string]*Game  `json:"games" description:"The list of games"`
	Stocks       map[string]*Stock `json:"stocks" description:"The list of stocks"`
}
