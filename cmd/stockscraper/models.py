import datetime
import os
import pathlib
from time import sleep
import orjson
from pydantic import BaseModel, RootModel
import requests
from utils import debug_print, green_print, red_print, yellow_print

class SupplementData(RootModel):
    root: dict[str, dict[int, dict[str, float]]]

class APIResponse(BaseModel):
    """A response from an API"""
    status_code: int
    content: str
    cached: bool

    def ok(self) -> bool:
        """Check if the response is OK (status code < 400 and >= 200)"""
        return self.status_code < 400 and self.status_code >= 200
    
    def to_json(self) -> dict:
        """Convert the response to JSON"""
        return orjson.loads(self.content)

class APIClient():
    """A client for the stockscraper API"""
    _sess: requests.Session
    alpha_vantage_key: str

    def __init__(self, alpha_vantage_key: str):
        self.alpha_vantage_key = alpha_vantage_key
        self._sess = requests.Session()

        self._sess.headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
        }
    
    def clear_cache(self, key: str): 
        pathlib.Path("cache").mkdir(parents=True, exist_ok=True)

        if os.path.exists(f"cache/{key}"):
            os.remove(f"cache/{key}")
    
    def find_in_cache(self, key: str) -> str | None:
        """Find a file in the cache folder"""
        if not os.path.exists(f"cache/{key}"):
            debug_print("APIClient.find_in_cache: Not found in cache:", key)
            return None
        
        with open(f"cache/{key}", "r") as f:
            return f.read()
    
    def cache(self, key: str, value: str) -> None:
        """Cache a file in the cache folder"""
        pathlib.Path("cache").mkdir(parents=True, exist_ok=True)

        with open(f"cache/{key}", "w") as f:
            f.write(value)
    
    def cached_get(self, cache_key: str, url: str, **kwargs) -> APIResponse:
        """Get a URL, but cache the response"""
        cached = self.find_in_cache(cache_key)
        if cached:
            return APIResponse(
                status_code=200,
                content=cached,
                cached=True
            )
        
        debug_print(f"APIClient.cached_get: Fetching {url} as not cached")

        res = self._sess.get(url, **kwargs)
        if res.ok:
            self.cache(cache_key, res.text)
        
        return APIResponse(
            status_code=res.status_code,
            content=res.text,
            cached=False
        )

class Stock(BaseModel):
    """A stock from Yahoo Finance"""
    exchange: str
    shortname: str
    quoteType: str | None = None
    symbol: str
    index: str 
    score: float | None = None
    typeDisp: str
    longname: str | None = None
    exchDisp: str
    sector: str | None = None
    industry: str | None = None
    industryDisp: str | None = None
    dispSecIndFlag: bool | None = None
    isYahooFinance: bool

    @staticmethod
    def get_from_company_name(api_client: APIClient, company_name: str) -> list["Stock"]:
        """Get a list of stocks from a company name"""
        res = api_client.cached_get(f"tickerMap@{company_name}", f"https://query2.finance.yahoo.com/v1/finance/search?q={company_name}")
        if not res.ok():
            raise ValueError(f"Failed to fetch ticker symbol for {company_name}: {res.text} [status code: {res.status_code}]")

        return [Stock(**stock) for stock in res.to_json()["quotes"]]
    
