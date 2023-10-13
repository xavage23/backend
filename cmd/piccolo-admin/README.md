# Piccolo-Admin Panel

1. Panel tables are generated using ``piccolo schema generate > admin/tables.py``, then replacing `default="",` with nothing in the generated file, then making manual fixes
2. Run ``piccolo migrations forwards user`` to apply the migrations for users, then apply sessions migrations with ``piccolo migrations forwards session_auth``
3. If using systemd, a known working ExecCommand is ``/home/xavage/.local/bin/uvicorn admin.app:app --port 1922``