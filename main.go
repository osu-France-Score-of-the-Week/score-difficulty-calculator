package main

import (
	"fmt"
	"log"
	"score-difficulty-calculator/cache"
	"score-difficulty-calculator/calculator"
	"score-difficulty-calculator/models"
	"score-difficulty-calculator/service"
	"sort"
	"strings"

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

	// Index scores by beatmap ID for lookup after calculation
	scoreByBeatmapID := make(map[int]models.Score, len(scores))
	for _, s := range scores {
		scoreByBeatmapID[s.Beatmap.ID] = s
	}

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
		fmt.Printf("BeatmapID: %d, Mods: %v, Attributes: %+v\n", entry.BeatmapID, entry.Mods, entry.Attributes)
	}

	computed := calculator.CalculateAll(entries, scoreByBeatmapID)

	type result struct {
		beatmapID int
		calc      calculator.ScoreResult
	}
	results := make([]result, 0, len(computed))
	for id, calc := range computed {
		results = append(results, result{beatmapID: id, calc: calc})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].calc.FinalScore > results[j].calc.FinalScore
	})

	fmt.Println("\n--- Scores by difficulty ---")
	for i, r := range results {
		s, ok := scoreByBeatmapID[r.beatmapID]
		if !ok {
			continue
		}
		mods := "NM"
		if len(s.Mods) > 0 {
			mods = strings.Join(s.Mods, ",")
		}
		fmt.Printf("#%-3d %-50s  acc: %5.2f%%  pp: %6.2f  mods: %-10s  miss: %-4d  aim: %6.4f  speed: %6.4f  map: %.4f  acc_mult: %.4f  miss_pen: %.4f  total: %.4f\n",
			i+1,
			fmt.Sprintf("%s - %s [%s]", s.BeatmapSet.Artist, s.BeatmapSet.Title, s.Beatmap.Version),
			s.Accuracy*100,
			s.PP,
			mods,
			r.calc.MissCount,
			r.calc.AimScore,
			r.calc.SpeedScore,
			r.calc.MapScore,
			r.calc.AccMult,
			r.calc.MissPenalty,
			r.calc.FinalScore,
		)
	}
	fmt.Println("----------------------------")
}
