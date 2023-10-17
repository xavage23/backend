package types

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type StockRatio struct {
	Name  string  `json:"name" description:"The name of the ratio"`
	Ratio float64 `json:"ratio" description:"The ratio value"`
}

type PriorPricePoint struct {
	Prices []int64 `json:"prices" description:"The price of the stock in cents"`
	Game   Game    `json:"game" description:"The game object"`
}

type Stock struct {
	ID           string            `db:"id" json:"id" description:"The ID of the stock"`
	GameID       string            `db:"game_id" json:"game_id" description:"The ID of the game"`
	Ticker       string            `db:"ticker" json:"ticker" description:"The ticker of the stock"`
	CompanyName  string            `db:"company_name" json:"company_name" description:"The name of the company"`
	Description  string            `db:"description" json:"description" description:"The description of the stock"`
	Prices       []int64           `db:"prices" json:"-" description:"The prices of the stock in cents"`
	CurrentPrice int64             `db:"-" json:"current_price" description:"The current price of the stock in cents"`
	KnownPrices  []int64           `db:"-" json:"known_prices" description:"The known prices of the stock in cents"`
	PriorPrices  []PriorPricePoint `db:"-" json:"prior_prices" description:"The prices of the stock in cents"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at" description:"The time the stock was created"`
	Ratios       []*StockRatio     `db:"-" json:"ratios,omitempty" description:"The ratios of the stock, may not always be present"`
	Includes     []string          `db:"-" json:"includes,omitempty" description:"The included data present in this stock"`
}

type StockList struct {
	Stocks     []*Stock `json:"stocks" description:"The list of stocks"`
	PriceIndex int      `json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
}

type News struct {
	ID              string      `db:"id" json:"id" description:"The ID of the news"`
	Title           string      `db:"title" json:"title" description:"The title of the news"`
	Description     string      `db:"description" json:"description" description:"The description of the news"`
	Published       bool        `db:"published" json:"published" description:"Whether the news has been published"`
	AffectedStockID pgtype.UUID `db:"affected_stock_id" json:"affected_stock_id" description:"The ID of the stock affected by the news"`
	AffectedStock   *Stock      `db:"-" json:"affected_stock" description:"The stock affected by the news, may not always be present"`
	GameID          string      `db:"game_id" json:"game_id" description:"The ID of the game"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at" description:"The time the news was created"`
}

type Portfolio struct {
	Stock   *Stock                  `json:"stock" description:"The stock"`
	Amounts map[int]PortfolioAmount `json:"amount" description:"The amount of the stock"`
}

type PortfolioAmount struct {
	SalePrice int64 `json:"sale_price" description:"The price the stock was sold at"`
	Amount    int64 `json:"amount" description:"The amount of the stock sold at sale_price"`
}
