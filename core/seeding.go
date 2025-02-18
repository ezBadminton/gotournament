package core

import (
	"math/rand"
	"slices"
)

const (
	SeedRandom = iota
	SeedSingle = iota
	SeedTiered = iota
)

func SeededShuffle[S ~[]E, E any](seeded S, unseeded S, seedingMode int, rngSeed int64) S {
	rng := rand.New(rand.NewSource(rngSeed))
	shuffle(unseeded, rng)

	if seedingMode == SeedSingle {
		return slices.Concat(seeded, unseeded)
	}

	// New rng so the seeded shuffle is not influenced by the length of unseeded
	rng = rand.New(rand.NewSource(^rngSeed))
	switch seedingMode {
	case SeedRandom:
		shuffle(seeded, rng)
	case SeedTiered:
		tieredShuffle(seeded, rng)
	}

	return slices.Concat(seeded, unseeded)
}

func tieredShuffle[S ~[]E, E any](slice S, rng *rand.Rand) {
	for start := 2; start < len(slice)-1; start *= 2 {
		end := min(len(slice)-1, 2*start)
		shuffle(slice[start:end], rng)
	}
}

func shuffle[S ~[]E, E any](slice S, rng *rand.Rand) {
	rng.Shuffle(
		len(slice),
		func(i, j int) { slice[i], slice[j] = slice[j], slice[i] },
	)
}
