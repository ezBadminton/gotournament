package internal

import (
	"cmp"
	"slices"
)

// A MatchMetricRanking ranks the players who played
// a list of matches by their performance according
// to some metrics
type MatchMetricRanking struct {
	BaseTieableRanking

	players []Player
	matches []*Match
}

func (r *MatchMetricRanking) UpdateRanks() {
	metrics := CreateMetrics(r.matches, nil)

	sortedByWins := sortByMetric(r.players, metrics, func(m *MatchMetrics) int { return m.Wins })

	tieBroken := make([][]Player, 0, len(sortedByWins)+5)
	for _, tie := range sortedByWins {
		broken := breakTie(metrics, r.matches, tie)
		tieBroken = append(tieBroken, broken...)
	}

	ranks := make([][]*Slot, 0, len(tieBroken))
	for _, tie := range tieBroken {
		tiedSlots := make([]*Slot, 0, len(tie))
		for _, p := range tie {
			tiedSlots = append(tiedSlots, NewPlayerSlot(p))
		}
		ranks = append(ranks, tiedSlots)
	}

	r.ProcessUpdate(ranks)
}

func NewMatchMetricRanking(
	entries Ranking,
	matches []*Match,
	rankingGraph *RankingGraph,
) *MatchMetricRanking {
	entrySlots := entries.GetRanks()
	players := make([]Player, 0, len(entrySlots))

	for _, s := range entrySlots {
		player := s.Player()
		if player != nil {
			players = append(players, player)
		}
	}

	ranking := &MatchMetricRanking{
		BaseTieableRanking: NewBaseTieableRanking(0),
		players:            players,
		matches:            matches,
	}
	ranking.UpdateRanks()

	rankingGraph.AddVertex(ranking)
	rankingGraph.AddEdge(entries, ranking)

	return ranking
}

// Attempts to break the tie between players with the same amount of wins.
//
// The tie-break operates in this order:
//   - If the tie has only 2 players it is forwarded to breakTwoWayTie
//   - Who won more sets in all their matches (according to stats)
//   - If that yields smaller ties they are recursively broken by breakTie
//   - Who won more points in all their matches
//   - If that yields 2-way-ties they are broken with breakTwoWayTie
//
// The returned list is descending in rank and each nested list is a rank
// of players. More than one player in a rank means the tie could not be fully
// broken.
func breakTie(metrics map[Player]*MatchMetrics, matches []*Match, tie []Player) [][]Player {
	tieSize := len(tie)
	if tieSize == 1 {
		return [][]Player{tie}
	}
	if tieSize == 2 {
		return breakTwoWayTie(metrics, matches, tie[0], tie[1])
	}

	sortedBySets := sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.SetDifference })

	if len(sortedBySets) > 1 {
		// Break emerged sub-ties
		subTieBroken := make([][]Player, 0, 5)
		for _, subTie := range sortedBySets {
			broken := breakTie(metrics, matches, subTie)
			subTieBroken = append(subTieBroken, broken...)
		}
		return subTieBroken
	}

	sortedByPoints := sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.PointDifference })

	// Attempt to break remaining 2-way-ties. Other ties are unbreakable.
	subTieBroken := make([][]Player, 0, 5)
	for _, subTie := range sortedByPoints {
		if len(subTie) == 2 {
			broken := breakTwoWayTie(metrics, matches, tie[0], tie[1])
			subTieBroken = append(subTieBroken, broken...)
		} else {
			subTieBroken = append(subTieBroken, subTie)
		}
	}

	return subTieBroken
}

// Attempts to break a two-way-tie between p1 and p2.
//
// The tie-break operates in this order:
//
// Who won more...
//   - direct encounters (inside matches)
//   - sets in the direct encounters
//   - points in the direct encounters
//   - sets in all their matches (according to the metrics)
//   - points in all their matches
//
// If none of those criteria are decisive the tie is unbreakable and
// [[p1, p2]] is returned. Otherwise [[winner],[loser]].
func breakTwoWayTie(metrics map[Player]*MatchMetrics, matches []*Match, p1, p2 Player) [][]Player {
	tie := []Player{p1, p2}
	directMetrics := CreateMetrics(matches, tie)

	metricSorted := sortByMetric(tie, directMetrics, func(m *MatchMetrics) int { return m.Wins })
	if len(metricSorted) == 2 {
		return metricSorted
	}

	metricSorted = sortByMetric(tie, directMetrics, func(m *MatchMetrics) int { return m.SetDifference })
	if len(metricSorted) == 2 {
		return metricSorted
	}

	metricSorted = sortByMetric(tie, directMetrics, func(m *MatchMetrics) int { return m.PointDifference })
	if len(metricSorted) == 2 {
		return metricSorted
	}

	metricSorted = sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.SetDifference })
	if len(metricSorted) == 2 {
		return metricSorted
	}

	metricSorted = sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.PointDifference })

	return metricSorted
}

// Sorts the players in descending buckets of one of the metrics returned by the getter
func sortByMetric(players []Player, metrics map[Player]*MatchMetrics, getter func(m *MatchMetrics) int) [][]Player {
	buckets := make(map[int][]Player)

	for _, p := range players {
		metric := getter(metrics[p])
		bucket, ok := buckets[metric]
		if !ok {
			bucket = make([]Player, 0, 3)
		}
		buckets[metric] = append(bucket, p)
	}

	sortedMetrics := make([]int, 0, len(buckets))
	for k := range buckets {
		sortedMetrics = append(sortedMetrics, k)
	}
	slices.SortFunc(sortedMetrics, func(a, b int) int { return cmp.Compare(b, a) })

	sortedPlayers := make([][]Player, 0, len(sortedMetrics))
	for v := range sortedMetrics {
		sortedPlayers = append(sortedPlayers, buckets[v])
	}

	return sortedPlayers
}
