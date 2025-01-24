package internal

import (
	"reflect"
	"slices"
	"testing"
)

func TestSeeding(t *testing.T) {
	original := make([]int, 15)
	for i := range original {
		original[i] = i
	}

	seedShuffled := slices.Clone(original)

	SeededShuffle(seedShuffled, SeedSingle, 42)
	if !reflect.DeepEqual(seedShuffled, original) {
		t.Fatal("seeds were shuffled with SeedSingle mode")
	}

	swaps := 0
	for rng := range 30 {
		seedShuffled = slices.Clone(original)
		SeededShuffle(seedShuffled, SeedRandom, int64(rng))

		if !containsAll(seedShuffled, original) {
			t.Fatal("the shuffle removed elements")
		}

		if seedShuffled[0] != original[0] {
			swaps += 1
		}
	}
	if swaps == 0 {
		t.Fatal("the shuffle never swapped the elements")
	}

	for rng := range 30 {
		seedShuffled = slices.Clone(original)
		SeededShuffle(seedShuffled, SeedTiered, int64(rng))

		if seedShuffled[0] != original[0] || seedShuffled[1] != original[1] {
			t.Fatal("the first two seeds should stay fixed in their tier")
		}

		tier2 := seedShuffled[2:4]
		originalTier2 := original[2:4]
		if !containsAll(tier2, originalTier2) {
			t.Fatal("elements were shuffled out of their tier")
		}

		tier3 := seedShuffled[4:8]
		originalTier3 := original[4:8]
		if !containsAll(tier3, originalTier3) {
			t.Fatal("elements were shuffled out of their tier")
		}

		tier4 := seedShuffled[8:14]
		originalTier4 := original[8:14]
		if !containsAll(tier4, originalTier4) {
			t.Fatal("elements were shuffled out of their tier")
		}
	}
}

func containsAll[S ~[]E, E comparable](a, b S) bool {
	if len(a) != len(b) {
		return false
	}

	for _, e := range a {
		if !slices.Contains(b, e) {
			return false
		}
	}
	return true
}
