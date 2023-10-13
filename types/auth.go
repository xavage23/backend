package types

type UserLogin struct {
	Username string `json:"username" description:"The username of the user"`
	Password string `json:"password" description:"The password of the user"`
}

type UserLoginResponse struct {
	UserID string `json:"user_id" description:"The id of the user"`
	Token  string `json:"token" description:"The token of the user"`
}
