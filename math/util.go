package math

import "math"

// Max returns maximal value and its indices of a slice
func Max(array []float64) (float64, []int) {
	max, imax := -math.MaxFloat64, []int{}
	for idx, value := range array {
		if max < value {
			imax = []int{idx}
			max = value
		} else if value == max {
			imax = append(imax, idx)
		}
	}
	return max, imax
}
