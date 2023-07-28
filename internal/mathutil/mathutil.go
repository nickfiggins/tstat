package mathutil

import "math"

// Percent returns the percent, rounded to two decimal places.
func Percent(num, den int64) float64 {
	if den == 0 {
		return 0
	}
	pct := float64(num) / float64(den)
	return round(pct * 100)
}

func round(f float64) float64 {
	return math.Round(f*10) / 10
}
