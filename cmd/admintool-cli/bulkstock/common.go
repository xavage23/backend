package bulkstock

import (
	"github.com/fatih/color"
	"github.com/go-playground/validator/v10"
)

var (
	v    = validator.New()
	bold = color.New(color.Bold).PrintlnFunc()
)

type ImportFile struct {
	Games []ImportGame `yaml:"games" validate:"required"`
}

type ImportGame struct {
	IdentifyColumnName  string        `yaml:"identify_column_name" validate:"required"`
	IdentifyColumnValue string        `yaml:"identify_column_value" validate:"required"`
	PriceTimes          []int64       `yaml:"price_times" validate:"required"`
	Stocks              []ImportStock `yaml:"stocks" validate:"required,dive,required"`
}

type ImportStock struct {
	Ticker      string             `yaml:"ticker" validate:"required"`
	CompanyName string             `yaml:"company_name" validate:"required"`
	Prices      []int64            `yaml:"prices" validate:"required"`
	Description string             `yaml:"description" validate:"required"`
	Ratios      []ImportStockRatio `yaml:"ratios" validate:"required,dive,required"`
}

type ImportStockRatio struct {
	Name       string  `yaml:"name" validate:"required"`
	ValueText  *string `yaml:"value_text"`
	Value      float64 `yaml:"value" validate:"required"`
	PriceIndex int64   `yaml:"price_index" validate:"required"`
}
