import uuid
from fastapi.responses import RedirectResponse
from pydantic import BaseModel
import redis
from admin.piccolo_app import APP_CONFIG
from admin.tables import Games, UserTransactions, Users as UserTable

from fastapi import FastAPI, Request
from piccolo_admin.endpoints import create_admin, FormConfig
from piccolo.engine import engine_finder
from starlette.routing import Mount
from argon2 import PasswordHasher
from xkcdpass import xkcd_password as xp
import random
import redis.asyncio as redis

redis_cli = redis.ConnectionPool.from_url("redis://localhost:6379/3")

def gen_pass() -> str:
    wordfile = xp.locate_wordfile()
    mywords = xp.generate_wordlist(wordfile=wordfile, min_length=5, max_length=8)
    return xp.generate_xkcdpassword(mywords, numwords=4, delimiter=" ")

def gen_random(length: int) -> str:
    """Generates a random alphanumeric string of sepcified length"""
    chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    return "".join([random.choice(chars) for _ in range(length)])

# Pydantic model for new user
class NewUser(BaseModel):
    new_username: str

    @staticmethod
    async def action(request: Request, data: "NewUser"):
        if not data.new_username:
            raise ValueError("Username cannot be empty")

        new_user = await UserTable.select(UserTable.username).where(UserTable.username == data.new_username).first().run()

        if new_user:
            raise ValueError("User with new username already exists")

        # Generate random 16 character password
        pwd = gen_pass()

        # Hash password
        ph = PasswordHasher()
        hashed_pwd = ph.hash(pwd)

        token = gen_random(512)

        # Create new user
        await UserTable.insert(
            UserTable(
                username=data.new_username,
                password=hashed_pwd,
                token=token,
            )
        )

        return "User created. Password is: " + pwd + "\n\nPlease save these values as they will not be shown again."

class ResetUserPassword(BaseModel):
    username: str
    new_password: str

    @staticmethod
    async def action(request: Request, data: "ResetUserPassword"):
        if not data.username:
            raise ValueError("Username cannot be empty")

        if not data.new_password:
            raise ValueError("Password cannot be empty")

        if len(data.new_password) < 8:
            raise ValueError("Password must be at least 8 characters long")

        # Find user
        user = await UserTable.select(UserTable.username).where(UserTable.username == data.username).first().run()

        if not user:
            raise ValueError("User does not exist")

        # Hash password
        ph = PasswordHasher()
        hashed_pwd = ph.hash(data.new_password)

        # Update user
        await UserTable.update(
            password=hashed_pwd
        ).where(UserTable.username == data.username).run()

        return "Password updated."

class RenameUser(BaseModel):
    old_username: str
    new_username: str

    @staticmethod
    async def action(request: Request, data: "RenameUser"):
        if not data.old_username:
            raise ValueError("Username cannot be empty")

        if not data.new_username:
            raise ValueError("Username cannot be empty")

        # Find user
        user = await UserTable.select(UserTable.username).where(UserTable.username == data.old_username).first().run()

        if not user:
            raise ValueError("User does not exist")

        new_user = await UserTable.select(UserTable.username).where(UserTable.username == data.new_username).first().run()

        if new_user:
            raise ValueError("User with new username already exists")

        # Update user
        await UserTable.update(
            username=data.new_username
        ).where(UserTable.username == data.old_username).run()

        return "Username updated."

class RedisClearCacheKey(BaseModel):
    key: str

    @staticmethod
    async def action(request: Request, data: "RedisClearCacheKey"):
        if not data.key:
            raise ValueError("Key cannot be empty")

        # Clear key
        conn = redis.Redis.from_pool(redis_cli)
        await conn.delete(data.key)

        return "Key cleared."

class FlushRedisCache(BaseModel):
    confirm: bool

    @staticmethod
    async def action(request: Request, data: "FlushRedisCache"):
        if not data.confirm:
            raise ValueError("Confirmation must be true")

        # Clear key
        conn = redis.Redis.from_pool(redis_cli)
        await conn.flushdb()

        return "Redis cache cleared."

