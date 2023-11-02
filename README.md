# Backend for Bulls and Bears

Bulls and Bears (AKA ``bb``) is a game at XAVAGE 23.

## Setup

1. Create a postgres database called ``xavage``
2. Run ``psql``, then restore ``sql/schema.sql`` into the database using ``\i sql/schema.sql``
3. Run ``make`` here (make sure you have Go 1.20+ installed), then run ``xavagebb``. This will create a ``config.yaml`` file before erroring. Fill in the config file with the configuration you wish to use (server port etc.)
4. Run ``xavagebb`` again. It should now work and you can now persist this using tmux or systemd etc.

## Admin Interface

### admintool-cli

Run ``make`` in ``cmd/admintool-cli`` to build the admin tool CLI. This can be used to create users etc (basically admin tooling).

#### Creating multiple users

``STRIP_SPECIFIC_CHARS="()" ./admintool-cli  users createmultiple``

The ``STRIP_SPECIFIC_CHARS`` environment variable is used to strip specific characters from the usernames. This is useful if you have a list of usernames enclosed in brackets such as what occurred in XAVAGE23 itself

### piccolo-admin

Piccolo Admin is an admin panel for Piccolo ORM. It can be used to view and edit the database. To run it, see ``cmd/piccolo-admin``.