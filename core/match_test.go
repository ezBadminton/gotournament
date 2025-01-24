package core

import "errors"

type TestScore struct {
	a, b int

	numSets int
}

// Points of first opponent
func (s *TestScore) Points1() []int {
	return s.score(s.a)
}

// Points of second opponent
func (s *TestScore) Points2() []int {
	return s.score(s.b)
}

func (s *TestScore) score(points int) []int {
	score := make([]int, 0, s.numSets)
	for range s.numSets {
		score = append(score, points)
	}
	return score
}

// Returns either 0 or 1 whether the
// first opponent won or the second.
// Errors when no winner is determined.
func (s *TestScore) GetWinner() (int, error) {
	if s.a > s.b {
		return 0, nil
	}
	if s.b > s.a {
		return 1, nil
	}
	return -1, errors.New("No winner")
}

func (s *TestScore) Invert() Score {
	return NewScore(s.b, s.a)
}

func NewScore(a, b int) *TestScore {
	return &TestScore{a, b, 1}
}
