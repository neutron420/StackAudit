package secrets

import "math"

func ShannonEntropy(value string) float64 {
	freq := map[rune]int{}
	for _, char := range value {
		freq[char]++
	}

	length := float64(len(value))
	if length == 0 {
		return 0
	}

	var entropy float64
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
