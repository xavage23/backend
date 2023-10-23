import datetime
import os
import pathlib
from time import sleep
from typing import Any
import orjson
from pydantic import BaseModel
import requests
from utils import debug_print, green_print, yellow_print

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
    debt_to_equity_ratio: float
    profit_margin: float

    @staticmethod
    def get_stock_ratios_for_time(api_client: APIClient, stock: Stock, prices: dict[int, StockPrice]):
        # This one is a bit painful to get
        # We essentially need to calculate the ratios ourselves
    
        # Step 1. Get EPS. This allows us to take care of both PE ratio and EPS
        # We do this using the alpha_vantage API
        #
        # Note that as we are using the EARNINGS function, we can cache this for all our stocks
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

                av_earnings = _AVEarnings(**json)

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
                        raise ValueError(f"No annual earnings found for {stock.symbol} at {epoch_time}: {res.content} [status code: {res.status_code}]")
                    
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

                    green_print("Earning DT:", earning_dt.strftime("%d/%m/%Y, %H:%M:%S"), "\nShare Price (normal):", sp.open, "\nShare Price (on earning)", sp_n.close, "\nEPS:", annual_earnings.reportedEPS, "\nPE ratio:", pe_ratio, "\nSymbol:", stock.symbol, "Date:", dt.strftime("%d/%m/%Y, %H:%M:%S"), "\nTimestamp:", epoch_time, "\n")

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

class _AVEarnings(BaseModel):
    annualEarnings: list[_AVAnnualEarning]
    quarterlyEarnings: list[_AVQuarterlyEarning]

# Internal class to allow seeing which stock exchanges we need to add
class BadStockExchangeException(Exception):
    pass