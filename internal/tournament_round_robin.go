package internal

import "slices"

type RoundRobinSettings struct {
	Passes        int
	WalkoverScore Score
}

type RoundRobinMatchMaker struct {
}

// Creates the matches of a round robin tournament.
// The RoundRobinSettings define a Passes int controlling
// how often all matchups are played through.
func (m *RoundRobinMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	rankingGraph := NewRankingGraph(entries)

	evenEntries := NewEvenRanking(entries, rankingGraph)
	entrySlots := evenEntries.GetRanks()

	rrSettings := settings.(*RoundRobinSettings)
	numPasses := rrSettings.Passes
	if numPasses < 1 {
		numPasses = 1
	}
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

	finalRanking := NewMatchMetricRanking(evenEntries, matches, rankingGraph, rrSettings.WalkoverScore)

	return matchList, rankingGraph, finalRanking, nil
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

	if matchI == 0 && roundI%2 != 0 {
		slot1, slot2 = slot2, slot1
	}
	if passI%2 != 0 {
		slot1, slot2 = slot2, slot1
	}

	return slot1, slot2
}

// Rotates the given index according to https://en.wikipedia.org/wiki/Round-robin_tournament#Circle_method
func roundRobinCircleIndex(index, length, round int) int {
	if index == 0 {
		return 0
	}
	index -= 1
	index -= round
	index += length - 1
	index %= length - 1
	index += 1
	return index
}

type RoundRobinEditingPolicy struct {
	editableMatches []*Match
	matches         []*Match
}

// Returns the comprehensive list of matches that are editable
func (e *RoundRobinEditingPolicy) EditableMatches() []*Match {
	return e.editableMatches
}

// Updates the return value of EditableMatches
func (e *RoundRobinEditingPolicy) Update() {
	editableMatches := make([]*Match, 0, len(e.matches))
	for _, m := range e.matches {
		winner, _ := m.GetWinner()
		wo := m.IsWalkover()
		bye := m.HasBye()
		if winner != nil && !wo && !bye {
			editableMatches = append(editableMatches, m)
		}
	}

	e.editableMatches = editableMatches
}

type RoundRobinWithdrawalPolicy struct {
	matchList *MatchList
}

// Withdraws the given player from the tournament.
// The specific matches that the player was withdrawn from
// are returned.
func (w *RoundRobinWithdrawalPolicy) WithdrawPlayer(player Player) []*Match {
	matches := w.matchList.MatchesOfPlayer(player)

	allMatchesComplete := true
	for _, m := range matches {
		winner, _ := m.GetWinner()
		if winner == nil {
			allMatchesComplete = false
			break
		}
	}

	var withdrawnMatches []*Match

	if allMatchesComplete {
		withdrawnMatches = []*Match{}
	} else {
		withdrawnMatches = matches
	}

	for _, m := range withdrawnMatches {
		m.WithdrawnPlayers = append(m.WithdrawnPlayers, player)
	}

	return withdrawnMatches
}

// Attempts to reenter the player into the tournament.
// On success the specific matches that the player
// was reentered into are returned.
func (w *RoundRobinWithdrawalPolicy) ReenterPlayer(player Player) []*Match {
	withdrawnMatches := make([]*Match, 0, 5)
	for _, m := range w.matchList.Matches {
		if slices.Contains(m.WithdrawnPlayers, player) {
			withdrawnMatches = append(withdrawnMatches, m)
		}
	}

	for _, m := range withdrawnMatches {
		m.WithdrawnPlayers = slices.DeleteFunc(m.WithdrawnPlayers, func(p Player) bool { return p == player })
	}

	return withdrawnMatches
}

type RoundRobin struct {
	BaseTournament
}

func NewRoundRobin(entries Ranking, passes int, walkoverScore Score) *RoundRobin {
	settings := &RoundRobinSettings{Passes: passes, WalkoverScore: walkoverScore}
	matchMaker := &RoundRobinMatchMaker{}
	matchList, rankingGraph, finalRanking, _ := matchMaker.MakeMatches(entries, settings)

	editingPolicy := &RoundRobinEditingPolicy{matches: matchList.Matches}
	editingPolicy.Update()

	withdrawalPolicy := &RoundRobinWithdrawalPolicy{matchList: matchList}

	tournament := NewBaseTournament(
		entries,
		finalRanking,
		matchMaker,
		matchList,
		rankingGraph,
		editingPolicy,
		withdrawalPolicy,
	)

	roundRobin := &RoundRobin{BaseTournament: tournament}
	roundRobin.Update(nil)

	return roundRobin
}
