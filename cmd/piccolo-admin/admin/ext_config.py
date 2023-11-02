import datetime
from piccolo_admin.endpoints import TableConfig
from piccolo_api.crud.hooks import Hook, HookType
from admin.tables import Games, Users

async def save_game_hook(row: Games):
    if row.enabled:
        row.enabled = datetime.datetime.now(tz=datetime.timezone.utc)
    return row.enabled

async def patch_game_hook(row_id, values: dict):
    if values.get("enabled"):
        values["enabled"] = datetime.datetime.now(tz=datetime.timezone.utc)

    return values

async def just_raise(*args, **kwargs):
    raise ValueError("This action is not allowed")

game_config = TableConfig(
    Games,
    hooks=[
        Hook(hook_type=HookType.pre_save, callable=save_game_hook),
        Hook(hook_type=HookType.pre_patch, callable=patch_game_hook),
    ]
)

user_config = TableConfig(
    Users,
    hooks=[
        Hook(hook_type=HookType.pre_save, callable=just_raise),
        Hook(hook_type=HookType.pre_patch, callable=just_raise),
    ]
)
