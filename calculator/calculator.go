package calculator

import (
	"math"
	"score-difficulty-calculator/cache"
	"score-difficulty-calculator/models"
)

// NormP controls how strongly the dominant component (aim or speed) is
// favoured over the weaker one when combining them into a single map score.
// p=2 (L2/Euclidean) is already more peak-biased than simple addition;
// raise p to accentuate the effect further (p=3, 4, …).
const NormP = 1.65

// MissPenaltyFactor controls how harshly misses reduce the final score.
// divisor = 1 + MissPenaltyFactor * log10(missCount + 1)
// Lower values = softer penalty. 0 = no penalty at all.
const MissPenaltyFactor = 0.4

// AccuracyPower controls how harshly low accuracy is penalized.
// Higher values widen the gap between e.g. 95% and 99% accuracy.
const AccuracyPower = 0.9

type ScoreResult struct {
	AimScore    float64
	SpeedScore  float64
	MapScore    float64
	AccMult     float64
	MissPenalty float64
	FinalScore  float64
	MissCount   int
}

func ComputeMapScore(beatmapID int, mods []string, attributes models.BeatmapAttributes) float64 {
	aimScore := math.Pow(attributes.AimDifficulty, 2) / 10 * math.Log10(attributes.AimDifficultStrainCount)
	speedScore := math.Pow(attributes.SpeedDifficulty, 2) / 10 * math.Log10(attributes.SpeedDifficultStrainCount/2)
	return 10.0 * math.Pow(math.Pow(aimScore, NormP)+math.Pow(speedScore, NormP), 1.0/NormP)
}

func ComputeDetailed(attributes models.BeatmapAttributes, score models.Score) ScoreResult {
	aimScore := math.Pow(attributes.AimDifficulty, 2) / 10 * math.Log10(attributes.AimDifficultStrainCount)
	speedScore := math.Pow(attributes.SpeedDifficulty, 2) / 10 * math.Log10(attributes.SpeedDifficultStrainCount/2)
	mapScore := 10.0 * math.Pow(math.Pow(aimScore, NormP)+math.Pow(speedScore, NormP), 1.0/NormP)

	missCount := score.Statistics.CountMiss
	accMult := math.Pow(score.Accuracy, AccuracyPower)
	missPenalty := 1 + MissPenaltyFactor*math.Log10(float64(missCount)+1)
	finalScore := mapScore * accMult / missPenalty

	return ScoreResult{
		AimScore:    aimScore,
		SpeedScore:  speedScore,
		MapScore:    mapScore,
		AccMult:     accMult,
		MissPenalty: missPenalty,
		FinalScore:  finalScore,
		MissCount:   missCount,
	}
}

// CalculateAll computes a detailed score for every cache entry that has a
// matching score in scoreByBeatmapID, combining map difficulty with the
// player's accuracy and miss count.
func CalculateAll(entries []cache.CacheEntry, scoreByBeatmapID map[int]models.Score) map[int]ScoreResult {
	results := make(map[int]ScoreResult, len(entries))
	for _, entry := range entries {
		s, ok := scoreByBeatmapID[entry.BeatmapID]
		if !ok {
			continue
		}
		results[entry.BeatmapID] = ComputeDetailed(entry.Attributes, s)
	}
	return results
}
