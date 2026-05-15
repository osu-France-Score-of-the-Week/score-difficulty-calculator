package models

import (
	"encoding/json"
	"log"
)

type Score struct {
	ID         int             `json:"id"`
	Accuracy   float64         `json:"accuracy"`
	PP         float64         `json:"pp"`
	Rank       string          `json:"rank"`
	Beatmap    Beatmap         `json:"beatmap"`
	BeatmapSet BeatmapSet      `json:"beatmapset"`
	User       User            `json:"user"`
	Mods       Mods            `json:"mods"`
	Statistics ScoreStatistics `json:"statistics"`
}

type ScoreStatistics struct {
	Count300  int `json:"count_300"`
	Count100  int `json:"count_100"`
	Count50   int `json:"count_50"`
	CountMiss int `json:"count_miss"`
}

type Mods []Mod

type Mod struct {
	Acronym  string                 `json:"acronym"`
	Settings map[string]interface{} `json:"settings"`
}

func (m Mods) Acronyms() []string {
	result := make([]string, 0, len(m))
	for _, mod := range m {
		if mod.Acronym != "" {
			result = append(result, mod.Acronym)
		}
	}
	return result
}

func (m *Mods) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*m = Mods{}
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	mods := make(Mods, 0, len(raw))
	for _, item := range raw {
		var acronym string
		if err := json.Unmarshal(item, &acronym); err == nil {
			mods = append(mods, Mod{Acronym: acronym})
			continue
		}

		var mod Mod
		if err := json.Unmarshal(item, &mod); err != nil {
			return err
		}
		mods = append(mods, mod)
	}

	*m = mods
	return nil
}

func (s *ScoreStatistics) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*s = ScoreStatistics{}
		log.Printf("[calc] statistics empty/null -> %+v", *s)
		return nil
	}

	var raw map[string]int
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("[calc] statistics unmarshal failed raw=%s err=%v", string(data), err)
		return err
	}

	*s = ScoreStatistics{
		Count300:  firstNonZero(raw["count_300"], raw["great"], raw["300"], raw["hit300"]),
		Count100:  firstNonZero(raw["count_100"], raw["ok"], raw["100"], raw["hit100"]),
		Count50:   firstNonZero(raw["count_50"], raw["meh"], raw["50"], raw["hit50"]),
		CountMiss: firstNonZero(raw["count_miss"], raw["miss"]),
	}
	log.Printf("[calc] statistics normalized raw=%v -> %+v", raw, *s)

	return nil
}

func firstNonZero(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}
