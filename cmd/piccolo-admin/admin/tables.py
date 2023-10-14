from enum import Enum
from piccolo.columns.base import OnDelete
from piccolo.columns.base import OnUpdate
from piccolo.columns.column_types import Boolean
from piccolo.columns.column_types import ForeignKey
from piccolo.columns.column_types import BigInt
from piccolo.columns.column_types import Serial
from piccolo.columns.column_types import Text
from piccolo.columns.column_types import Timestamp
from piccolo.columns.column_types import Timestamptz
from piccolo.columns.column_types import UUID
from piccolo.columns.column_types import Varchar
from piccolo.columns.column_types import Array
from piccolo.columns.column_types import Integer
from piccolo.columns.defaults.timestamp import TimestampNow
from piccolo.columns.defaults.timestamptz import TimestamptzNow
from piccolo.columns.defaults.uuid import UUID4
from piccolo.columns.indexes import IndexMethod
from piccolo.table import Table
from piccolo.columns.readable import Readable

class Games(Table, tablename="games"):
    @classmethod
    def get_readable(cls):
        return Readable(template="%s - %s", columns=[cls.id, cls.code])

    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    code = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_number = Integer(
        default=1,
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    description = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    enabled = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    trading_allowed = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    current_price_index = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        help_text="0 means first price, then keep incrementing",
    )
    initial_balance = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        help_text="Balance is in cents, not dollars",
    )


class Users(Table, tablename="users"):
    @classmethod
    def get_readable(cls):
        return Readable(template="%s - %s", columns=[cls.id, cls.username])

    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    username = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    password = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    token = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    enabled = Boolean(
        default=True,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    root = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )

class Migration(Table, tablename="migration"):
    id = Serial(
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    name = Varchar(
        length=200,
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    app_name = Varchar(
        length=200,
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    ran_on = Timestamp(
        default=TimestampNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class PiccoloUser(Table, tablename="piccolo_user"):
    id = Serial(
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    username = Varchar(
        length=100,
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    password = Varchar(
        length=255,
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    email = Varchar(
        length=255,
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    active = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    admin = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    first_name = Varchar(
        length=255,
        default="",
        null=True,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    last_name = Varchar(
        length=255,
        default="",
        null=True,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    superuser = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    last_login = Timestamp(
        default=TimestampNow(),
        null=True,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class Sessions(Table, tablename="sessions"):
    id = Serial(
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    token = Varchar(
        length=100,
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    user_id = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    expiry_date = Timestamp(
        default=TimestampNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    max_expiry_date = Timestamp(
        default=TimestampNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class News(Table, tablename="news"):
    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    title = Text(
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class Stocks(Table, tablename="stocks"):
    @classmethod
    def get_readable(cls):
        return Readable(template="%s - %s (%s) %s", columns=[cls.id, cls.ticker, cls.company_name, cls.game_id])

    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    ticker = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    company_name = Text(
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    description = Text(
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    prices = Array(
        base_column=BigInt(),
        default=[],
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        help_text="Price is in cents, not dollars"
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class GameAllowedUsers(Table, tablename="game_allowed_users"):
    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    user_id = ForeignKey(
        references=Users,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class GameUsers(Table, tablename="game_users"):
    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    user_id = ForeignKey(
        references=Users,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    initial_balance = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        help_text="This is the initial balance of the user. Balance is in cents, not dollars"
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )

class UserTransactions(Table, tablename="user_transactions"):
    @classmethod
    def get_readable(cls):
        return Readable(template="%s [%s %s] by %s", columns=[cls.id, cls.action, cls.stock_id, cls.user_id])

    class UserTransactionAction(str, Enum):
        buy = "buy"
        sell = "sell"

    id = UUID(
        default=UUID4(),
        null=False,
        primary_key=True,
        unique=False,
        index=True,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    user_id = ForeignKey(
        references=Users,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    stock_id = ForeignKey(
        references=Stocks,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column="id",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    price_index = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        price="The index of the price to use, this should be the same as the current_price_index when the transaction was created"
    )
    sale_price = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        price="The price of the stock at the time of the transaction, in cents, not dollars"
    )
    amount = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    action = Text(
        default="",
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
        choices=UserTransactionAction
    )
    created_at = Timestamptz(
        default=TimestamptzNow(),
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )

