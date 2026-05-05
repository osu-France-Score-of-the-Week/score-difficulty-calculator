package models

type Ranking struct {
	Ranking RankingItems `json:"ranking"`
}

type RankingItems []User
