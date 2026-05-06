package models

type Score struct {
	ID         int             `json:"id"`
	Accuracy   float64         `json:"accuracy"`
	PP         float64         `json:"pp"`
	Rank       string          `json:"rank"`
	Beatmap    Beatmap         `json:"beatmap"`
	BeatmapSet BeatmapSet      `json:"beatmapset"`
	User       User            `json:"user"`
	Mods       []string        `json:"mods"`
	Statistics ScoreStatistics `json:"statistics"`
}

type ScoreStatistics struct {
	Count300  int `json:"count_300"`
	Count100  int `json:"count_100"`
	Count50   int `json:"count_50"`
	CountMiss int `json:"count_miss"`
}