class StockPrice(BaseModel):
    """A stock price from Yahoo Finance"""
    open: float
    close: float
    high: float
    low: float
    volume: int
    adjclose: float | None = None
    timestamp: list[int] | None = None

    @staticmethod
    def get_stock_price(api_client: APIClient, ticker: str, epoch_time: int):
        # Convert the Unix epoch time to a human-readable date format
        date = datetime.datetime.utcfromtimestamp(epoch_time)
        next_date = datetime.datetime.utcfromtimestamp(epoch_time + 86400)

        # Define the URL for Yahoo Finance API
        url = f'https://query2.finance.yahoo.com/v8/finance/chart/{ticker}?period1={int(date.timestamp())}&period2={int(next_date.timestamp())}&interval=1d'
        
        # Send the GET request to Yahoo Finance API
        res = api_client.cached_get(f"{ticker}@{epoch_time}", url)

        if not res.ok():
            raise ValueError(f"Failed to fetch stock price for {ticker} at {epoch_time}: {res.content} [status code: {res.status_code}]")

        if not res.cached:
            sleep(1)

        # Parse the JSON response
        res_json = res.to_json()

        if not res_json.get("chart", {}).get("result", []):
            raise ValueError(f"No stock prices found for {ticker} at {epoch_time}: {res.content} [status code: {res.status_code}]")

        sp = StockPrice(
            open=res_json["chart"]["result"][0]["indicators"]["quote"][0]["open"][0],
            close=res_json["chart"]["result"][0]["indicators"]["quote"][0]["close"][0],
            high=res_json["chart"]["result"][0]["indicators"]["quote"][0]["high"][0],
            low=res_json["chart"]["result"][0]["indicators"]["quote"][0]["low"][0],
            volume=res_json["chart"]["result"][0]["indicators"]["quote"][0]["volume"][0],
            adjclose=res_json["chart"]["result"][0]["indicators"]["adjclose"][0]["adjclose"][0],
            date=epoch_time,
            timestamp=res_json["chart"]["result"][0]["timestamp"]
        )

        if not sp.timestamp:
            raise ValueError(f"No timestamp for stock prices found for {ticker} at {epoch_time}: {res.content} [status code: {res.status_code}]")

        # Ensure sp.timestamp is in range of date and next_date
        if sp.timestamp[0] <= epoch_time:
            # Subtract and take abs to ensure its not a false positive
            if abs(sp.timestamp[0] - epoch_time) > 86400:
                yellow_print(f"Timestamp {sp.timestamp} for stock prices found for {ticker} at {epoch_time} is out of range! Must be between {int(date.timestamp())} and {int(next_date.timestamp())}")
            
        if sp.timestamp[-1] >= int(next_date.timestamp()):
            # Subtract and take abs to ensure its not a false positive
            if abs(sp.timestamp[-1] - int(next_date.timestamp())) > 86400:
                yellow_print(f"Timestamp {sp.timestamp} for stock prices found for {ticker} at {epoch_time} is out of range! Must be between {int(date.timestamp())} and {int(next_date.timestamp())}")

        return sp

