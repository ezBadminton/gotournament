package internal

type RoundRobinSettings struct {
	Passes int
}

type RoundRobinMatchMaker struct {
}

// Creates the matches of a round robin tournament.
// The RoundRobinSettings define a Passes int controling
// how often all matchups are played through.
func (m *RoundRobinMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	rankingGraph := NewRankingGraph(entries)

	evenEntries := NewEvenRanking(entries, rankingGraph)
	entrySlots := evenEntries.GetRanks()

	numPasses := settings.(*RoundRobinSettings).Passes
	numRounds := len(entrySlots) - 1
	numMatches := len(entrySlots) / 2

	rounds := make([]*Round, 0, numPasses*numRounds)
	for passI := range numPasses {
		for roundI := range numRounds {
			round := createRound(entrySlots, passI, roundI)
			rounds = append(rounds, round)
		}
	}
	matches := make([]*Match, 0, numPasses*numRounds*numMatches)
	for _, r := range rounds {
		matches = append(matches, r.Matches...)
	}

	matchList := &MatchList{Rounds: rounds, Matches: matches}

	// TODO final ranking

	return matchList, rankingGraph, nil, nil
}

func createRound(entrySlots []*Slot, passI, roundI int) *Round {
	numMatches := len(entrySlots) / 2
	round := &Round{
		Matches: make([]*Match, 0, numMatches),
	}

	for matchI := range numMatches {
		slot1, slot2 := pickOpponents(entrySlots, passI, roundI, matchI)
		match := NewMatch(slot1, slot2)
		round.Matches = append(round.Matches, match)
	}

	return round
}

// Returns the opponents of the specified match by its three indices
// while making sure the share of first-named matches is evenly
// distributed among the players
func pickOpponents(entrySlots []*Slot, passI, roundI, matchI int) (*Slot, *Slot) {
	i1 := matchI
	i2 := len(entrySlots) - 1 - matchI

	i1 = roundRobinCircleIndex(i1, len(entrySlots), roundI)
	i2 = roundRobinCircleIndex(i2, len(entrySlots), roundI)

	slot1 := entrySlots[i1]
	slot2 := entrySlots[i2]

	if matchI == 0 && roundI%2 == 0 {
		slot1, slot2 = slot2, slot1
	}
	if passI%2 == 0 {
		slot1, slot2 = slot2, slot1
	}

	return slot1, slot2
}

// Rotates the given index according to https://en.wikipedia.org/wiki/Round-robin_tournament#Circle_method
func roundRobinCircleIndex(index, length, round int) int {
	if index == 0 {
		return 0
	}
	index += round - 1
	index %= length - 1
	index += 1
	return index
}
