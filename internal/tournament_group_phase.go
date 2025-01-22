package internal

import (
	"iter"
)

type GroupPhaseSettings struct {
	NumGroups, NumQualifications int
	WalkoverScore                Score
	RankingGraph                 *RankingGraph
}

type GroupPhaseMatchMaker struct {
	Groups []*RoundRobin
}

func (m *GroupPhaseMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	groupPhaseSettings := settings.(*GroupPhaseSettings)
	numGroups := groupPhaseSettings.NumGroups
	numQualifications := groupPhaseSettings.NumQualifications
	qualsPerGroup := numQualifications / numGroups
	if numQualifications%numGroups != 0 {
		qualsPerGroup += 1
	}
	walkoverScore := groupPhaseSettings.WalkoverScore

	rankingGraph := groupPhaseSettings.RankingGraph

	entrySlots := entries.GetRanks()

	slotGroups := groupSlots(entrySlots, numGroups)
	m.Groups = make([]*RoundRobin, 0, len(slotGroups))

	for _, slots := range slotGroups {
		groupEntries := NewSlotRanking(slots)
		roundRobin := NewGroupRoundRobin(groupEntries, qualsPerGroup, walkoverScore, rankingGraph)
		rankingGraph.AddEdge(entries, groupEntries)
		m.Groups = append(m.Groups, roundRobin)
	}

	matchList := m.createMatchList()

	crossGroupRanking := NewCrossGroupRanking(
		entries,
		matchList.Matches,
		groupPhaseSettings.WalkoverScore,
		numQualifications,
	)

	finalRanking := NewGroupPhaseRanking(
		m.Groups,
		numQualifications,
		crossGroupRanking,
		rankingGraph,
	)

	for _, g := range m.Groups {
		rankingGraph.AddEdge(g.FinalRanking, crossGroupRanking)
	}

	return matchList, rankingGraph, finalRanking, nil
}

func (m *GroupPhaseMatchMaker) createMatchList() *MatchList {
	lastGroup := m.Groups[len(m.Groups)-1]
	maxNumRounds := len(lastGroup.MatchList.Rounds)
	maxNumMatches := len(lastGroup.MatchList.Matches)

	rounds := make([]*Round, 0, maxNumRounds)
	matches := make([]*Match, 0, len(m.Groups)*maxNumMatches)
	for i := range maxNumRounds {
		groupRounds := collectRounds(i, m.Groups)
		roundMatches := intertwineRounds(groupRounds)
		matches = append(matches, roundMatches...)
		round := &Round{Matches: roundMatches, NestedRounds: groupRounds}
		rounds = append(rounds, round)
	}

	matchList := &MatchList{Matches: matches, Rounds: rounds}

	return matchList
}

func collectRounds(roundI int, groups []*RoundRobin) []*Round {
	rounds := make([]*Round, 0, len(groups))
	for _, g := range groups {
		if roundI > len(g.MatchList.Rounds)-1 {
			continue
		}
		rounds = append(rounds, g.MatchList.Rounds[roundI])
	}
	return rounds
}

func intertwineRounds(rounds []*Round) []*Match {
	lastRound := rounds[len(rounds)-1]
	maxMatches := len(lastRound.Matches)
	matches := make([]*Match, 0, len(rounds)*maxMatches)
	for i := range maxMatches {
		for _, r := range rounds {
			if i > len(r.Matches)-1 {
				continue
			}
			matches = append(matches, r.Matches[i])
		}
	}
	return matches
}

// Groups the given slots into numGroups groups.
// The slots are distributed among the groups in a "snaking"
// order going back and forth for seeding purposes.
func groupSlots(slots []*Slot, numGroups int) [][]*Slot {
	groups := make([][]*Slot, 0, numGroups)
	maxGroupSize := len(slots) / numGroups
	if len(slots)%numGroups != 0 {
		maxGroupSize += 1
	}
	for range numGroups {
		groups = append(groups, make([]*Slot, 0, maxGroupSize))
	}

	for len(slots) > 0 {
		snakeDirection := len(groups[0])%2 == 0
		sliceSize := min(len(slots), numGroups)
		currentSlots := slots[:sliceSize]
		slots = slots[sliceSize:]

		iter := directionalSeq(currentSlots, snakeDirection)
		for i, slot := range iter {
			// The higher index groups get the remaining slots
			// if not divisible by numGroups
			i += (numGroups - sliceSize)
			groups[i] = append(groups[i], slot)
		}
	}

	return groups
}

// Returns an index-value-sequence that iterates the given slice normally
// when the direction bool is true, otherwise iterates in
// reverse order. The index is ascending in both cases.
func directionalSeq[V any](slice []V, direction bool) iter.Seq2[int, V] {
	l := len(slice)
	iterator := func(yield func(int, V) bool) {
		for i := range l {
			v := i
			if !direction {
				v = l - i - 1
			}
			if !yield(i, slice[v]) {
				return
			}
		}
	}

	return iterator
}

type GroupPhase struct {
	BaseTournament
}

func NewGroupPhase(
	entries Ranking,
	numGroups, numQualifications int,
	walkoverScore Score,
	rankingGraph *RankingGraph,
) *GroupPhase {
	settings := &GroupPhaseSettings{
		NumGroups:         numGroups,
		NumQualifications: numQualifications,
		WalkoverScore:     walkoverScore,
		RankingGraph:      rankingGraph,
	}

	matchMaker := &GroupPhaseMatchMaker{}
	matchList, rankingGraph, finalRanking, _ := matchMaker.MakeMatches(entries, settings)

	editingPolicy := &RoundRobinEditingPolicy{
		matches: matchList.Matches,
	}

	withdrawalPolicy := &RoundRobinWithdrawalPolicy{
		matchList: matchList,
	}

	tournament := NewBaseTournament(
		entries,
		finalRanking,
		matchMaker,
		matchList,
		rankingGraph,
		editingPolicy,
		withdrawalPolicy,
	)

	groupPhase := &GroupPhase{BaseTournament: tournament}
	groupPhase.Update(nil)

	return groupPhase
}
