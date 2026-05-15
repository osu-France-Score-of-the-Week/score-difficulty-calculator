package models

type AnalyzeRequest struct {
	Beatmap           Beatmap           `json:"beatmap"`
	BeatmapAttributes BeatmapAttributes `json:"beatmap_attributes"`
	Score             Score             `json:"score"`
}
