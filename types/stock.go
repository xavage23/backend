package types

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type StockRatio struct {
	ID         string         `db:"id" json:"id" description:"The ID of the stock ratio"`
	Name       string         `db:"name" json:"name" description:"The name of the stock ratio"`
	ValueText  pgtype.Text    `db:"value_text" json:"value_text" description:"The value of the stock ratio as text"`
	Value      pgtype.Numeric `db:"value" json:"value" description:"The value of the stock ratio"`
	PriceIndex int            `db:"price_index" json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
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
	KnownRatios  []KnownRatios     `db:"-" json:"known_ratios" description:"The known ratios of the stock, may not always be present"`
	PriorRatios  []PriorRatios     `db:"-" json:"prior_ratios" description:"The prior ratios of the stock, may not always be present"`
	Includes     []string          `db:"-" json:"includes,omitempty" description:"The included data present in this stock"`
}

type KnownRatios struct {
	Ratios     []StockRatio `json:"ratios" description:"The ratios of the stock"`
	PriceIndex int          `json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
}

type PriorRatios struct {
	Ratios     []StockRatio `json:"ratios" description:"The ratios of the stock"`
	PriceIndex int          `json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
	Game       Game         `json:"game" description:"The game object"`
}

type StockList struct {
	Stocks     []*Stock `json:"stocks" description:"The list of stocks"`
	PriceIndex int      `json:"price_index" description:"The price index/snapshot of the stock at the time of the transaction"`
}

type News struct {
	ID              string          `db:"id" json:"id" description:"The ID of the news"`
	Title           string          `db:"title" json:"title" description:"The title of the news"`
	Description     string          `db:"description" json:"description" description:"The description of the news"`
	Published       bool            `db:"published" json:"published" description:"Whether the news has been published"`
	AffectedStockID pgtype.UUID     `db:"affected_stock_id" json:"affected_stock_id" description:"The ID of the stock affected by the news"`
	AffectedStock   *Stock          `db:"-" json:"affected_stock" description:"The stock affected by the news, may not always be present"`
	GameID          string          `db:"game_id" json:"game_id" description:"The ID of the game"`
	ShowAt          pgtype.Interval `db:"show_at" json:"show_at" description:"The time at which the news should be shown"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at" description:"The time the news was created"`
}

type Portfolio struct {
	Stock   *Stock                    `json:"stock" description:"The stock"`
	Amounts map[int64]PortfolioAmount `json:"amount" description:"A map of the sale price to the amount of the stock"`
}

type PortfolioAmount struct {
	Amount int64 `json:"amount" description:"The amount of the stock sold at sale_price"`
}
