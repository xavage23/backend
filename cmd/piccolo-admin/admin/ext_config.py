import datetime
from piccolo_admin.endpoints import TableConfig
from piccolo_api.crud.hooks import Hook, HookType
from admin.tables import Games, Users

async def save_game_hook(row: Games):
    if row.enabled:
        row.enabled = datetime.datetime.now(tz=datetime.timezone.utc)
    return row.enabled

async def patch_game_hook(row_id, values: dict):
    print(row_id, values)
    if values.get("enabled"):
        # Check current status of game for enabled status
        game = await Games.objects().where(Games.id == row_id).first().run()

        if game.enabled:
            print(values["enabled"], game.enabled)
            values["enabled"] = game.enabled
            return values
        else:
            values["enabled"] = None # Ensure games aren't enabled by updates outside of form
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
