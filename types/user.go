package types

type User struct {
	ID       string `db:"id" json:"id" description:"The ID of the user"`
	Username string `db:"username" json:"username" description:"The username of the user"`
	Enabled  bool   `db:"enabled" json:"enabled" description:"Whether the user is enabled"`
	Root     bool   `db:"root" json:"root" description:"Whether the user is a root user"`
}
