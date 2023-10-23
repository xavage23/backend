#!/bin/python3
"""
Given a list of company names, find:
    - The stocks ticker symbol
    - The stock price at specific dates
"""
import os
from time import sleep
from pydantic import BaseModel
from models import BadStockExchangeException, Stock, StockRatios, APIClient, StockPrice, SupplementData
from utils import red_print, bold_print, debug_print
from ruamel.yaml import YAML

companies: list[str] = []
accepted_exchanges: list[str] = []
times: list[int] = []

company_ignore_list: list[str] = []
if os.path.exists("data/company_ignore.txt"):
    with open("data/company_ignore.txt") as f:
        company_ignore_list = [entry.strip().split("#")[0].strip() for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/companies.txt") as f:
    companies = [entry.strip().split("#")[0].strip() for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/accepted_exchanges.txt") as f:
    accepted_exchanges = [entry.strip().split("#")[0].strip() for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/times.txt") as f:
    times = [int(entry.strip().split("#")[0].strip()) for entry in f.read().splitlines() if entry.strip() and not entry.startswith("#")]

with open("data/alphavantage_key.txt") as f:
    ak = f.read().strip()

for company in companies:
    if company in company_ignore_list:
        red_print(f"Company {company} is in the ignore list. Remove if not desired")
        companies.remove(company)
        sleep(1)

supplement_data = SupplementData(root={})

if os.path.exists("data/supplement.yaml"):
    with open("data/supplement.yaml") as f:
        yaml=YAML(typ='safe')
        supplement_data = SupplementData(root=yaml.load(f))

api_client = APIClient(ak)

class SDData(BaseModel):
    stock: Stock
    prices: dict[int, StockPrice]
    ratios: dict[int, StockRatios]

stock_data: dict[str, SDData] = {}

ratio_bad_stocks: list[str] = []

supplemental_ratios: dict[str, dict[int, StockRatios]] = {}
for ticker, data in supplement_data.root.items():
    if len(data) == 0:
        continue # ignore the ticker
    
    for price, ratios in data.items():
        if ticker not in supplemental_ratios:
            supplemental_ratios[ticker] = {}

        supplemental_ratios[ticker][int(price)] = StockRatios(**ratios)

del ticker, data, company_ignore_list # Ensure we don't accidentally use this variable later

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

print("Stocks with unknown ratios:", ratio_bad_stocks, "count =", len(ratio_bad_stocks))