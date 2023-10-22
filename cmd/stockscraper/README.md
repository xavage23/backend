# Stock Scraper

Given a list of company names, this program will scrape the ticker and stock price of each company and output the results to a CSV file.

## Data

Note that data entries support comments starting with ``#`` either at the start of a new line or after a parameter. The following files are needed:

- ``accepted_exchanges.txt``: A list of accepted exchanges for finding a stock ticker.
- ``companies.txt``: A list of company names to scrape, newline separated. Parentheses are not supported.