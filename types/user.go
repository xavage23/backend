package types

import "time"

type User struct {
	ID        string    `db:"id" json:"id" description:"The ID of the user"`
	Username  string    `db:"username" json:"username" description:"The username of the user"`
	Enabled   bool      `db:"enabled" json:"enabled" description:"Whether the user is enabled"`
	CreatedAt time.Time `db:"created_at" json:"created_at" description:"The time the user was created"`
}
