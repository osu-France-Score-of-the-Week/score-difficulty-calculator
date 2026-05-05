package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"score-difficulty-calculator/models"
	"strings"
)

func (s *OsuService) Authenticate() (models.OAuthResponse, error) {
	client := &http.Client{}

	clientID := os.Getenv("OSU_CLIENT_ID")
	clientSecret := os.Getenv("OSU_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return models.OAuthResponse{}, fmt.Errorf("missing OSU_CLIENT_ID or OSU_CLIENT_SECRET environment variable")
	}

	request := models.OAuthRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
		Scope:        "public",
	}

	form := request.ToValues()

	req, err := http.NewRequest(
		"POST",
		"https://osu.ppy.sh/oauth/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return models.OAuthResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return models.OAuthResponse{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.OAuthResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return models.OAuthResponse{}, fmt.Errorf("osu OAuth error: %s", string(body))
	}

	var oauth models.OAuthResponse
	if err := json.Unmarshal(body, &oauth); err != nil {
		return models.OAuthResponse{}, err
	}

	fmt.Println("Successfully obtained osu access token")

	return oauth, nil
}