class StockRatios(BaseModel):
    pe_ratio: float
    earnings_per_share: float
    debt_to_equity_ratio: float | None = None
    profit_margin: float | None = None

    @staticmethod
    def get_stock_ratios_for_time(api_client: APIClient, stock: Stock, prices: dict[int, StockPrice]):
        # This one is a bit painful to get
        # We essentially need to calculate the ratios ourselves
    
        match stock.exchDisp:
            case "NASDAQ" | "NYSE":
                res = api_client.cached_get(f"{stock.symbol}@EPS", f"https://www.alphavantage.co/query?function=EARNINGS&symbol={stock.symbol}&apikey={api_client.alpha_vantage_key}")

                if not res.ok():
                    raise ValueError(f"Failed to fetch EPS for {stock.symbol}: {res.content} [status code: {res.status_code}]")

                json = res.to_json()

                if json.get("Note"):
                    api_client.clear_cache(f"{stock.symbol}@EPS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 minute: {json}")
                    sleep(60 * 1)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)

                if json.get("Information"):
                    api_client.clear_cache(f"{stock.symbol}@EPS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 day: {json}")
                    sleep(86400)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)

                av_earnings = _AVEarnings(**json)

                # We also need the balance sheet of the stock, fetch that as well
                res_bs = api_client.cached_get(f"{stock.symbol}@BS", f"https://www.alphavantage.co/query?function=BALANCE_SHEET&symbol={stock.symbol}&apikey={api_client.alpha_vantage_key}")

                if not res_bs.ok():
                    raise ValueError(f"Failed to fetch balance sheet for {stock.symbol}: {res_bs.content} [status code: {res_bs.status_code}]")

                json_bs = res_bs.to_json()

                if json_bs.get("Note"):
                    api_client.clear_cache(f"{stock.symbol}@BS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 minute: {json_bs}")
                    sleep(60 * 1)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)

                if json_bs.get("Information"):
                    api_client.clear_cache(f"{stock.symbol}@BS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 day: {json_bs}")
                    sleep(86400)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)
                
                # Parse the balance sheet
                parsed_json_bs = []
                
                for bs in json_bs.get("quarterlyReports", []):
                    bs_parsed = {}
                    for k, v in bs.items():
                        if k not in ["fiscalDateEnding", "reportedCurrency"]:
                            if v == "None":
                                bs_parsed[k] = None
                            else:
                                bs_parsed[k] = float(v)
                        else:
                            bs_parsed[k] = v
                    
                    parsed_json_bs.append(bs_parsed)
                
                balance_sheet = _AVBalanceSheet(reports=parsed_json_bs)

                # Lastly fetch the income statement as well
                res_is = api_client.cached_get(f"{stock.symbol}@IS", f"https://www.alphavantage.co/query?function=INCOME_STATEMENT&symbol={stock.symbol}&apikey={api_client.alpha_vantage_key}")

                if not res_is.ok():
                    raise ValueError(f"Failed to fetch income statement for {stock.symbol}: {res_is.content} [status code: {res_is.status_code}]")

                json_is = res_is.to_json()

                if json_is.get("Note"):
                    api_client.clear_cache(f"{stock.symbol}@IS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 minute: {json_is}")
                    sleep(60 * 1)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)
                
                if json_is.get("Information"):
                    api_client.clear_cache(f"{stock.symbol}@IS")
                    yellow_print(f"Alpha Vantage API rate limit reached! Waiting 1 day: {json_is}")
                    sleep(86400)
                    return StockRatios.get_stock_ratios_for_time(api_client, stock, prices)

                # Parse the income sheet
                parsed_json_is = []
                
                for iss in json_is.get("quarterlyReports", []):
                    is_parsed = {}
                    for k, v in iss.items():
                        if k not in ["fiscalDateEnding", "reportedCurrency"]:
                            if v == "None":
                                is_parsed[k] = None
                            else:
                                is_parsed[k] = float(v)
                        else:
                            is_parsed[k] = v
                    
                    parsed_json_is.append(is_parsed)

                income_sheet = _AVIncomeStatement(reports=parsed_json_is)

                srs: dict[int, StockRatios] = {}
                for epoch_time, sp in prices.items():
                    dt = datetime.datetime.utcfromtimestamp(epoch_time)

                    # Find, for the time period, the annual earnings
                    annual_earnings: _AVAnnualEarning | None = None
                    earning_dt: datetime.datetime = None

                    for earnings in av_earnings.annualEarnings:
                        earning_dt = datetime.datetime.strptime(earnings.fiscalDateEnding, "%Y-%m-%d")

                        if earning_dt.year == dt.year and earnings.reportedEPS:
                            annual_earnings = earnings
                            break
                    
                    if not annual_earnings:
                        raise ValueError(f"No annual earnings found for {stock.symbol} at {epoch_time}")
                    
                    # Now we have the EPS, we can calculate the PE ratio
                    # To ensure some level of accuracy, get the stock price on the day of the reported EPS
                    sp_n: StockPrice | None = None

                    taken_ts = []
                    while not sp_n:
                        # Ensure that we are not a weekend
                        # If saturday, move to friday (-1 days)
                        # If sunday, move to friday (-2 day)
                        if earning_dt.weekday() == 5:
                            yellow_print(int(earning_dt.timestamp()), "is a saturday, taking previous day (friday)")
                            earning_dt += datetime.timedelta(days=-1)
                        elif earning_dt.weekday() == 6:
                            yellow_print(int(earning_dt.timestamp()), "is a sunday, taking 2 previous days (friday)")
                            earning_dt += datetime.timedelta(days=-2)

                        while int(earning_dt.timestamp()) in taken_ts:
                            earning_dt += datetime.timedelta(days=1)

                        taken_ts.append(int(earning_dt.timestamp()))

                        try:
                            sp_n = StockPrice.get_stock_price(api_client, stock.symbol, int(earning_dt.timestamp()))
                        except KeyError:
                            yellow_print("Failed to fetch stock price for", stock.symbol, "at", int(earning_dt.timestamp()), "taking next day")
                            while int(earning_dt.timestamp()) in taken_ts:
                                earning_dt += datetime.timedelta(days=1)
                            debug_print("New date:", earning_dt.strftime("%d/%m/%Y, %H:%M:%S"))
                        
                    pe_ratio = sp_n.close / annual_earnings.reportedEPS

                    green_print("Earning DT:", earning_dt.strftime("%d/%m/%Y, %H:%M:%S"), "\nShare Price (normal):", sp.open, "\nShare Price (on earning)", sp_n.close, "\nEPS:", annual_earnings.reportedEPS, "\nPE ratio:", pe_ratio, "\nSymbol:", stock.symbol, "\nDate:", dt.strftime("%d/%m/%Y, %H:%M:%S"), "\nTimestamp:", epoch_time, "\n")

                    debt_to_equity_ratio: int | None = None
                    profit_margin: int | None = None

                    # Find, for the time period, the balance sheet
                    balance_sheet_filtered: list[_AVBalanceSheetData] = []

                    for bs in balance_sheet.reports:
                        bs_dt = datetime.datetime.strptime(bs.fiscalDateEnding, "%Y-%m-%d")

                        if bs_dt.year == dt.year:
                            balance_sheet_filtered.append(bs)
                    
                    if not balance_sheet_filtered:
                        red_print("No balance sheet found for", stock.symbol, "at", epoch_time, ". Debt to equity ratio will None")
                        sleep(1)
                    
                    # Find, for the time period, the income statement
                    income_sheet_filtered: list[_AVIncomeSheetData] = []

                    for iss in income_sheet.reports:
                        iss_dt = datetime.datetime.strptime(iss.fiscalDateEnding, "%Y-%m-%d")

                        if iss_dt.year == dt.year:
                            income_sheet_filtered.append(iss)

                    if not income_sheet_filtered:
                        red_print("No income statement found for", stock.symbol, "at", epoch_time, ". Profit margin will None")
                        sleep(1)

                    debug_print(f"Have {len(balance_sheet_filtered)} [{balance_sheet_filtered}] balance sheets for {stock.symbol} at {epoch_time}")

                    for bs in balance_sheet_filtered:
                        if bs.totalShareholderEquity and (bs.shortTermDebt or bs.longTermDebt):
                            total_debt = (bs.shortTermDebt or 0) + (bs.longTermDebt or 0)
                            debt_to_equity_ratio = total_debt / bs.totalShareholderEquity
                            break
                    
                    if not debt_to_equity_ratio:
                        red_print("No debt found for", stock.symbol, "at", epoch_time, ". Debt to equity ratio will be None")
                        sleep(1)

                    debug_print(f"Have {len(income_sheet_filtered)} [{income_sheet_filtered}] income sheets for {stock.symbol} at {epoch_time}")

                    for iss in income_sheet_filtered:
                        if iss.netIncome and iss.totalRevenue:
                            profit_margin = (iss.netIncome / iss.totalRevenue) * 100
                            break
                        elif iss.costofGoodsAndServicesSold and iss.totalRevenue:
                            profit_margin = ((iss.totalRevenue - iss.costofGoodsAndServicesSold) / iss.totalRevenue) * 100
                            break
                    
                    if not profit_margin:
                        red_print("No profit margin found for", stock.symbol, "at", epoch_time, ". Profit margin will be None")
                        sleep(1)
                    
                    green_print("Debt to equity ratio:", debt_to_equity_ratio, "\nProfit margin:", profit_margin, "\nSymbol:", stock.symbol, "\nDate:", dt.strftime("%d/%m/%Y, %H:%M:%S"), "\nTimestamp:", epoch_time, "\n")

                    srs[epoch_time] = StockRatios(
                        pe_ratio=pe_ratio,
                        earnings_per_share=annual_earnings.reportedEPS,
                        debt_to_equity_ratio=debt_to_equity_ratio,
                        profit_margin=profit_margin
                    )
                
                return srs
            case _:
                yellow_print(f"Stock exchange {stock.exchDisp} not implemented yet")
                raise BadStockExchangeException("Stock exchange not implemented yet")

