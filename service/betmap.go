package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"score-difficulty-calculator/models"
)

func (s *OsuService) GetBeatmapAttributes(id int, mods []string, token string) (models.BeatmapAttributes, error) {
	client := &http.Client{}

	body, err := json.Marshal(map[string]interface{}{
		"mods":       mods,
		"ruleset_id": 0,
	})
	if err != nil {
		return models.BeatmapAttributes{}, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://osu.ppy.sh/api/v2/beatmaps/%d/attributes", id),
		bytes.NewReader(body),
	)
	if err != nil {
		return models.BeatmapAttributes{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return models.BeatmapAttributes{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return models.BeatmapAttributes{}, fmt.Errorf("osu API error: %s — %s", resp.Status, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.BeatmapAttributes{}, err
	}

	var response models.BeatmapAttributesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return models.BeatmapAttributes{}, err
	}

	return response.Attributes, nil
}
