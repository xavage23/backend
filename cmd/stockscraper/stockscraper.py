#!/bin/python3
"""
Given a list of company names, find:
    - The stocks ticker symbol
    - The stock price at specific dates
"""
import os
import sys
from time import sleep
from models import BadStockExchangeException, ImportFile, ImportGame, ImportStock, ImportStockRatio, SDData, RoundMap, Stock, StockRatios, APIClient, StockPrice, SupplementData
from utils import red_print, bold_print, debug_print, yellow_print, text_strip
from ruamel.yaml import YAML


if len(sys.argv) != 2:
    red_print(f"Usage: {sys.argv[0]} <output file>")
    exit(1)

companies: list[str] = []
accepted_exchanges: list[str] = []
times: list[int] = []
alphavantage_key: str = ""
supplemental_ratios: dict[str, dict[int, StockRatios]] = {}
yaml = YAML(typ='safe')
company_ignore_list: list[str] = []
supplement_data = SupplementData(root={})
stock_data: dict[str, SDData] = {}
round_maps: list[RoundMap] = []

if os.path.exists("data/company_ignore.txt"):
    with open("data/company_ignore.txt") as f:
        company_ignore_list = [text_strip(entry) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/companies.txt") as f:
    companies = [text_strip(entry) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/accepted_exchanges.txt") as f:
    accepted_exchanges = [text_strip(entry) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/times.txt") as f:
    times = [int(text_strip(entry)) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/alphavantage_key.txt") as f:
    alphavantage_key = f.read().strip()

with open("data/roundmap.txt") as f:
    round_maps = [RoundMap.from_line(text_strip(entry)) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

for company in companies:
    if company in company_ignore_list:
        red_print(f"Company {company} is in the ignore list. Remove if not desired")
        companies.remove(company)
        sleep(1)

if os.path.exists("data/supplement.yaml"):
    with open("data/supplement.yaml") as f:
        supplement_data = SupplementData(root=yaml.load(f))

api_client = APIClient(alphavantage_key)

for ticker, data in supplement_data.root.items():
    if len(data) == 0:
        continue # ignore the ticker
    
    for price, ratios in data.items():
        if ticker not in supplemental_ratios:
            supplemental_ratios[ticker] = {}

        supplemental_ratios[ticker][int(price)] = StockRatios(**ratios)

# End of config section

del company, ticker, data, price, ratios, company_ignore_list, alphavantage_key, supplement_data # Ensure we don't accidentally use these variable later

# Now parse the stocks

ratio_bad_stocks: list[str] = []

for index, company in enumerate(companies):
    bold_print(f"{index + 1}/{len(companies)}:", company)

    stock: Stock = None

    # Find ticker symbol from company name on Yahoo Finance
    try:
        res = Stock.get_from_company_name(api_client, company)
    except Exception as err:
        red_print(f"Failed to fetch ticker symbol for {company}, {err}")
        exit(1)

    if not res:
        debug_print(res)
        red_print(f"ERROR: Failed to fetch ticker symbol for {company}")
        exit(1)

    if len(res) == 1:
        stock = res[0]
    else: 
        for s in res:
            if s.exchDisp in accepted_exchanges or s.exchange in accepted_exchanges:
                stock = s
                break
    
    if not stock:
        for r in res:
            debug_print(r)
        red_print(f"Failed to find ticker symbol for {company}")
        exit(1)
    
    debug_print("Found stock:", stock)

    # Get the value of the stock at the given times
    prices: dict[int, StockRatios] = {}
    for time in times:
        try:
            res = StockPrice.get_stock_price(api_client, stock.symbol, time)
        except Exception as err:
            red_print(f"Failed to fetch stock prices for {company}, {err}")
            exit(1)
        debug_print(res)

        if not res:
            red_print(f"Failed to fetch stock prices for {company}")
            exit(1)
                
        prices[time] = res

    ratios: dict[int, StockRatios] = {}

    if stock.symbol in supplemental_ratios:
        for price in prices:
            if price not in supplemental_ratios[stock.symbol]:
                red_print(f"Failed to find supplemental ratios for {ticker} at {price}")
                exit(1)

        ratios = supplemental_ratios[stock.symbol]
    else:
        try:
            ratios = StockRatios.get_stock_ratios_for_time(api_client, stock, prices)
            #debug_print(ratios)
        except BadStockExchangeException:
            ratio_bad_stocks.append(stock.symbol)
        except Exception as err:
            red_print(f"Failed to fetch stock ratios for {company}, {err}")
            exit(1)

    stock_data[stock.symbol] = SDData(stock=stock, prices=prices, ratios=ratios)

if ratio_bad_stocks:
    yellow_print("Stocks with unknown ratios:", ratio_bad_stocks, "count =", len(ratio_bad_stocks))
    sleep(3)

bold_print("=> Processing collected data to importable format")

# Now create the ImportFile
import_games: list[ImportGame] = []

for round_map in round_maps:
    pt: list[int] = []
    time_index_map: dict[int, int] = {}
    stocks: list[ImportStock] = []

    for i in round_map.time_indexes:
        pt.append(times[i])
        time_index_map[times[i]] = i
    
    for ticker, sd in stock_data.items():
        debug_print("Processing", ticker, "for round", round_map.display())

        # Get corresponding prices for the stock based on pt
        prices: list[int] = []

        for pp in pt:
            prices.append(int(sd.prices[pp].close * 100))

        # Get corresponding ratios for the stock
        ratios: list[ImportStockRatio] = []

        for ts, ratio in sd.ratios.items():
            if ts not in time_index_map:
                debug_print("Skipping", ts, "for", ticker, "because it is not in the time index map [likely for next round]")
                continue

            debug_print(ts, ratio)
            price_index = time_index_map[ts]

            ratios.append(
                ImportStockRatio(
                    name="P/E Ratio",
                    value=ratio.pe_ratio,
                    price_index=price_index
                )
            )

            ratios.append(
                ImportStockRatio(
                    name="Earnings Per Share",
                    value=ratio.earnings_per_share,
                    price_index=price_index
                )
            )

            if ratio.debt_to_equity_ratio:
                ratios.append(
                    ImportStockRatio(
                        name="Debt to Equity Ratio",
                        value=ratio.debt_to_equity_ratio,
                        price_index=price_index
                    )
                )

            if ratio.profit_margin:
                ratios.append(
                    ImportStockRatio(
                        name="Profit Margin",
                        value=ratio.profit_margin,
                        price_index=price_index
                    )
                )
        
        company_name = sd.stock.longname if sd.stock.longname else sd.stock.shortname

        description = f"""
{company_name} is a {sd.stock.sector if sd.stock.sector else ""} company on the {sd.stock.exchange} ({sd.stock.exchDisp}) exchange.

The company sells '{sd.stock.typeDisp}' type of stock.
"""

        if sd.stock.score:
            description += f"\n\nYahoo Finance Score: {sd.stock.score}."

        stocks.append(
            ImportStock(
                ticker=ticker,
                company_name=company_name,
                prices=prices,
                description=description,
                ratios=ratios
            )
        )

    import_games.append(
        ImportGame(
            identify_column_name=round_map.identify_column_name,
            identify_column_value=round_map.identify_column_value,
            price_times=pt,
            stocks=stocks
        )
    )

import_file = ImportFile(games=import_games)

bold_print("=> Writing import file")

yaml = YAML()

with open(sys.argv[1], "w") as f:
    yaml.dump(import_file.model_dump(), f)