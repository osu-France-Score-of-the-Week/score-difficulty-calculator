package models

type Beatmap struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
}

type BeatmapAttributes struct {
	Difficulty                float64 `json:"difficulty"`
	MaxCombo                  int     `json:"max_combo"`
	AimDifficulty             float64 `json:"aim_difficulty"`
	AimDifficultySliderCount  float64 `json:"aim_difficulty_slider_count"`
	SpeedDifficulty           float64 `json:"speed_difficulty"`
	SpeedNoteCount            float64 `json:"speed_note_count"`
	SliderFactor              float64 `json:"slider_factor"`
	AimDifficultStrainCount   float64 `json:"aim_difficult_strain_count"`
	SpeedDifficultStrainCount float64 `json:"speed_difficult_strain_count"`
}
