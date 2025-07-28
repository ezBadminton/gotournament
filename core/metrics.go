package core

import "slices"

type MatchMetrics struct {
	NumMatches int `json:"numMatches"`
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`

	NumSets   int `json:"numSets"`
	SetWins   int `json:"setWins"`
	SetLosses int `json:"setLosses"`

	PointWins   int `json:"pointWins"`
	PointLosses int `json:"pointLosses"`

	SetDifference   int `json:"-"`
	PointDifference int `json:"-"`

	Withdrawn bool `json:"-"`
}

func (m *MatchMetrics) UpdateDifferences() {
	m.SetDifference = m.SetWins - m.SetLosses
	m.PointDifference = m.PointWins - m.PointLosses
}

// Add the other match metrics to this one
func (m *MatchMetrics) Add(other *MatchMetrics) {
	m.NumMatches += other.NumMatches
	m.Wins += other.Wins
	m.Losses += other.Losses

	m.NumSets += other.NumSets
	m.SetWins += other.SetWins
	m.SetLosses += other.SetLosses

	m.PointWins += other.PointWins
	m.PointLosses += other.PointLosses

	m.UpdateDifferences()
}

type baseMatchMetricSource struct {
	matches       []*Match
	walkoverScore Score
}

// Creates a MatchMetrics struct for each player in the matches.
// If the players slice is not nil/empty only the matches where both
// opponents are in the slice are counted.
func (s *baseMatchMetricSource) CreateMetrics(
	players []Player,
) map[Player]*MatchMetrics {
	metrics := make(map[Player]*MatchMetrics)
	s.extractMatchMetricsFromSlice(s.matches, players, metrics)
	return metrics
}

func (s *baseMatchMetricSource) extractMatchMetricsFromSlice(
	matches []*Match,
	players []Player,
	metrics map[Player]*MatchMetrics,
) {
	for _, match := range matches {
		s.extractMatchMetrics(match, players, metrics)
	}

	for _, m := range metrics {
		m.UpdateDifferences()
	}
}

func (s *baseMatchMetricSource) extractMatchMetrics(
	match *Match,
	players []Player,
	metrics map[Player]*MatchMetrics,
) {
	p1 := match.Slot1.Player
	p2 := match.Slot2.Player
	if p1 == nil || p2 == nil {
		return
	}

	doCount1 := len(players) == 0 || slices.Contains(players, p1)
	doCount2 := len(players) == 0 || slices.Contains(players, p2)
	if !doCount1 || !doCount2 {
		return
	}

	winnerSlot, _ := match.GetWinner()
	if winnerSlot == nil {
		return
	}
	winner := winnerSlot.Player

	m1, ok := metrics[p1]
	if !ok {
		m1 = &MatchMetrics{}
		metrics[p1] = m1
	}

	m2, ok := metrics[p2]
	if !ok {
		m2 = &MatchMetrics{}
		metrics[p2] = m2
	}

	m1.NumMatches += 1
	m2.NumMatches += 1

	if winner == p1 {
		m1.Wins += 1
		m2.Losses += 1
	} else {
		m2.Wins += 1
		m1.Losses += 1
	}

	score := match.Score
	if score == nil {
		switch winnerSlot {
		case match.Slot1:
			score = s.walkoverScore
			m2.Withdrawn = true
		case match.Slot2:
			score = s.walkoverScore.Invert()
			m1.Withdrawn = true
		}
	}

	score1 := score.Points1()
	score2 := score.Points2()
	for i := range len(score1) {
		m1.NumSets += 1
		m2.NumSets += 1

		points1 := score1[i]
		points2 := score2[i]

		m1.PointWins += points1
		m1.PointLosses += points2
		m2.PointWins += points2
		m2.PointLosses += points1

		if points1 == points2 {
			continue
		}
		if points1 > points2 {
			m1.SetWins += 1
			m2.SetLosses += 1
		} else {
			m2.SetWins += 1
			m1.SetLosses += 1
		}
	}
}

// Adds zeroed metrics to the metrics map for players which are
// not already present in the map but are in the players slice
func addZeroMetrics(metrics map[Player]*MatchMetrics, players []Player) {
	for _, p := range players {
		_, ok := metrics[p]
		if !ok {
			metrics[p] = &MatchMetrics{}
		}
	}
}