class ClearUserTransactionsOfUser(BaseModel):
    user_id: uuid.UUID
    game_id: uuid.UUID
    pretend_mode: bool

    @staticmethod
    async def action(request: Request, data: "ClearUserTransactionsOfUser"):
        if not data.user_id:
            raise ValueError("User ID cannot be empty")

        if not data.game_id:
            raise ValueError("Game ID cannot be empty")

        user = await UserTable.select(UserTable.id).where(UserTable.id == data.user_id).first().run()

        if not user:
            raise ValueError("User ID does not exist. Please ensure that you are using the user ID [see the Users table] and not the username.")

        game = await Games.select(Games.id).where(Games.id == data.game_id).first().run()

        if not game:
            raise ValueError("Game ID does not exist. Please ensure that you are using the game ID [see the Games table] and not the game code.")

        # Delete transactions
        num_rows = await UserTransactions.count().where(UserTransactions.user_id == data.user_id, UserTransactions.game_id == data.game_id).run()

        if data.pretend_mode:
            raise ValueError(f"A total of {num_rows} transactions will be deleted. Please run this command again without the pretend_mode flag to confirm.")

        await UserTransactions.delete().where(UserTransactions.user_id == data.user_id, UserTransactions.game_id == data.game_id).run()
            

app = FastAPI(
    routes=[
        Mount(
            "/admin/",
            create_admin(
                tables=[t for t in APP_CONFIG.table_classes if t._meta.tablename not in ("sessions", "migration", "piccolo_user")],
                forms=[
                    FormConfig(
                        name="New User",
                        pydantic_model=NewUser,
                        endpoint=NewUser.action,
                        description="Creates a new user. Only use this to create users as using the users table directly will not hash the password."
                    ),
                    FormConfig(
                        name="Reset User Password",
                        pydantic_model=ResetUserPassword,
                        endpoint=ResetUserPassword.action,
                        description="Resets a user's password. Only use this to reset passwords as using the users table directly will not hash the password."
                    ),
                    FormConfig(
                        name="Rename User",
                        pydantic_model=RenameUser,
                        endpoint=RenameUser.action,
                        description="Renames a user. Only use this to rename users as using the users table directly will not hash the password."
                    ),
                    FormConfig(
                        name="Redis Clear Cache Key",
                        pydantic_model=RedisClearCacheKey,
                        endpoint=RedisClearCacheKey.action,
                        description="""
Clears a key from the Redis cache. E.g.

prior_stock_prices:{game_id}:{ticker} to clear the prior stock prices cache for a game and ticker.
(potential, not implemented yet) stock_list:{game_id}:?wpp={true/false} to clear the stock list cache for a game and whether to include prior prices for stocks.
"""
                    ),
                    FormConfig(
                        name="Flush Redis Cache",
                        pydantic_model=FlushRedisCache,
                        endpoint=FlushRedisCache.action,
                        description="Flushes the Redis cache. This will clear all cached data."
                    ),
                    FormConfig(
                        name="Clear User Transactions Of User",
                        pydantic_model=ClearUserTransactionsOfUser,
                        endpoint=ClearUserTransactionsOfUser.action,
                        description="WORK_IN_PROGRESS! Clears all transactions of a user in a game. Can be used for anti-spam purposes et al. Before committing, be sure to keep pretend_mode enabled to avoid irreversible data loss."
                    ),
                ],
                # Required when running under HTTPS:
                production=True,
                allowed_hosts=['stocksim2-admin.narc.live']
            )
        ),
    ],
)

@app.on_event("startup")
async def open_database_connection_pool():
    try:
        engine = engine_finder()
        await engine.start_connection_pool()
    except Exception:
        print("Unable to connect to the database")
        exit(1)


@app.on_event("shutdown")
async def close_database_connection_pool():
    try:
        engine = engine_finder()
        await engine.close_connection_pool()
        await redis_cli.aclose()
    except Exception:
        print("Unable to connect to the database")
        exit(1)

@app.get("/")
def admin_panel():
    return RedirectResponse("/admin") 