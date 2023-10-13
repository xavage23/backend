from piccolo.columns.base import OnDelete
from piccolo.columns.base import OnUpdate
from piccolo.columns.column_types import Boolean
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


class Games(Table, tablename="games"):
    code = Text(
        default="",
        null=False,
        primary_key=True,
        unique=False,
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
    game_started = Timestamptz(
        default=TimestamptzNow(),
        null=True,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    initial_balance = Integer(
        default=0,
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


class GameUser(Table, tablename="game_user"):
    id = Text(
        default="",
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
    game_code = ForeignKey(
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
    balance = Integer(
        default=0,
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
    game_code = ForeignKey(
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


class Stock(Table, tablename="stock"):
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
    game_code = ForeignKey(
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
    ticker = Text(
        default="",
        null=False,
        primary_key=False,
        unique=False,
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
    start_price = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    end_price = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    current_price = Text(
        default="start",
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


class UserTransaction(Table, tablename="user_transaction"):
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
    game_code = ForeignKey(
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
        references=Stock,
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
    stock_price = Integer(
        default=0,
        null=False,
        primary_key=False,
        unique=False,
        index=False,
        index_method=IndexMethod.btree,
        db_column_name=None,
        secret=False,
    )
    amount = Integer(
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

