package secrets

import "testing"

func TestShannonEntropy(t *testing.T) {
	low := ShannonEntropy("aaaaaaaaaaaaaaaaaaaa")
	high := ShannonEntropy("a8d2f9b1c7e3x1z7k9p2")

	if low >= high {
		t.Fatalf("expected higher entropy for random input")
	}
}
