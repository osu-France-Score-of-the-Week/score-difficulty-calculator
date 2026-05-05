package models

type UserStatistics struct {
	User User `json:"user"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
