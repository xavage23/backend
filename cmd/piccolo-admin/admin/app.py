from fastapi.responses import RedirectResponse
from pydantic import BaseModel
from admin.piccolo_app import APP_CONFIG
from admin.tables import Users as UserTable

from fastapi import FastAPI, Request
from piccolo_admin.endpoints import create_admin, FormConfig
from piccolo.engine import engine_finder
from starlette.routing import Mount
from argon2 import PasswordHasher
from xkcdpass import xkcd_password as xp
import random

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
    except Exception:
        print("Unable to connect to the database")
        exit(1)

@app.get("/")
def admin_panel():
    return RedirectResponse("/admin") 