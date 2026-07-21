package search

import (
	"math"
)

func Similarity(a, b string) float64 {

	a = Normalize(a)
	b = Normalize(b)

	if a == "" || b == "" {
		return 0
	}

	distance := levenshtein(a, b)

	maxLen := math.Max(
		float64(len(a)),
		float64(len(b)),
	)

	score := 1 - float64(distance)/maxLen

	if score < 0 {
		return 0
	}

	return score
}

func levenshtein(a, b string) int {

	rows := len(b) + 1
	cols := len(a) + 1

	matrix := make([][]int, rows)

	for i := range matrix {
		matrix[i] = make([]int, cols)
	}

	for i := 0; i < rows; i++ {
		matrix[i][0] = i
	}

	for j := 0; j < cols; j++ {
		matrix[0][j] = j
	}

	for i := 1; i < rows; i++ {

		for j := 1; j < cols; j++ {

			cost := 0

			if a[j-1] != b[i-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(b)][len(a)]
}

func min(a, b, c int) int {

	if a < b && a < c {
		return a
	}

	if b < c {
		return b
	}

	return c
}
