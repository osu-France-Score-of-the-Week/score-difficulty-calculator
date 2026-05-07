package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"score-difficulty-calculator/cache"
	"score-difficulty-calculator/calculator"
	"score-difficulty-calculator/models"
	"score-difficulty-calculator/service"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const scoresCacheFile = "scores_cache.json"

var ignoredStatuses = map[string]bool{
	"graveyard": true,
	"pending":   true,
	"wip":       true,
}

func saveScores(scores map[int]models.Score) error {
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(scoresCacheFile, data, 0644)
}

func loadScores() (map[int]models.Score, error) {
	data, err := os.ReadFile(scoresCacheFile)
	if err != nil {
		return nil, err
	}
	var scores map[int]models.Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, err
	}
	return scores, nil
}

func main() {
	skipFetch := flag.Bool("skip-fetch", false, "Skip fetching pinned scores and map attributes, use cached data")
	flag.Parse()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}

	beatmapCache, err := cache.NewBeatmapCache("beatmap_cache.json")
	if err != nil {
		log.Fatalf("Error initializing beatmap cache: %v", err)
	}

	scoreByBeatmapID := make(map[int]models.Score)

	if *skipFetch {
		scoreByBeatmapID, err = loadScores()
		if err != nil {
			log.Fatalf("Could not load %s (run without --skip-fetch first): %v", scoresCacheFile, err)
		}
		fmt.Printf("Loaded %d score(s) from %s\n", len(scoreByBeatmapID), scoresCacheFile)
	} else {
		osuService := service.NewOsuService(beatmapCache)

		authResponse, err := osuService.Authenticate()
		if err != nil {
			log.Fatalf("Authentication error: %v", err)
		}
		token := authResponse.AccessToken

		ranking, err := osuService.GetRanking(token)
		if err != nil {
			log.Fatalf("GetRanking error: %v", err)
		}
		fmt.Printf("Fetched %d players from osu API\n", len(ranking.Ranking))

		for _, entry := range ranking.Ranking {
			player := entry.User
			pinnedScores, err := osuService.GetUserPinnedScores(player.ID, token)
			if err != nil {
				fmt.Printf("  [WARN] Could not fetch pinned scores for %s: %v\n", player.Username, err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			fmt.Printf("  %s — %d pinned score(s)\n", player.Username, len(pinnedScores))

			for _, s := range pinnedScores {
				if ignoredStatuses[s.Beatmap.Status] {
					continue
				}
				scoreByBeatmapID[s.Beatmap.ID] = s
			}
			time.Sleep(500 * time.Millisecond)
		}

		fmt.Printf("\nTotal unique maps collected: %d\n", len(scoreByBeatmapID))

		if err := saveScores(scoreByBeatmapID); err != nil {
			fmt.Printf("[WARN] Could not save scores to %s: %v\n", scoresCacheFile, err)
		}

		fetched := 0
		for _, s := range scoreByBeatmapID {
			if _, ok := beatmapCache.Get(s.Beatmap.ID, s.Mods); ok {
				continue
			}
			attrs, err := osuService.GetBeatmapAttributes(s.Beatmap.ID, s.Mods, token)
			if err != nil {
				fmt.Printf("  [WARN] Could not fetch attributes for beatmap %d: %v\n", s.Beatmap.ID, err)
				time.Sleep(time.Second)
				continue
			}
			if err := beatmapCache.Set(s.Beatmap.ID, s.Mods, attrs); err != nil {
				fmt.Printf("  [WARN] Could not cache attributes for beatmap %d: %v\n", s.Beatmap.ID, err)
			}
			fetched++
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Printf("Fetched %d new beatmap attribute(s) from API\n", fetched)
	}

	// Calculate and rank
	entries := beatmapCache.GetAll()
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
		fmt.Printf("#%-3d %-20s  %-50s  acc: %5.2f%%  pp: %6.2f  mods: %-10s  miss: %-4d  aim: %6.4f  speed: %6.4f  map: %.4f  acc_mult: %.4f  miss_pen: %.4f  total: %.4f\n",
			i+1,
			s.User.Username,
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
