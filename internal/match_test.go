package internal

import "errors"

type TestScore struct {
	a, b int
}

// Points of first opponent
func (s *TestScore) Points1() []int {
	return []int{s.a}
}

// Points of second opponent
func (s *TestScore) Points2() []int {
	return []int{s.b}
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

func NewScore(a, b int) *TestScore {
	return &TestScore{a, b}
}
