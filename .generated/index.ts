// Code generated by tygo. DO NOT EDIT.

//////////
// source: assets.go

export interface AssetMetadata {
  exists: boolean;
  path?: string;
  default_path: string;
  type?: string;
  size?: number /* int64 */;
  last_modified?: string /* RFC3339 */;
  errors?: string[];
}

//////////
// source: auth.go

export interface UserLogin {
  username: string;
  password: string;
}
export interface UserLoginResponse {
  user_id: string;
  token: string;
}

//////////
// source: common.go

/**
 * This represents a IBL xavagebb API Error
 */
export interface ApiError {
  context?: { [key: string]: string};
  message: string;
}
/**
 * Paged result common
 */
export interface PagedResult<T extends any> {
  count: number /* uint64 */;
  per_page: number /* uint64 */;
  results: T;
}

//////////
// source: game.go

export type GameMigrationMethod = string;
export const GameMigrationMethodMoveEntireTransactionHistory: GameMigrationMethod = "move_entire_transaction_history";
export const GameMigrationMethodCondensedMigration: GameMigrationMethod = "condensed_migration";
export const GameMigrationMethodNoMigration: GameMigrationMethod = "no_migration";
export interface Game {
  id: string;
  code: string;
  enabled: string | null /* RFC3339, nullable */;
  initially_enabled: string | null /* RFC3339, nullable */;
  trading_enabled: boolean;
  name: string;
  created_at: string /* RFC3339 */;
  price_times: string | null /* RFC3339, nullable */[];
  current_price_index: number /* int */;
  initial_balance: number /* int64 */;
  game_number: number /* int */;
  old_stocks_carry_over: boolean;
  game_migration_method: GameMigrationMethod;
  publicly_listed: boolean;
}
export interface AvailableGame {
  game: Game;
  can_join: boolean;
  is_enabled: boolean;
}
export interface GameJoinRequest {
  game_code: string;
}
export interface GameJoinResponse {
  id: string;
  new: boolean;
}
export interface GameUser {
  id: string;
  user_id: string;
  game_id: string;
  game: Game;
  initial_balance: number /* int64 */;
  current_balance: number /* int64 */;
  created_at: string /* RFC3339 */;
}
export interface Leaderboard {
  user?: User;
  initial_balance: number /* int64 */;
  current_balance: number /* int64 */;
  portfolio_value: number /* int64 */;
}

//////////
// source: stock.go

export interface StockRatio {
  id: string;
  name: string;
  value_text: string | null /* nullable */;
  value: number /* nullable */;
  price_index: number /* int */;
}
export interface PriorPricePoint {
  prices: number /* int64 */[];
  game: Game;
}
export interface Stock {
  id: string;
  game_id: string;
  ticker: string;
  company_name: string;
  description: string;
  current_price: number /* int64 */;
  known_prices: number /* int64 */[];
  prior_prices: PriorPricePoint[];
  created_at: string /* RFC3339 */;
  known_ratios: KnownRatios[];
  prior_ratios: PriorRatios[];
  includes?: string[];
}
export interface KnownRatios {
  ratios: StockRatio[];
  price_index: number /* int */;
}
export interface PriorRatios {
  ratios: StockRatio[];
  price_index: number /* int */;
  game: Game;
}
export interface StockList {
  stocks: (Stock | undefined)[];
  price_index: number /* int */;
}
export interface News {
  id: string;
  title: string;
  description: string;
  published: boolean;
  affected_stock_id: string /* uuid */;
  affected_stock?: Stock;
  game_id: string;
  show_at: any /* pgtype.Interval */;
  created_at: string /* RFC3339 */;
}
export interface Portfolio {
  stock?: Stock;
  amount: { [key: number /* int64 */]: PortfolioAmount};
}
export interface PortfolioAmount {
  amount: number /* int64 */;
}

//////////
// source: transact.go

export interface CreateTransaction {
  stock_id: string;
  amount: number /* int64 */;
  action: string;
}
export interface UserTransaction {
  id: string;
  user_id: string;
  game_id: string;
  origin_game_id: string;
  stock_id: string;
  price_index: number /* int */;
  sale_price: number /* int64 */;
  amount: number /* int64 */;
  action: string;
  created_at: string /* RFC3339 */;
}
export interface TransactionList {
  transactions: UserTransaction[];
  users: { [key: string]: User | undefined};
  games: { [key: string]: Game | undefined};
  stocks: { [key: string]: Stock | undefined};
}

//////////
// source: user.go

export interface User {
  id: string;
  username: string;
  enabled: boolean;
  created_at: string /* RFC3339 */;
}
