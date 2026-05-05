package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"score-difficulty-calculator/models"
	"strconv"
)

func (s *OsuService) GetUserTopScores(id int, token string) ([]models.Score, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", "https://osu.ppy.sh/api/v2/users/"+strconv.Itoa(id)+"/scores/best", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("limit", "100")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osu API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var scores []models.Score
	if err := json.Unmarshal(body, &scores); err != nil {
		return nil, err
	}

	return scores, nil
}
