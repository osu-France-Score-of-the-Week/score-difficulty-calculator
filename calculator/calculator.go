package calculator

import (
	"math"
	"log"
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
// Lower values = softer penalty.
const MissPenaltyFactor = 0.3

// AccuracyPower controls how harshly low accuracy is penalized.
// Higher values widen the gap between e.g. 95% and 99% accuracy.
// Lower values reduce the reward for high accuracy.
const AccuracyPower = 0.8

// SliderFactor controls how much it increases API's slider score
// Lower values = increase slider influence in score calculation
const SliderFactor = 1

const CSThreshold = 6.0
const CSMaxMultiplier = 1.08

const HighARThreshold = 10.5
const HighARMaxMultiplier = 1.05

const LowARThreshold = 9.0
const LowARMaxMultiplier = 1.5

const ODThreshold = 9.5
const ODMaxMultiplier = 1.02

const SpeedMultiplier = 0.75

const AimMultiplier = 1

// AccuracyOnAimFactor controls how much accuracy affects aim scores (< 1 = less impact).
// Speed maps reward accuracy more than aim maps.
const AccuracyOnAimFactor = 0.6

// MissOnSpeedFactor controls how much misses affect speed scores (< 1 = less impact).
// Aim maps are more sensitive to misses than speed maps.
const MissOnSpeedFactor = 0.6

type ScoreResult struct {
	AimScore    float64
	SpeedScore  float64
	MapScore    float64
	AccMult     float64
	SliderScore float64
	MissPenalty float64
	ARMult      float64
	EffectiveAR float64
	EffectiveOD float64
	FinalScore  float64
	MissCount   int
}

func finiteOrZero(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

// arMult returns a bonus multiplier based on effective AR.
// Low AR (< 9): harder to read (screen clutter) → exponential bonus, max at AR 0.
// High AR (> 10.5): harder to read (objects too fast) → linear bonus, max at AR 11.
func computeARMult(ar float64) float64 {
	mult := 1.0

	// Low AR bonus: exponential, higher reward for very low AR
	if ar < LowARThreshold {
		// Normalized from 0 to 1, where ar=0 gives t=1 and ar=9 gives t=0
		t := (LowARThreshold - ar) / LowARThreshold
		// Exponential curve: more aggressive for very low AR
		mult += math.Pow(t, 1.5) * (LowARMaxMultiplier - 1)
	} else if ar > HighARThreshold {
		// High AR bonus: linear, moderate reward
		t := (ar - HighARThreshold) / (11.0 - HighARThreshold)
		mult += t * (HighARMaxMultiplier - 1)
	}
	return mult
}

// csMult returns a bonus multiplier based on CS.
// CS rewards exponentially from 6.0 to 10.0, representing increased difficulty.
func computeCSMult(cs float64) float64 {
	if cs < CSThreshold {
		return 1.0
	}
	// Exponential bonus from CS 6 to 10
	t := (cs - CSThreshold) / (10.0 - CSThreshold)
	return 1.0 + math.Pow(t, 1.8)*(CSMaxMultiplier-1)
}

// odMult returns a bonus multiplier based on OD and accuracy.
// Low OD = easier to acc → if accuracy is bad, penalty.
// High OD (> 9.5) = harder to acc → slight reward for good accuracy.
// No multiplier for OD 9.5-10, very light multiplier above 10.
func computeODMult(od float64, accuracy float64) float64 {
	if od <= 9.5 {
		return 1.0
	}
	// Very light bonus above 9.5, increases gradually toward 11.11
	t := (od - 9.5) / (11.11 - 9.5)
	// Only reward if accuracy is very good (above 98%)
	if accuracy > 0.98 {
		return 1.0 + t*(ODMaxMultiplier-1)*(accuracy-0.98)/0.02
	}
	return 1.0
}

// effectiveAccPower returns AccuracyPower adjusted for OD.
// Low OD = easier to acc → increase power (penalize bad accuracy more).
// OD at or above ODThreshold → neutral (AccuracyPower unchanged).
func effectiveAccPower(od float64) float64 {
	if od < 9.5 {
		t := (9.5 - od) / 9.5
		return AccuracyPower * (1 + t*0.5)
	}
	return AccuracyPower
}

func Compute(attributes models.BeatmapAttributes, score models.Score, beatmap models.Beatmap) (float64, ScoreResult) {
	log.Printf(
		"[calc] compute start beatmap=%d score=%d acc=%.6f pp=%.2f cs=%.2f od=%.2f ar=%.2f max_combo=%d mods=%v stats=%+v attrs={star=%.3f aim=%.3f speed=%.3f slider=%.3f aim_strain=%.3f speed_strain=%.3f}",
		beatmap.ID,
		score.ID,
		score.Accuracy,
		score.PP,
		beatmap.CS,
		beatmap.OD,
		beatmap.AR,
		attributes.MaxCombo,
		score.Mods,
		score.Statistics,
		attributes.StarRating,
		attributes.AimDifficulty,
		attributes.SpeedDifficulty,
		attributes.SliderFactor,
		attributes.AimDifficultStrainCount,
		attributes.SpeedDifficultStrainCount,
	)

	if math.IsNaN(score.Accuracy) || math.IsInf(score.Accuracy, 0) || score.Accuracy < 0 {
		log.Printf("[calc] invalid accuracy beatmap=%d score=%d acc=%v", beatmap.ID, score.ID, score.Accuracy)
		return 0, ScoreResult{}
	}

	scoreCS := CalculateCSForMod(beatmap, score.Mods)
	scoreOD := CalculateODForMod(beatmap, score.Mods)
	scoreAR := CalculateARForMod(beatmap, score.Mods)
	scoreCS = finiteOrZero(scoreCS)
	scoreOD = finiteOrZero(scoreOD)
	scoreAR = finiteOrZero(scoreAR)

	aimScore := math.Pow(attributes.AimDifficulty, 2) / 10 * math.Sqrt(max(attributes.AimDifficultStrainCount, 500)) * AimMultiplier
	speedScore := math.Pow(attributes.SpeedDifficulty, 2) / 10 * math.Sqrt(max(attributes.SpeedDifficultStrainCount, 250)) * SpeedMultiplier
	mapScore := math.Pow(math.Pow(aimScore, NormP)+math.Pow(speedScore, NormP), 1.0/NormP)
	aimScore = finiteOrZero(aimScore)
	speedScore = finiteOrZero(speedScore)
	mapScore = finiteOrZero(mapScore)
	if mapScore == 0 {
		log.Printf("[calc] zero mapScore beatmap=%d score=%d aim=%.6f speed=%.6f", beatmap.ID, score.ID, aimScore, speedScore)
		return 0, ScoreResult{
			AimScore:    aimScore,
			SpeedScore:  speedScore,
			MapScore:    mapScore,
			ARMult:      computeARMult(scoreAR),
			EffectiveAR: scoreAR,
			EffectiveOD: scoreOD,
			MissCount:   score.Statistics.CountMiss,
		}
	}

	// Apply map stat bonuses
	csMult := computeCSMult(scoreCS)
	arMult := computeARMult(scoreAR)

	mapScore *= csMult * arMult

	// Calculate aim/speed ratio to modulate penalties
	totalDiff := aimScore + speedScore
	if totalDiff <= 0 {
		log.Printf("[calc] non-positive totalDiff beatmap=%d score=%d aim=%.6f speed=%.6f", beatmap.ID, score.ID, aimScore, speedScore)
		return 0, ScoreResult{
			AimScore:    aimScore,
			SpeedScore:  speedScore,
			MapScore:    mapScore,
			ARMult:      arMult,
			EffectiveAR: scoreAR,
			EffectiveOD: scoreOD,
			MissCount:   score.Statistics.CountMiss,
		}
	}
	aimRatio := aimScore / totalDiff
	speedRatio := speedScore / totalDiff

	// OD-adjusted accuracy penalty
	accPower := effectiveAccPower(scoreOD)

	missCount := score.Statistics.CountMiss

	// Modulated accuracy penalty: less impact on aim-heavy maps
	modulatedAccPower := accPower * (1 - aimRatio*(1-AccuracyOnAimFactor))
	accMult := math.Pow(score.Accuracy, modulatedAccPower)

	// OD bonus: slight reward for high OD with good accuracy
	odMult := computeODMult(scoreOD, score.Accuracy)

	// Normalize miss count by map length (combo)
	maxCombo := float64(attributes.MaxCombo)
	if maxCombo <= 0 {
		maxCombo = 1
	}
	missRatio := float64(missCount) / maxCombo

	// Modulated miss penalty: less impact on speed-heavy maps
	// Use exponential penalty (power 1.5) so misses compound heavily
	baseMissImpact := MissPenaltyFactor * math.Pow(missRatio*100, 1.5)
	modulatedMissImpact := baseMissImpact * (1 - speedRatio*(1-MissOnSpeedFactor))
	missPenalty := 1 + modulatedMissImpact

	sliderScore := math.Sqrt(SliderFactor * attributes.SliderFactor)
	sliderScore = finiteOrZero(sliderScore)

	// Calculate final score
	finalScore := mapScore * accMult * odMult / missPenalty / sliderScore
	finalScore = finiteOrZero(finalScore)
	log.Printf(
		"[calc] result beatmap=%d score=%d final=%.6f aim=%.6f speed=%.6f map=%.6f cs_mult=%.6f ar_mult=%.6f acc_mult=%.6f od_mult=%.6f miss_pen=%.6f slider=%.6f eff_ar=%.6f eff_od=%.6f misses=%d",
		beatmap.ID,
		score.ID,
		finalScore,
		aimScore,
		speedScore,
		mapScore,
		csMult,
		arMult,
		accMult,
		odMult,
		missPenalty,
		sliderScore,
		scoreAR,
		scoreOD,
		missCount,
	)

	return finalScore, ScoreResult{
		AimScore:    aimScore,
		SpeedScore:  speedScore,
		MapScore:    mapScore,
		AccMult:     accMult,
		SliderScore: sliderScore,
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
		_, results[entry.BeatmapID] = Compute(entry.Attributes, s, s.Beatmap)
	}
	return results
}

// modSet converts a mods slice to a lookup map for O(1) presence checks.
func modSet(mods models.Mods) map[string]bool {
	s := make(map[string]bool, len(mods))
	for _, m := range mods {
		if m.Acronym != "" {
			s[m.Acronym] = true
		}
	}
	return s
}

func CalculateCSForMod(beatmap models.Beatmap, mods models.Mods) float64 {
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

func CalculateODForMod(beatmap models.Beatmap, mods models.Mods) float64 {
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

func CalculateARForMod(beatmap models.Beatmap, mods models.Mods) float64 {
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
