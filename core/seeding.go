package core

import "math/rand"

const (
	SeedRandom = iota
	SeedSingle = iota
	SeedTiered = iota
)

func SeededShuffle[S ~[]E, E any](slice S, seedingMode int, rngSeed int64) {
	if seedingMode == SeedSingle {
		return
	}

	rng := rand.New(rand.NewSource(rngSeed))
	switch seedingMode {
	case SeedRandom:
		shuffle(slice, rng)
	case SeedTiered:
		tieredShuffle(slice, rng)
	}
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
