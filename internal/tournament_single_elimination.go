package internal

import "slices"

type EliminationMatchMaker struct {
	EliminationGraph *EliminationGraph

	WinnerRankings map[*Match]*WinnerRanking
}

// Creates a MatchList and a Qualification Graph
// The participating players are passed as a Ranking
// of entries as well as some arbitrary tournament mode
// specific settings.
//
// Given the same entries and settings, this method always
// returns the same MatchList, RankingGraph and final Ranking.
// Any RNG values are seeded by a value from the settings.
//
// Can return an error when the ranking is empty or
// invalid settings are passed.
func (m *EliminationMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	rankingGraph := NewRankingGraph(entries)

	m.WinnerRankings = make(map[*Match]*WinnerRanking)

	m.EliminationGraph = NewEliminationGraph()

	balancedEntries := NewBalancedRanking(entries, rankingGraph)
	entrySlots := balancedEntries.GetRanks()

	numRounds := getNumRounds(len(entrySlots))

	rounds := make([]*Round, 0, numRounds)
	for i := range numRounds {
		round := &Round{}
		rounds = append(rounds, round)
		if i == 0 {
			round.Matches = CreateSeededMatches(entrySlots)
		} else {
			round.Matches = CreatePairedMatches(entrySlots)
		}

		entrySlots = createWinnerSlots(round.Matches, rankingGraph, m.WinnerRankings)

		if i == 0 {
			for _, s := range entrySlots {
				rankingGraph.AddEdge(balancedEntries, s.Placement().Ranking())
			}
		} else {
			lastRound := rounds[i-1]
			linkMatches(lastRound.Matches, round.Matches, m.EliminationGraph)
		}
	}

	numMatches := getNumMatches(numRounds)
	matches := make([]*Match, 0, numMatches)
	for _, r := range rounds {
		matches = append(matches, r.Matches...)
	}

	matchList := &MatchList{Matches: matches, Rounds: rounds}

	finals := matches[len(matches)-1]
	finalsRanking := m.WinnerRankings[finals]
	finalRanking := NewEliminationRanking(matchList, entries, finalsRanking, rankingGraph)

	return matchList, rankingGraph, finalRanking, nil
}

// Creates matches with the slots taken pair-wise from
// the entrySlots
func CreatePairedMatches(entrySlots []*Slot) []*Match {
	matches := make([]*Match, 0, len(entrySlots)<<1)
	for i := 0; i < len(entrySlots); i += 2 {
		match := NewMatch(entrySlots[i], entrySlots[i+1])
		matches = append(matches, match)
	}

	return matches
}

// Creates matches with the slots being arranged for
// a seeded elimination round
func CreateSeededMatches(entrySlots []*Slot) []*Match {
	numRounds := getNumRounds(len(entrySlots))
	seedMatchups := arrangeSeeds(numRounds)
	matches := make([]*Match, 0, len(seedMatchups))

	for _, matchup := range seedMatchups {
		match := NewMatch(entrySlots[matchup.seed1], entrySlots[matchup.seed2])
		matches = append(matches, match)
	}

	return matches
}

type seedMatchup struct {
	seed1 int
	seed2 int
}

// Arranges the seeds for the first elimination round of
// a total of numRounds.
//
// The arrangement ensures that the top 2 seeds can only
// meet in the final, the top 4 seeds can only meet
// in the semi-final, etc...
//
// More info: https://en.wikipedia.org/wiki/Single-elimination_tournament#Seeding
func arrangeSeeds(numRounds int) []*seedMatchup {
	// Start with the final between the first two seeds
	matchups := []*seedMatchup{{0, 1}}
	totalSeeds := 2

	// Work down the tournament tree by round (semis, quarters, ...)
	for i := 1; i < numRounds; i += 1 {
		nextMatchups := make([]*seedMatchup, 0, totalSeeds)
		totalSeeds *= 2
		for _, parent := range matchups {
			s1 := parent.seed1
			s2 := parent.seed2

			nextMatchups = append(
				nextMatchups,
				&seedMatchup{s1, totalSeeds - 1 - s1},
				&seedMatchup{s2, totalSeeds - 1 - s2},
			)
		}

		matchups = nextMatchups
	}

	return matchups
}

func createWinnerSlots(matches []*Match, rankingGraph *RankingGraph, winnerRankings map[*Match]*WinnerRanking) []*Slot {
	slots := make([]*Slot, 0, len(matches))
	for _, m := range matches {
		ranking := NewWinnerRanking(m, rankingGraph)
		if winnerRankings != nil {
			winnerRankings[m] = ranking
		}
		placement := NewPlacement(ranking, 0)
		slot := NewPlacementSlot(placement)
		slots = append(slots, slot)
	}
	return slots
}

func linkMatches(round, followingRound []*Match, eliminationGraph *EliminationGraph) {
	for i := range followingRound {
		match1 := round[2*i]
		match2 := round[2*i+1]
		followingMatch := followingRound[i]

		eliminationGraph.AddVertex(match1)
		eliminationGraph.AddVertex(match2)
		eliminationGraph.AddVertex(followingMatch)

		eliminationGraph.AddEdge(match1, followingMatch)
		eliminationGraph.AddEdge(match2, followingMatch)
	}
}

func getNumRounds(numSlots int) int {
	rounds := 0
	for numSlots > 1 {
		numSlots >>= 1
		rounds += 1
	}
	return rounds
}

