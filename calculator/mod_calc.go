package calculator

import "math"

const DTMultiplier = 1.5

const HTMultiplier = 0.75

// approachRateToMs converts an AR value to its hit window in milliseconds.
func approachRateToMs(ar float64) float64 {
	if ar <= 5 {
		return 1800 - ar*120
	}
	return 1200 - (ar-5)*150
}

// msToApproachRate converts a hit window in milliseconds back to an AR value.
// Iterates in steps of 0.1 (AR 0.0 → 11.0) and returns the closest match.
func msToApproachRate(ms float64) float64 {
	smallestDiff := 100000.0
	for ar := 0; ar <= 110; ar++ {
		newDiff := math.Abs(approachRateToMs(float64(ar)/10) - ms)
		if newDiff < smallestDiff {
			smallestDiff = newDiff
		} else {
			return float64(ar-1) / 10
		}
	}
	return 11.0
}

// CalculateMultipliedAR returns the effective AR after applying a speed multiplier
// (e.g. 1.5 for DT/NC).
func CalculateMultipliedAR(ar, multiplier float64) float64 {
	return msToApproachRate(approachRateToMs(ar) / multiplier)
}

// overallDifficultyToMs converts an OD value to its hit window in milliseconds.
func overallDifficultyToMs(od float64) float64 {
	return -6*od + 79.5
}

// msToOverallDifficulty converts a hit window in milliseconds back to an OD value.
func msToOverallDifficulty(ms float64) float64 {
	return (79.5 - ms) / 6
}

// CalculateMultipliedOD returns the effective OD after applying a speed multiplier
// (e.g. 1.5 for DT/NC), rounded to 1 decimal place.
func CalculateMultipliedOD(od, multiplier float64) float64 {
	newMs := overallDifficultyToMs(od) / multiplier
	newOD := msToOverallDifficulty(newMs)
	return math.Round(newOD*10) / 10
}
