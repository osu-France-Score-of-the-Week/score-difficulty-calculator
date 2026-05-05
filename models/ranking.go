package models

type Ranking struct {
	Ranking []UserStatistics `json:"ranking"`
}

type UserStatistics struct {
	PP         float64 `json:"pp"`
	GlobalRank int     `json:"global_rank"`
	User       User    `json:"user"`
}
