package main

import (
	"fmt"
	"log"
	"score-difficulty-calculator/cache"
	"score-difficulty-calculator/service"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	beatmapCache, err := cache.NewBeatmapCache("beatmap_cache.json")
	if err != nil {
		log.Fatalf("Error initializing beatmap cache: %v", err)
	}

	osuService := service.NewOsuService(beatmapCache)

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

	// for _, score := range scores {
	// 	attributes, err := osuService.GetBeatmapAttributes(score.Beatmap.ID, score.Mods, token)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// 	beatmapCache.Set(score.Beatmap.ID, score.Mods, attributes)
	// 	time.Sleep(time.Second)
	// 	fmt.Printf("Fetched beatmap attributes: %+v\n", attributes)
	// }

	entries := beatmapCache.GetAll()
	fmt.Printf("Loaded %d cached entries\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("BeatmapID: %d, Mods: %v\n", entry.BeatmapID, entry.Mods)
	}
}