func getNumMatches(numRounds int) int {
	numMatches := 0
	for i := range numRounds {
		numMatches += 1 << i
	}
	return numMatches
}

type EliminationEditingPolicy struct {
	editableMatches  []*Match
	matchList        *MatchList
	eliminationGraph *EliminationGraph
}

func (e *EliminationGraph) nextPlayableMatches(match *Match) []*Match {
	nextMatches := e.GetDependants(match)

	skipped := make([]*Match, 0, 4)
	for _, m := range nextMatches {
		if m.HasBye() || m.IsWalkover() {
			skipped = append(skipped, m)
		}
	}

	afterSkipped := make([]*Match, 0, 4)
	for _, m := range skipped {
		afterSkipped = append(afterSkipped, e.nextPlayableMatches(m)...)
	}

	playable := make([]*Match, 0, len(nextMatches)+len(afterSkipped)-len(skipped))
	for _, m := range nextMatches {
		if !slices.Contains(skipped, m) {
			playable = append(playable, m)
		}
	}
	playable = append(playable, afterSkipped...)

	return playable
}

// Returns the comprehensive list of matches that are editable
func (e *EliminationEditingPolicy) EditableMatches() []*Match {
	return e.editableMatches
}

// Updates the return value of EditableMatches
func (e *EliminationEditingPolicy) Update() {
	matches := e.matchList.Matches
	editable := make([]*Match, 0, len(matches))

	for _, m := range matches {
		if e.isEditable(m) {
			editable = append(editable, m)
		}
	}

	e.editableMatches = editable
}

func (e *EliminationEditingPolicy) isEditable(match *Match) bool {
	winner, _ := match.GetWinner()
	if winner == nil || match.IsWalkover() || match.HasBye() {
		return false
	}

	nextMatches := e.eliminationGraph.nextPlayableMatches(match)

	editable := true
	for _, m := range nextMatches {
		if !m.StartTime.IsZero() {
			editable = false
			break
		}
	}

	return editable
}

type EliminationWithdrawalPolicy struct {
	tournament       *SingleElimination
	eliminationGraph *EliminationGraph
}

// Withdraws the given player from the tournament.
// The specific matches that the player was withdrawn from
// are returned.
func (e *EliminationWithdrawalPolicy) WithdrawPlayer(player Player) []*Match {
	playerMatches := e.tournament.MatchesOfPlayer(player)
	var walkoverMatch *Match

	for _, m := range playerMatches {
		winner, _ := m.GetWinner()
		if winner == nil {
			walkoverMatch = m
			break
		}

		if m.HasDrawnBye() || !(m.HasBye() || m.IsWalkover()) {
			continue
		}

		nextMatches := e.eliminationGraph.nextPlayableMatches(m)
		nextMatchesStarted := MatchesStarted(nextMatches...)

		walkoverEffective := len(nextMatches) == 0 || nextMatchesStarted

		if !walkoverEffective {
			walkoverMatch = m
			break
		}
	}

	if walkoverMatch == nil {
		return nil
	} else {
		walkoverMatch.WithdrawnPlayers = append(walkoverMatch.WithdrawnPlayers, player)
		return []*Match{walkoverMatch}
	}
}

// Attempts to reenter the player into the tournament.
// On success the specific matches that the player
// was reentered into are returned.
func (e *EliminationWithdrawalPolicy) ReenterPlayer(player Player) []*Match {
	matches := e.tournament.MatchList.Matches
	withdrawnMatches := make([]*Match, 0, 1)
	for _, m := range matches {
		if m.IsPlayerWithdrawn(player) {
			withdrawnMatches = append(withdrawnMatches, m)
		}
	}

	reenteredMatches := make([]*Match, 0, len(withdrawnMatches))
	for _, m := range withdrawnMatches {
		nextMatches := e.eliminationGraph.nextPlayableMatches(m)
		nextMatchesStarted := MatchesStarted(nextMatches...)
		if !nextMatchesStarted {
			reenteredMatches = append(reenteredMatches, m)
			m.WithdrawnPlayers = slices.DeleteFunc(m.WithdrawnPlayers, func(p Player) bool { return p == player })
		}
	}

	return reenteredMatches
}

type SingleElimination struct {
	BaseTournament
}

func NewSingleElimination(entries Ranking) *SingleElimination {
	matchMaker := &EliminationMatchMaker{}
	matchList, rankingGraph, finalRanking, _ := matchMaker.MakeMatches(entries, nil)
	eliminationGraph := matchMaker.EliminationGraph

	editingPolicy := &EliminationEditingPolicy{
		matchList:        matchList,
		eliminationGraph: eliminationGraph,
	}
	editingPolicy.Update()

	tournament := BaseTournament{
		Entries:       entries,
		FinalRanking:  finalRanking,
		MatchMaker:    matchMaker,
		MatchList:     matchList,
		RankingGraph:  rankingGraph,
		EditingPolicy: editingPolicy,
	}

	singleElimination := &SingleElimination{
		BaseTournament: tournament,
	}

	singleElimination.WithdrawalPolicy = &EliminationWithdrawalPolicy{
		tournament:       singleElimination,
		eliminationGraph: eliminationGraph,
	}
	singleElimination.Update(nil)

	return singleElimination
}
