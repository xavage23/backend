from fastapi.responses import HTMLResponse
from admin.piccolo_app import APP_CONFIG

from fastapi import FastAPI
from piccolo_admin.endpoints import create_admin
from piccolo.engine import engine_finder
from starlette.routing import Mount

with open("admin/index.html", "r") as f:
    index_html = f.read()

app = FastAPI(
    routes=[
        Mount(
            "/admin/",
            create_admin(
                tables=[t for t in APP_CONFIG.table_classes if t._meta.tablename not in ("sessions", "migration", "piccolo_user")],
                # Required when running under HTTPS:
                production=True,
                allowed_hosts=['stocksim2-admin.narc.live']
            )
        ),
    ],
)

@app.get("/")
def admin_panel():
    return HTMLResponse(index_html)

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
