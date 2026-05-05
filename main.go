package main

import (
	"fmt"
	"log"
	"score-difficulty-calculator/service"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	err2 := godotenv.Load()
	if err2 != nil {
		log.Fatalf("Error loading .env file")
	}

	osuService := service.NewOsuService()

	authResponse, err := osuService.Authenticate()
	if err != nil {
		fmt.Println(err)
		return
	}

	token := authResponse.AccessToken
	fmt.Println(token)

	ranking, err := osuService.GetRanking(token)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Fetched %d players from osu API\n", len(ranking.Ranking))
	fmt.Printf("First player: %s\n", ranking.Ranking[0].User.Username)
	scores, err := osuService.GetUserTopScores(ranking.Ranking[0].User.ID, token)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Fetched %d scores from osu API\n", len(scores))
}
