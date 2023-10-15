create extension if not exists "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL UNIQUE, -- Argon2id hashed password
    token TEXT NOT NULL UNIQUE, -- API token
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    root BOOLEAN NOT NULL DEFAULT FALSE, -- Whether or not the user is 'root' (admin)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- needed by piccolo
    code TEXT NOT NULL UNIQUE CHECK (code <> ''), -- Game code
    game_number INTEGER NOT NULL DEFAULT 1, -- Game number, all games below this game number will have their transactions moved to this game upon joining
    name TEXT NOT NULL UNIQUE CHECK (name <> ''), -- Game description
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    trading_allowed BOOLEAN NOT NULL DEFAULT FALSE,
    old_stocks_carry_over BOOLEAN NOT NULL DEFAULT TRUE, -- Whether or not stocks from previous games must carry over
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_price_index INTEGER NOT NULL DEFAULT 0, -- The current price index of the game
    initial_balance BIGINT NOT NULL -- The initial balance of a user in the game in cents
);

CREATE TABLE IF NOT EXISTS game_allowed_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);

CREATE TABLE IF NOT EXISTS game_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Game User token
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    initial_balance BIGINT NOT NULL, -- The initial balance of the user in cents
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);

CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    ticker TEXT NOT NULL, -- AAPL etc
    company_name TEXT NOT NULL,
    description TEXT NOT NULL,
    prices BIGINT[] NOT NULL, -- The prices of the stock in cents
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, ticker)
);

CREATE TABLE IF NOT EXISTS news (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    origin_game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    stock_id UUID NOT NULL REFERENCES stock (id) ON UPDATE CASCADE ON DELETE CASCADE,
    price_index INTEGER NOT NULL, -- The price of the stock at the time of the transaction in cents
    sale_price BIGINT NOT NULL, -- The price of the stock at the time of the transaction in cents
    amount BIGINT NOT NULL, -- The amount of stocks bought/sold
    action TEXT NOT NULL, -- BUY or SELL
    past BOOLEAN NOT NULL DEFAULT FALSE, -- Whether or not the transaction is in the past (from a previous game)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
