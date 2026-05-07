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
const NormP = 1.2

// MissPenaltyFactor controls how harshly misses reduce the final score.
// divisor = 1 + MissPenaltyFactor * sqrt(missCount)
// sqrt grows faster than log10 for large miss counts (more linear feel)
// while still being gentle for 1-2 misses. Lower values = softer penalty.
const MissPenaltyFactor = 0.1

// AccuracyPower controls how harshly low accuracy is penalized.
// Higher values widen the gap between e.g. 95% and 99% accuracy.
const AccuracyPower = 0.9

// SliderFactor controls how much it increases API's slider score
// Lower values = increase slider influence in score calculation
const SliderFactor = 1

const CSThreshold = 6.5
const CSMaxMultiplier = 1.3

const HighARThreshold = 10.5
const HighARMultiplier = 1

const LowARThreshold = 8
const LowARMultiplier = 2

const ODThreshold = 9
const ODMaxMultiplier = 1.3

const SpeedMultiplier = 0.75

const AimMultiplier = 1

type ScoreResult struct {
	AimScore    float64
	SpeedScore  float64
	MapScore    float64
	AccMult     float64
	MissPenalty float64
	ARMult      float64
	EffectiveAR float64
	EffectiveOD float64
	FinalScore  float64
	MissCount   int
}

// arMult returns a bonus multiplier based on effective AR.
// High AR (> HighARThreshold): harder to read → bonus up to HighARMultiplier at AR 11.
// Low AR  (< LowARThreshold):  cluttered screen → bonus up to LowARMultiplier at AR 0.
func computeARMult(ar float64) float64 {
	mult := 1.0
	if ar > HighARThreshold {
		t := (ar - HighARThreshold) / (11.0 - HighARThreshold)
		mult += t * (HighARMultiplier - 1)
	} else if ar < LowARThreshold {
		t := (LowARThreshold - ar) / LowARThreshold
		mult += t * (LowARMultiplier - 1)
	}
	return mult
}

// effectiveAccPower returns AccuracyPower adjusted for OD.
// Low OD = easier to acc → increase power (penalize bad accuracy more).
// OD at or above ODThreshold → neutral (AccuracyPower unchanged).
func effectiveAccPower(od float64) float64 {
	if od < ODThreshold {
		t := (ODThreshold - od) / ODThreshold
		return AccuracyPower * (1 + t*(ODMaxMultiplier-1))
	}
	return AccuracyPower
}

func ComputeMapScore(beatmapID int, mods []string, attributes models.BeatmapAttributes) float64 {
	aimScore := math.Pow(attributes.AimDifficulty, 2) / 10 * math.Log10(attributes.AimDifficultStrainCount) * AimMultiplier
	speedScore := math.Pow(attributes.SpeedDifficulty, 2) / 10 * math.Log10(attributes.SpeedDifficultStrainCount) * SpeedMultiplier
	return 10.0 * math.Pow(math.Pow(aimScore, NormP)+math.Pow(speedScore, NormP), 1.0/NormP)
}

func ComputeDetailed(attributes models.BeatmapAttributes, score models.Score) ScoreResult {
	scoreCS := CalculateCSForMod(score.Beatmap, score.Mods)
	scoreOD := CalculateODForMod(score.Beatmap, score.Mods)
	scoreAR := CalculateARForMod(score.Beatmap, score.Mods)

	aimScore := math.Pow(attributes.AimDifficulty, 2) / 10 * math.Log10(attributes.AimDifficultStrainCount) * AimMultiplier
	speedScore := math.Pow(attributes.SpeedDifficulty, 2) / 10 * math.Log10(attributes.SpeedDifficultStrainCount) * SpeedMultiplier
	mapScore := 10.0 * math.Pow(math.Pow(aimScore, NormP)+math.Pow(speedScore, NormP), 1.0/NormP)

	// CS bonus
	if scoreCS >= CSThreshold {
		mapScore *= 1 + (CSMaxMultiplier-1)*math.Pow((scoreCS-CSThreshold)/(10-CSThreshold), 1.2)
	}

	// AR bonus (high AR or low AR both reward the score)
	arMult := computeARMult(scoreAR)

	// OD-adjusted accuracy penalty
	accPower := effectiveAccPower(scoreOD)

	missCount := score.Statistics.CountMiss
	accMult := math.Pow(score.Accuracy, accPower)
	missPenalty := 1 + MissPenaltyFactor*math.Sqrt(float64(missCount))
	sliderScore := SliderFactor * attributes.SliderFactor
	finalScore := mapScore * arMult * accMult / missPenalty * (1 / sliderScore)

	return ScoreResult{
		AimScore:    aimScore,
		SpeedScore:  speedScore,
		MapScore:    mapScore,
		AccMult:     accMult,
		MissPenalty: missPenalty,
		ARMult:      arMult,
		EffectiveAR: scoreAR,
		EffectiveOD: scoreOD,
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

// modSet converts a mods slice to a lookup map for O(1) presence checks.
func modSet(mods []string) map[string]bool {
	s := make(map[string]bool, len(mods))
	for _, m := range mods {
		s[m] = true
	}
	return s
}

func CalculateCSForMod(beatmap models.Beatmap, mods []string) float64 {
	cs := beatmap.CS
	m := modSet(mods)
	// EZ and HR are mutually exclusive in osu!, but we apply in safe order anyway
	if m["EZ"] {
		cs *= 0.5
	}
	if m["HR"] {
		cs = math.Min(cs*1.3, 10.0)
	}
	return cs
}

func CalculateODForMod(beatmap models.Beatmap, mods []string) float64 {
	od := beatmap.OD
	m := modSet(mods)
	// 1. Apply base-value modifiers first
	if m["EZ"] {
		od *= 0.5
	}
	if m["HR"] {
		od = math.Min(od*1.4, 10.0)
	}
	// 2. Apply speed modifier via time-domain conversion
	if m["DT"] || m["NC"] {
		od = CalculateMultipliedOD(od, DTMultiplier)
	} else if m["HT"] {
		od = CalculateMultipliedOD(od, HTMultiplier)
	}
	return od
}

func CalculateARForMod(beatmap models.Beatmap, mods []string) float64 {
	ar := beatmap.AR
	m := modSet(mods)
	// 1. Apply base-value modifiers first
	if m["EZ"] {
		ar *= 0.5
	}
	if m["HR"] {
		ar = math.Min(ar*1.4, 10.0)
	}
	// 2. Apply speed modifier via time-domain conversion
	if m["DT"] || m["NC"] {
		ar = CalculateMultipliedAR(ar, DTMultiplier)
	} else if m["HT"] {
		ar = CalculateMultipliedAR(ar, HTMultiplier)
	}
	return ar
}