# Internal classes to represent datasets from sources
class _AVAnnualEarning(BaseModel):
    fiscalDateEnding: str
    reportedEPS: float

class _AVQuarterlyEarning(BaseModel):
    fiscalDateEnding: str
    reportedDate: str
    reportedEPS: float | str | None = None
    estimatedEPS: float | str
    surprise: float | str
    surprisePercentage: float | str

# Balance sheet
class _AVEarnings(BaseModel):
    annualEarnings: list[_AVAnnualEarning]
    quarterlyEarnings: list[_AVQuarterlyEarning]

class _AVBalanceSheetData(BaseModel):
    fiscalDateEnding: str
    commonStockSharesOutstanding: float | None = None
    shortTermDebt: float | None = None
    longTermDebt: float | None = None
    totalShareholderEquity: float | None = None

# Note: this only includes the fields we need
class _AVBalanceSheet(BaseModel):
    reports: list[_AVBalanceSheetData]

# Income sheet
class _AVIncomeSheetData(BaseModel):
    fiscalDateEnding: str
    netIncome: float | None = None
    totalRevenue: float | None = None
    costofGoodsAndServicesSold: float | None = None

# Note: this only includes the fields we need
class _AVIncomeStatement(BaseModel):
    reports: list[_AVIncomeSheetData]

