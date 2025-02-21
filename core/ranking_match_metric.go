package core

import (
	"cmp"
	"slices"
)

// A MatchMetricRanking ranks the players who played
// a list of matches by their performance according
// to some metrics
type MatchMetricRanking struct {
	BaseTieableRanking

	// Each player's metrics which are the basis for the ranks.
	// The metrics are updated in the updateRanks call
	Metrics map[Player]*MatchMetrics

	entrySlots []*Slot
	players    []Player
	matches    []*Match

	walkoverScore Score
}

func (r *MatchMetricRanking) updateRanks() {
	metrics := CreateMetrics(r.matches, nil, r.walkoverScore)
	addZeroMetrics(metrics, r.players)

	r.Metrics = metrics

	sortedByWins := sortByMetric(r.players, metrics, func(m *MatchMetrics) int { return m.Wins })

	tieBroken := make([][]Player, 0, len(sortedByWins)+5)
	for _, tie := range sortedByWins {
		broken := r.breakTie(tie)
		tieBroken = append(tieBroken, broken...)
	}

	ranks := make([][]*Slot, 0, len(tieBroken))
	for _, tie := range tieBroken {
		tiedSlots := make([]*Slot, 0, len(tie))
		for _, p := range tie {
			i := slices.Index(r.players, p)
			tiedSlots = append(tiedSlots, r.entrySlots[i])
		}
		ranks = append(ranks, tiedSlots)
	}

	r.ProcessUpdate(ranks)
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
func (r *MatchMetricRanking) breakTie(tie []Player) [][]Player {
	metrics := r.Metrics
	tieSize := len(tie)
	if tieSize == 1 {
		return [][]Player{tie}
	}
	if tieSize == 2 {
		return r.breakTwoWayTie(tie[0], tie[1])
	}

	sortedBySets := sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.SetDifference })

	if len(sortedBySets) > 1 {
		// Break emerged sub-ties
		subTieBroken := make([][]Player, 0, 5)
		for _, subTie := range sortedBySets {
			broken := r.breakTie(subTie)
			subTieBroken = append(subTieBroken, broken...)
		}
		return subTieBroken
	}

	sortedByPoints := sortByMetric(tie, metrics, func(m *MatchMetrics) int { return m.PointDifference })

	// Attempt to break remaining 2-way-ties. Other ties are unbreakable.
	subTieBroken := make([][]Player, 0, 5)
	for _, subTie := range sortedByPoints {
		if len(subTie) == 2 {
			broken := r.breakTwoWayTie(subTie[0], subTie[1])
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
func (r *MatchMetricRanking) breakTwoWayTie(p1, p2 Player) [][]Player {
	metrics := r.Metrics
	matches := r.matches
	tie := []Player{p1, p2}
	directMetrics := CreateMetrics(matches, tie, r.walkoverScore)
	addZeroMetrics(directMetrics, tie)

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
	for _, v := range sortedMetrics {
		sortedPlayers = append(sortedPlayers, buckets[v])
	}

	return sortedPlayers
}

func createMatchMetricRanking(
	entries Ranking,
	matches []*Match,
	walkoverScore Score,
	requiredUntiedRanks int,
	rankingGraph *RankingGraph,
) *MatchMetricRanking {
	entrySlots := entries.Ranks()
	players := make([]Player, 0, len(entrySlots))

	for _, s := range entrySlots {
		player := s.Player
		if player != nil {
			players = append(players, player)
		}
	}

	ranking := &MatchMetricRanking{
		BaseTieableRanking: NewBaseTieableRanking(requiredUntiedRanks),
		entrySlots:         entrySlots,
		players:            players,
		matches:            matches,
		walkoverScore:      walkoverScore,
	}
	ranking.updateRanks()

	if rankingGraph != nil {
		rankingGraph.AddVertex(ranking)
		rankingGraph.AddEdge(entries, ranking)
	}

	return ranking
}

func NewRoundRobinRanking(
	entries Ranking,
	matches []*Match,
	walkoverScore Score,
	rankingGraph *RankingGraph,
) *MatchMetricRanking {
	return createMatchMetricRanking(
		entries,
		matches,
		walkoverScore,
		0,
		rankingGraph,
	)
}

func NewCrossGroupRanking(
	entries Ranking,
	matches []*Match,
	walkoverScore Score,
	numQualifications int,
) *MatchMetricRanking {
	return createMatchMetricRanking(
		entries,
		matches,
		walkoverScore,
		numQualifications,
		nil,
	)
}
