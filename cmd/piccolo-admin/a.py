from piccolo.columns.base import OnDelete
from piccolo.columns.base import OnUpdate
from piccolo.columns.column_types import BigInt
from piccolo.columns.column_types import Boolean
from piccolo.columns.column_types import Column
from piccolo.columns.column_types import ForeignKey
from piccolo.columns.column_types import Integer
from piccolo.columns.column_types import Serial
from piccolo.columns.column_types import Text
from piccolo.columns.column_types import Timestamp
from piccolo.columns.column_types import Timestamptz
from piccolo.columns.column_types import UUID
from piccolo.columns.column_types import Varchar
from piccolo.columns.defaults.timestamp import TimestampNow
from piccolo.columns.defaults.timestamptz import TimestamptzNow
from piccolo.columns.defaults.uuid import UUID4
from piccolo.columns.indexes import IndexMethod
from piccolo.table import Table


class Users(Table, tablename="users"):
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


class Games(Table, tablename="games"):
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
    initial_balance = BigInt(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    name = Text(
        default="",
        null=False,
        primary_key=False,
        unique=True,
        index=True,
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
    current_price_index = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_number = Integer(
        default=1,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    old_stocks_carry_over = Boolean(
        default=True,
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
    user_id = Integer(
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
    user_id = ForeignKey(
        references=Users,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column=None,
        null=False,
        primary_key=False,
        unique=True,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column=None,
        null=False,
        primary_key=False,
        unique=True,
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
        target_column=None,
        null=False,
        primary_key=False,
        unique=True,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.cascade,
        on_update=OnUpdate.cascade,
        target_column=None,
        null=False,
        primary_key=False,
        unique=True,
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
    initial_balance = BigInt(
        default=0,
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
        target_column=None,
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
        target_column=None,
        null=False,
        primary_key=False,
        unique=True,
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
    prices = Column(
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


class UserTransactions(Table, tablename="user_transactions"):
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
        target_column=None,
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
        target_column=None,
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
        target_column=None,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
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
    price_index = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
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
    )
    past = Boolean(
        default=False,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    origin_game_id = ForeignKey(
        references=Games,
        on_delete=OnDelete.restrict,
        on_update=OnUpdate.cascade,
        target_column=None,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )


"""
WARNING: Unrecognised column types, added `Column` as a placeholder:
stocks.prices ['ARRAY']
stocks.prices ['ARRAY']
"""
"""
WARNING: Unable to parse the following indexes:
CREATE UNIQUE INDEX game_allowed_users_user_id_game_id_key ON public.game_allowed_users USING btree (user_id, game_id);
CREATE UNIQUE INDEX game_user_user_id_game_id_key ON public.game_users USING btree (user_id, game_id);
CREATE UNIQUE INDEX gidtik_unique ON public.stocks USING btree (game_id, ticker);
"""

