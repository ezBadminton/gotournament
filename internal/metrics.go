package internal

import "slices"

type MatchMetrics struct {
	NumMatches, Wins, Losses       int
	NumSets, SetWins, SetLosses    int
	PointWins, PointLosses         int
	SetDifference, PointDifference int
	Withdrawn                      bool
}

func (m *MatchMetrics) UpdateDifferences() {
	m.SetDifference = m.SetWins - m.SetLosses
	m.PointDifference = m.PointWins - m.PointLosses
}

// Creates a MatchMetrics struct for each player in the matches.
// If the players slice is not nil/empty only the matches where both
// opponents are in the slice are counted.
func CreateMetrics(
	matches []*Match,
	players []Player,
	walkoverScore Score,
) map[Player]*MatchMetrics {
	metrics := make(map[Player]*MatchMetrics)

	for _, m := range matches {
		extractMatchMetrics(m, players, metrics, walkoverScore)
	}

	for _, m := range metrics {
		m.UpdateDifferences()
	}

	return metrics
}

func extractMatchMetrics(
	match *Match,
	players []Player,
	metrics map[Player]*MatchMetrics,
	walkoverScore Score,
) {
	p1 := match.Slot1.Player()
	p2 := match.Slot2.Player()
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
	winner := winnerSlot.Player()

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
		if winnerSlot == match.Slot1 {
			score = walkoverScore
			m2.Withdrawn = true
		} else if winnerSlot == match.Slot2 {
			score = walkoverScore.Invert()
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
