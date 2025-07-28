package core

import (
	"iter"
)

type GroupPhase struct {
	BaseTournament[*GroupPhaseRanking]
	Groups []*RoundRobin
}

func (t *GroupPhase) initTournament(
	entries Ranking,
	numGroups, numQualifications int,
	walkoverScore Score,
	rankingGraph *RankingGraph,
) {
	qualsPerGroup := numQualifications / numGroups
	if numQualifications%numGroups != 0 {
		qualsPerGroup += 1
	}

	entrySlots := entries.Ranks()

	slotGroups := groupSlots(entrySlots, numGroups)
	t.Groups = make([]*RoundRobin, 0, len(slotGroups))

	for _, slots := range slotGroups {
		groupEntries := NewSlotRanking(slots)
		roundRobin, err := newGroupRoundRobin(groupEntries, qualsPerGroup, walkoverScore, rankingGraph)
		if err != nil {
			panic("could not get new group round robin")
		}
		rankingGraph.AddEdge(entries, groupEntries)
		t.Groups = append(t.Groups, roundRobin)
	}

	matchList := t.createMatchList()

	crossGroupRanking := NewCrossGroupRanking(
		entries,
		t.Groups,
		matchList.Matches,
		walkoverScore,
		numQualifications,
	)

	finalRanking := NewGroupPhaseRanking(
		t.Groups,
		numQualifications,
		crossGroupRanking,
		rankingGraph,
	)

	for _, g := range t.Groups {
		rankingGraph.AddEdge(g.FinalRanking, crossGroupRanking)
	}

	t.addTournamentData(matchList, rankingGraph, finalRanking)
}

func (t *GroupPhase) createMatchList() *matchList {
	lastGroup := t.Groups[len(t.Groups)-1]
	maxNumRounds := len(lastGroup.matchList.Rounds)
	maxNumMatches := len(lastGroup.matchList.Matches)

	rounds := make([]*Round, 0, maxNumRounds)
	matches := make([]*Match, 0, len(t.Groups)*maxNumMatches)
	for i := range maxNumRounds {
		groupRounds := collectRounds(i, t.Groups)
		roundMatches := intertwineRounds(groupRounds)
		matches = append(matches, roundMatches...)
		round := &Round{Matches: roundMatches, NestedRounds: groupRounds}
		rounds = append(rounds, round)
	}

	matchList := &matchList{Matches: matches, Rounds: rounds}

	return matchList
}

func collectRounds(roundI int, groups []*RoundRobin) []*Round {
	rounds := make([]*Round, 0, len(groups))
	for _, g := range groups {
		if roundI > len(g.matchList.Rounds)-1 {
			continue
		}
		rounds = append(rounds, g.matchList.Rounds[roundI])
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

func newGroupPhase(
	entries Ranking,
	numGroups, numQualifications int,
	walkoverScore Score,
	rankingGraph *RankingGraph,
) *GroupPhase {
	groupPhase := &GroupPhase{
		BaseTournament: newBaseTournament[*GroupPhaseRanking](entries),
	}
	groupPhase.initTournament(
		entries,
		numGroups,
		numQualifications,
		walkoverScore,
		rankingGraph,
	)

	matchList := groupPhase.matchList

	editingPolicy := &RoundRobinEditingPolicy{
		matches: matchList.Matches,
	}

	withdrawalPolicy := &RoundRobinWithdrawalPolicy{
		matchList: matchList,
	}

	groupPhase.addPolicies(editingPolicy, withdrawalPolicy)

	groupPhase.Update(nil)

	return groupPhase
}
