import datetime
import os
import pathlib
from time import sleep
from typing import Any
import orjson
from pydantic import BaseModel
import requests
from utils import debug_print, yellow_print

class YahooAPIResponse(BaseModel):
    """A response from the Yahoo Finance API"""
    status_code: int
    content: str
    cached: bool

    def ok(self) -> bool:
        """Check if the response is OK (status code < 400 and >= 200)"""
        return self.status_code < 400 and self.status_code >= 200
    
    def to_json(self) -> dict:
        """Convert the response to JSON"""
        return orjson.loads(self.content)

class YahooAPIClient():
    """A client for the Yahoo Finance API"""
    _sess: requests.Session

    def __init__(self):
        self._sess = requests.Session()

        self._sess.headers = {
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0"
        }
    
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
    
    def cached_get(self, cache_key: str, url: str, params: dict[str, Any] = None) -> YahooAPIResponse:
        """Get a URL, but cache the response"""
        cached = self.find_in_cache(cache_key)
        if cached:
            return YahooAPIResponse(
                status_code=200,
                content=cached,
                cached=True
            )
        
        debug_print(f"YahooAPIClient.cached_get: Fetching {url} as not cached")

        res = self._sess.get(url, params=params)
        if res.ok:
            self.cache(cache_key, res.text)
        
        return YahooAPIResponse(
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
    def get_from_company_name(api_client: YahooAPIClient, company_name: str) -> list["Stock"]:
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
    def get_stock_price(api_client: YahooAPIClient, ticker: str, epoch_time: int):
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
        if sp.timestamp[0] <= epoch_time or sp.timestamp[-1] >= int(next_date.timestamp()):
            yellow_print(f"Timestamp {sp.timestamp} for stock prices found for {ticker} at {epoch_time} is out of range! Must be between {int(date.timestamp())} and {int(next_date.timestamp())}")

        return sp
