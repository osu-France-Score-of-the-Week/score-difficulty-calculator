package models

type Beatmap struct {
	ID      int     `json:"id"`
	Version string  `json:"version"`
	Status  string  `json:"status"`
	CS      float64 `json:"cs"`
	OD      float64 `json:"accuracy"`
	AR      float64 `json:"ar"`
}

type BeatmapAttributesResponse struct {
	Attributes BeatmapAttributes `json:"attributes"`
}

type BeatmapAttributes struct {
	StarRating                float64 `json:"star_rating"`
	MaxCombo                  int     `json:"max_combo"`
	AimDifficulty             float64 `json:"aim_difficulty"`
	AimDifficultSliderCount   float64 `json:"aim_difficult_slider_count"`
	SpeedDifficulty           float64 `json:"speed_difficulty"`
	SpeedNoteCount            float64 `json:"speed_note_count"`
	SliderFactor              float64 `json:"slider_factor"`
	AimDifficultStrainCount   float64 `json:"aim_difficult_strain_count"`
	SpeedDifficultStrainCount float64 `json:"speed_difficult_strain_count"`
}