# Internal class to allow seeing which stock exchanges we need to add
class BadStockExchangeException(Exception):
    pass

# Internal class to represent a stock and its data
class SDData(BaseModel):
    stock: Stock
    prices: dict[int, StockPrice]
    ratios: dict[int, StockRatios]

# Internal class to represent the round map file
class RoundMap(BaseModel):
    identify_column_name: str
    identify_column_value: str
    time_indexes: list[int]

    def display(self) -> str:
        return f"{self.identify_column_name}:{self.identify_column_value} ({', '.join([str(i) for i in self.time_indexes])})"

    @staticmethod
    def from_line(line: str) -> "RoundMap":
        line_split = line.split(" ")

        if not line_split:
            raise ValueError(f"Failed to parse line '{line}' in roundmap")
        
        identify_column = line_split[0]

        identify_column_split = identify_column.split(":")

        if len(identify_column_split) != 2:
            raise ValueError(f"Line '{line}' in roundmap does not have a valid identify column")

        return RoundMap(
            identify_column_name=identify_column_split[0],
            identify_column_value=identify_column_split[1],
            time_indexes=[int(i) for i in line_split[1:]]
        )

# Internal classes to represent the output YAML import file
class ImportStockRatio(BaseModel):
    name: str
    value_text: str | None = None
    value: float
    price_index: int

class ImportStock(BaseModel):
    ticker: str
    company_name: str
    prices: list[int]
    description: str
    ratios: list[ImportStockRatio]

class ImportGame(BaseModel):
    identify_column_name: str
    identify_column_value: str
    price_times: list[int]
    stocks: list[ImportStock]

class ImportFile(BaseModel):
    games: list[ImportGame]