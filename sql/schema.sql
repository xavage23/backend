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
    description TEXT NOT NULL UNIQUE CHECK (description <> ''), -- Game description
    passphrase TEXT NOT NULL UNIQUE CHECK (passphrase <> ''), -- Game passphrase
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_price TEXT NOT NULL DEFAULT 'start', -- The current price of the stock, 'start' means start price and 'end' means end price
    initial_balance BIGINT NOT NULL -- The initial balance of a user in the game in cents
);

CREATE TABLE IF NOT EXISTS game_allowed_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);

CREATE TABLE IF NOT EXISTS game_user (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Game User token
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    balance BIGINT NOT NULL, -- The current balance of the user in cents
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);

CREATE TABLE IF NOT EXISTS stock (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    ticker TEXT NOT NULL, -- AAPL etc
    company_name TEXT NOT NULL,
    start_price BIGINT NOT NULL, -- The price of the stock at the start of the game in cents
    end_price BIGINT NOT NULL, -- The price of the stock at the end of the game in cents
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS news (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_transaction (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    game_id UUID NOT NULL REFERENCES games (id) ON UPDATE CASCADE ON DELETE CASCADE,
    stock_id UUID NOT NULL REFERENCES stock (id) ON UPDATE CASCADE ON DELETE CASCADE,
    stock_price BIGINT NOT NULL, -- The price of the stock at the time of the transaction in cents
    amount BIGINT NOT NULL, -- The amount of stocks bought/sold
    action TEXT NOT NULL, -- BUY or SELL
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
