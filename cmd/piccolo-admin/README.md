# Piccolo-Admin Panel

1. Run ``piccolo migrations forwards user`` to apply the migrations for users, then apply sessions migrations with ``piccolo migrations forwards session_auth``
2. If using systemd, a known working ExecCommand is ``/home/xavage/.local/bin/uvicorn admin.app:app --port 1922``

## Note

Panel tables were generated using ``piccolo schema generate > admin/tables.py``, then replacing `default="",` with nothing in the generated file, then making manual fixes

To create a schema diff, do piccolo schema generate > admin/tables.py