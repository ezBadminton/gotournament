package badminton

import (
	"errors"

	"github.com/ezBadminton/gotournament/internal"
)

var (
	ErrPointsZero = errors.New("winning points are zero or less")
	ErrSetsZero   = errors.New("winning sets are zero or less")
	ErrMaxPoints  = errors.New("max points are less than winning points")

	ErrUndetermined = errors.New("the winner is undeterminable from the score")

	ErrEmpty           = errors.New("empty score")
	ErrUndeterminedSet = errors.New("a set has equal points")
	ErrUnequalSets     = errors.New("opponents have unequal number of sets")
	ErrTooManySets     = errors.New("too many sets")
	ErrTooFewSets      = errors.New("too few sets")
	ErrNegativePoints  = errors.New("negative points")
	ErrTooManyPoints   = errors.New("points exceed the max points setting")
	ErrTooFewPoints    = errors.New("set winner points are less than the winning point setting")
	ErrInvalidMargin   = errors.New("the winning point margin is invalid")
	ErrUnneededSets    = errors.New("score contains unneeded extra sets")
	ErrEqualSetWins    = errors.New("both opponents won an equal number of sets")
)

type scoreSettings struct {
	WinningPoints, WinningSets, MaxPoints int
	TwoPointMargin                        bool
}

func NewScoreSettings(
	winningPoints, winningSets, maxPoints int,
	twoPointMargin bool,
) (scoreSettings, error) {
	if !twoPointMargin {
		maxPoints = winningPoints
	}

	scoreSettings := scoreSettings{
		winningPoints, winningSets, maxPoints, twoPointMargin,
	}

	if winningPoints <= 0 {
		return scoreSettings, ErrPointsZero
	}
	if winningSets <= 0 {
		return scoreSettings, ErrSetsZero
	}
	if maxPoints < winningPoints {
		return scoreSettings, ErrMaxPoints
	}

	return scoreSettings, nil
}

type score struct {
	a, b []int
}

func (s *score) Points1() []int {
	return s.a
}

func (s *score) Points2() []int {
	return s.b
}

func (s *score) GetWinner() (int, error) {
	setWins := 0
	for i := range len(s.a) {
		if s.a[i] > s.b[i] {
			setWins += 1
		}
		if s.b[i] > s.a[i] {
			setWins -= 1
		}
	}

	if setWins > 0 {
		return 0, nil
	}
	if setWins < 0 {
		return 1, nil
	}

	return -1, ErrUndetermined
}

func (s *score) Invert() internal.Score {
	score := &score{
		a: s.b,
		b: s.a,
	}
	return score
}

func NewScore(
	a, b []int,
	settings scoreSettings,
) (*score, error) {
	switch {
	case len(a) == 0 || len(b) == 0:
		return nil, ErrEmpty
	case len(a) != len(b):
		return nil, ErrUnequalSets
	case len(a) < settings.WinningSets:
		return nil, ErrTooFewSets
	case len(a) >= 2*settings.WinningSets:
		return nil, ErrTooManySets
	}

	winningMargin := 1
	if settings.TwoPointMargin {
		winningMargin = 2
	}

	setWinsA, setWinsB := 0, 0
	for i := range len(a) {
		w := max(a[i], b[i])
		l := min(a[i], b[i])

		switch {
		case setWinsA == settings.WinningSets || setWinsB == settings.WinningSets:
			return nil, ErrUnneededSets
		case w == l:
			return nil, ErrUndeterminedSet
		case l < 0:
			return nil, ErrNegativePoints
		case w < settings.WinningPoints:
			return nil, ErrTooFewPoints
		case w > settings.MaxPoints:
			return nil, ErrTooManyPoints
		case w < settings.MaxPoints && w > settings.WinningPoints && w-l != winningMargin:
			fallthrough
		case w == settings.MaxPoints && w > settings.WinningPoints && w-l > winningMargin:
			return nil, ErrInvalidMargin
		}

		if a[i] > b[i] {
			setWinsA += 1
		} else {
			setWinsB += 1
		}
	}

	if setWinsA == setWinsB {
		return nil, ErrEqualSetWins
	}

	return &score{a, b}, nil
}

func MaxScore(settings scoreSettings) *score {
	a := make([]int, settings.WinningSets)
	b := make([]int, settings.WinningSets)
	for i := range settings.WinningSets {
		a[i] = settings.WinningPoints
	}
	return &score{a, b}
}
