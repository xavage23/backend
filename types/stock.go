package types

import "time"

type StockRatio struct {
	Name  string  `json:"name" description:"The name of the ratio"`
	Ratio float64 `json:"ratio" description:"The ratio value"`
}

type Stock struct {
	ID           string        `db:"id" json:"id" description:"The ID of the stock"`
	GameID       string        `db:"game_id" json:"game_id" description:"The ID of the game"`
	Ticker       string        `db:"ticker" json:"ticker" description:"The ticker of the stock"`
	CompanyName  string        `db:"company_name" json:"company_name" description:"The name of the company"`
	StartPrice   int64         `db:"start_price" json:"-" description:"The price of the stock at the start of the game in cents"`
	EndPrice     int64         `db:"end_price" json:"-" description:"The price of the stock at the end of the game in cents"`
	CurrentPrice int64         `db:"-" json:"current_price" description:"The current price of the stock in cents"`
	KnownPrices  []int64       `db:"-" json:"known_prices" description:"The known prices of the stock in cents"`
	CreatedAt    time.Time     `db:"created_at" json:"created_at" description:"The time the stock was created"`
	Ratios       []*StockRatio `db:"-" json:"ratios,omitempty" description:"The ratios of the stock, may not always be present"`
}
