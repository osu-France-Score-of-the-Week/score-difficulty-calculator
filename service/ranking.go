package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"score-difficulty-calculator/models"
)

func (s *OsuService) GetRanking(token string) (models.Ranking, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		"https://osu.ppy.sh/api/v2/rankings/osu/performance?country=FR",
		nil,
	)
	if err != nil {
		return models.Ranking{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return models.Ranking{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return models.Ranking{}, fmt.Errorf("osu API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Ranking{}, err
	}

	var ranking models.Ranking
	if err := json.Unmarshal(body, &ranking); err != nil {
		return models.Ranking{}, err
	}

	return ranking, nil
}
