package core

import "slices"

type RoundRobin struct {
	BaseTournament[*MatchMetricRanking]
}

// Creates the matches of a round robin tournament.
// The RoundRobinSettings define a Passes int controlling
// how often all matchups are played through.
func (t *RoundRobin) initTournament(
	entries Ranking,
	passes int,
	walkoverScore Score,
	rankingGraph *RankingGraph,
) error {
	if len(entries.Ranks()) < 1 {
		return ErrTooFewEntries
	}

	if rankingGraph == nil {
		rankingGraph = NewRankingGraph(entries)
	} else {
		rankingGraph.AddVertex(entries)
	}

	evenEntries := NewEvenRanking(entries, rankingGraph)
	entrySlots := evenEntries.Ranks()

	if passes < 1 {
		passes = 1
	}
	numRounds := len(entrySlots) - 1
	numMatches := len(entrySlots) / 2

	rounds := make([]*Round, 0, passes*numRounds)
	for passI := range passes {
		for roundI := range numRounds {
			round := createRound(entrySlots, passI, roundI)
			rounds = append(rounds, round)
		}
	}
	matches := make([]*Match, 0, passes*numRounds*numMatches)
	for _, r := range rounds {
		matches = append(matches, r.Matches...)
	}

	matchList := &matchList{Rounds: rounds, Matches: matches}

	finalRanking := NewRoundRobinRanking(
		evenEntries,
		matches,
		walkoverScore,
		rankingGraph,
	)

	t.addTournamentData(matchList, rankingGraph, finalRanking)

	return nil
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
func (e *RoundRobinEditingPolicy) UpdateEditableMatches() {
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
	matchList *matchList
}

// Withdraws the given player from the tournament.
// The specific matches that the player was withdrawn from
// are returned.
func (w *RoundRobinWithdrawalPolicy) WithdrawPlayer(player Player) []*Match {
	withdrawMatches := w.ListWithdrawMatches(player)
	withdrawFromMatches(player, withdrawMatches)
	return withdrawMatches
}

// Attempts to reenter the player into the tournament.
// On success the specific matches that the player
// was reentered into are returned.
func (w *RoundRobinWithdrawalPolicy) ReenterPlayer(player Player) []*Match {
	reenterMatches := w.ListReenterMatches(player)
	reenterIntoMatches(player, reenterMatches)
	return reenterMatches
}

func (w *RoundRobinWithdrawalPolicy) ListWithdrawMatches(player Player) []*Match {
	withdrawnMatches := w.matchList.MatchesOfPlayer(player)
	return withdrawnMatches
}

// Attempts to reenter the player into the tournament.
// On success the specific matches that the player
// was reentered into are returned.
func (w *RoundRobinWithdrawalPolicy) ListReenterMatches(player Player) []*Match {
	withdrawnMatches := make([]*Match, 0, 5)
	for _, m := range w.matchList.Matches {
		isWithdrawn := slices.ContainsFunc(
			m.WithdrawnPlayers,
			func(p Player) bool { return p.Id() == player.Id() },
		)
		if isWithdrawn {
			withdrawnMatches = append(withdrawnMatches, m)
		}
	}
	return withdrawnMatches
}

func createRoundRobin(entries Ranking, passes int, walkoverScore Score, rankingGraph *RankingGraph) (*RoundRobin, error) {
	roundRobin := &RoundRobin{
		BaseTournament: newBaseTournament[*MatchMetricRanking](entries),
	}
	err := roundRobin.initTournament(
		entries,
		passes,
		walkoverScore,
		rankingGraph,
	)
	if err != nil {
		return nil, err
	}

	matchList := roundRobin.matchList

	editingPolicy := &RoundRobinEditingPolicy{matches: matchList.Matches}

	withdrawalPolicy := &RoundRobinWithdrawalPolicy{matchList: matchList}

	roundRobin.addPolicies(editingPolicy, withdrawalPolicy)

	roundRobin.Update(nil)

	return roundRobin, nil
}

func NewRoundRobin(entries Ranking, passes int, walkoverScore Score) (*RoundRobin, error) {
	return createRoundRobin(entries, passes, walkoverScore, nil)
}

func newGroupRoundRobin(entries Ranking, requiredUntiedRanks int, walkoverScore Score, rankingGraph *RankingGraph) (*RoundRobin, error) {
	tournament, err := createRoundRobin(entries, 1, walkoverScore, rankingGraph)
	if err != nil {
		return nil, err
	}
	tournament.FinalRanking.RequiredUntiedRanks = requiredUntiedRanks
	return tournament, nil
}
