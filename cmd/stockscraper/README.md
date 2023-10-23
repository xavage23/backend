# Stock Scraper

Given a list of company names, this program will scrape the ticker and stock price of each company and output the results to a JSON file that can then be imported using ``admintool-cli``.

## Data

Note that data entries support comments starting with ``#`` either at the start of a new line or after a parameter. The following files are needed:

- ``accepted_exchanges.txt``: A list of accepted exchanges for finding a stock ticker. The defaults are good enough for most cases including the XAVAGE23 Bulls and Bears event this stock simulator was and is made for.
- ``companies.txt``: A list of company names to scrape, newline separated. Parentheses are not supported and should thus be commented out (e.g. ``Company Name # (Blah)``)
- ``company_ignore.txt``: In case you want to temporarily ignore a company, you can add it to this file. This is useful if you want to temporarily ignore a company that is not working properly / needs further review
- ``times.txt`` the unix epoch (unix timestamp, see example in times.txt to get the stock price for. You can use [Epoch Converter](https://www.epochconverter.com/) to convert dates to unix epoch.
- ``alphavantage_key.txt``: Your Alpha Vantage API key. See [API Keys](#api-keys) for more information.

## Unknown Stocks

StockScraper only handles US stock ratios (prices is supported for all stocks listed on Yahoo Finance). 

If prices are not found, then you may need to comment it out and manually add it to the JSON generated. For ratios, use ``data/supplement.yaml`` to provide ratios for such stocks. For non-US stocks, you may also need to provide ratios in this way.

A good data source for these types of stocks is ``https://companiesmarketcap.com`` and ``https://www.wsj.com/market-data/quotes/<ticker>/financials/annual/balance-sheet`` and ``https://ycharts.com/``

## API Keys

We make use of Yahoo Finance for most data including resolving tickers and getting stock prices on the specific times

Some data is not available on Yahoo Finances. For this, we make use of the Alpha Vantage API. To use this, you must have an API key. You can get one for free at [Alpha Vantage](https://www.alphavantage.co/). Then save this to ``data/alphavantage_key.txt``.