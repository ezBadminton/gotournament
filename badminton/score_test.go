package badminton

import (
	"reflect"
	"testing"
)

func TestScoreSettings(t *testing.T) {
	_, err := NewScoreSettings(0, 2, 30, true)
	if err != ErrPointsZero {
		t.Fatal("zero points did not error")
	}

	_, err = NewScoreSettings(21, 0, 30, true)
	if err != ErrSetsZero {
		t.Fatal("zero sets did not error")
	}

	_, err = NewScoreSettings(21, 2, 20, true)
	if err != ErrMaxPoints {
		t.Fatal("max points less than winning points did not error")
	}

	_, err = NewScoreSettings(21, 2, 21, true)
	if err != nil {
		t.Fatal("max points equal to winning points did error")
	}

	settings, err := NewScoreSettings(21, 2, 0, false)
	if err != nil || settings.MaxPoints != 21 {
		t.Fatal("max points not overridden when two point winning margin is false")
	}

	_, err = NewScoreSettings(21, 2, 30, true)
	if err != nil {
		t.Fatal("standard badminton score setting did error")
	}
}

func TestScoreInterface(t *testing.T) {
	settings, _ := NewScoreSettings(21, 2, 30, true)

	a := []int{21, 21}
	b := []int{1, 2}
	score, err := NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(a, score.Points1()) || !reflect.DeepEqual(b, score.Points2()) {
		t.Fatal("score interface did not reproduce the given points")
	}

	inverted := score.Invert()
	if !reflect.DeepEqual(a, inverted.Points2()) || !reflect.DeepEqual(b, inverted.Points1()) {
		t.Fatal("score interface did not invert the given points")
	}
}

func TestScoreErrors(t *testing.T) {
	settings, _ := NewScoreSettings(21, 2, 30, true)

	a := []int{}
	b := []int{}
	_, err := NewScore(a, b, settings)
	if err != ErrEmpty {
		t.Fatal("empty score did not error")
	}

	a = []int{21, 22}
	b = []int{18}
	_, err = NewScore(a, b, settings)
	if err != ErrUnequalSets {
		t.Fatal("unequal sets did not error")
	}

	a = []int{21}
	b = []int{18}
	_, err = NewScore(a, b, settings)
	if err != ErrTooFewSets {
		t.Fatal("too few sets did not error")
	}

	a = []int{21, 21, 21, 21}
	b = []int{18, 18, 19, 7}
	_, err = NewScore(a, b, settings)
	if err != ErrTooManySets {
		t.Fatal("too many sets did not error")
	}

	a = []int{21, 7, 10}
	b = []int{23, 21, 21}
	_, err = NewScore(a, b, settings)
	if err != ErrUnneededSets {
		t.Fatal("unneeded extra sets did not error")
	}

	a = []int{21, 7, 21}
	b = []int{23, 21, 0}
	_, err = NewScore(a, b, settings)
	if err != ErrUnneededSets {
		t.Fatal("unneeded extra sets did not error")
	}

	a = []int{21, 21}
	b = []int{21, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrUndeterminedSet {
		t.Fatal("undetermined set did not error")
	}

	a = []int{21, 21}
	b = []int{-1, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrNegativePoints {
		t.Fatal("negative points did not error")
	}

	a = []int{20, 21}
	b = []int{17, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrTooFewPoints {
		t.Fatal("too few points did not error")
	}

	a = []int{31, 21}
	b = []int{29, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrTooManyPoints {
		t.Fatal("too many points did not error")
	}

	a = []int{29, 21}
	b = []int{28, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrInvalidMargin {
		t.Fatal("invalid winning margin did not error")
	}

	a = []int{30, 21}
	b = []int{27, 18}
	_, err = NewScore(a, b, settings)
	if err != ErrInvalidMargin {
		t.Fatal("invalid winning margin did not error")
	}

	a = []int{21, 7}
	b = []int{7, 21}
	_, err = NewScore(a, b, settings)
	if err != ErrEqualSetWins {
		t.Fatal("equal set wins did not error")
	}
}

func TestValidScores(t *testing.T) {
	settings, _ := NewScoreSettings(21, 2, 30, true)

	a := []int{21, 21}
	b := []int{10, 19}
	score, err := NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ := score.GetWinner()
	if winner != 0 {
		t.Fatal("score returned the wrong winner")
	}

	a = []int{0, 9}
	b = []int{21, 21}
	score, err = NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ = score.GetWinner()
	if winner != 1 {
		t.Fatal("score returned the wrong winner")
	}

	a = []int{0, 21, 21}
	b = []int{21, 17, 23}
	score, err = NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ = score.GetWinner()
	if winner != 1 {
		t.Fatal("score returned the wrong winner")
	}

	a = []int{0, 21, 0}
	b = []int{21, 0, 21}
	score, err = NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ = score.GetWinner()
	if winner != 1 {
		t.Fatal("score returned the wrong winner")
	}

	settings, _ = NewScoreSettings(21, 2, 21, false)

	a = []int{21, 21}
	b = []int{20, 8}
	score, err = NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ = score.GetWinner()
	if winner != 0 {
		t.Fatal("score returned the wrong winner")
	}

	settings, _ = NewScoreSettings(21, 3, 21, false)

	a = []int{21, 21, 20, 21}
	b = []int{20, 8, 21, 10}
	score, err = NewScore(a, b, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winner, _ = score.GetWinner()
	if winner != 0 {
		t.Fatal("score returned the wrong winner")
	}
}

func TestMaxScore(t *testing.T) {
	settings, _ := NewScoreSettings(21, 2, 30, false)
	score := MaxScore(settings)
	a := []int{21, 21}
	b := []int{0, 0}
	if !reflect.DeepEqual(score.a, a) || !reflect.DeepEqual(score.b, b) {
		t.Fatal("max score is incorrect")
	}

	settings, _ = NewScoreSettings(15, 4, 30, false)
	score = MaxScore(settings)
	a = []int{15, 15, 15, 15}
	b = []int{0, 0, 0, 0}
	if !reflect.DeepEqual(score.a, a) || !reflect.DeepEqual(score.b, b) {
		t.Fatal("max score is incorrect")
	}
}
