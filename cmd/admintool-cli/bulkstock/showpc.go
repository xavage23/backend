package bulkstock

import (
	"admintool-cli/common"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func ShowPriceChanges(progname string, args []string) {
	fileName := args[0]

	var onlyShowPriceTimes bool

	for _, arg := range args {
		if arg == "--only-show-price-times" {
			onlyShowPriceTimes = true
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

	for _, round := range importFile.Games {
		bold(round.IdentifyColumnName + ": " + round.IdentifyColumnValue)

		for i, priceTime := range round.PriceTimes {
			common.StatusBoldBlue("Price", i, time.Unix(priceTime, 0).Format("2006-01-02 15:04:05"))
		}

		if onlyShowPriceTimes {
			continue
		}

		for i, stock := range round.Stocks {
			var priceChange = []string{}

			for _, price := range stock.Prices {
				var priceInDollars = float64(price) / 100
				priceChange = append(priceChange, fmt.Sprintf("%.2f", priceInDollars))
			}

			fmt.Printf("%d. %s (%s): %s\n", i, stock.Ticker, stock.CompanyName, strings.Join(priceChange, " -> "))
		}

		fmt.Printf("\n\n")
	}
}
