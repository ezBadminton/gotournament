package internal

import (
	"slices"
)

type ConsolationSettings struct {
	numConsolationRounds, placesToPlayOut int
}

type ConsolationBracket struct {
	*SingleElimination
	Consolations []*ConsolationBracket
}

func newBracket(elimination *SingleElimination) *ConsolationBracket {
	bracket := &ConsolationBracket{
		SingleElimination: elimination,
		Consolations:      make([]*ConsolationBracket, 0, 8),
	}
	return bracket
}

type SingleEliminationWithConsolationMatchMaker struct {
	MainBracket      *ConsolationBracket
	Brackets         []*ConsolationBracket
	RankingGraph     *RankingGraph
	EliminationGraph *EliminationGraph
}

func (m *SingleEliminationWithConsolationMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	consolationSettings := settings.(*ConsolationSettings)

	m.Brackets = make([]*ConsolationBracket, 0, 16)

	mainElimination := NewSingleElimination(entries)
	m.MainBracket = newBracket(mainElimination)
	m.RankingGraph = mainElimination.RankingGraph
	m.EliminationGraph = mainElimination.MatchMaker.(*EliminationMatchMaker).EliminationGraph

	m.createConsolationBrackets(m.MainBracket, 0, consolationSettings)

	m.Brackets = append(m.Brackets, m.MainBracket)
	slices.Reverse(m.Brackets)

	matchList := m.createMatchList()

	finalsRankings := make([]Ranking, 0, len(m.Brackets))
	for _, b := range m.Brackets {
		finals := b.MatchList.Matches[len(b.MatchList.Matches)-1]
		finalsRanking := b.MatchMaker.(*EliminationMatchMaker).WinnerRankings[finals]
		finalsRankings = append(finalsRankings, finalsRanking)
	}

	finalRanking := NewEliminationRanking(matchList, entries, finalsRankings, m.RankingGraph)

	return matchList, m.RankingGraph, finalRanking, nil
}

func (m *SingleEliminationWithConsolationMatchMaker) createConsolationBrackets(
	winnerBracket *ConsolationBracket,
	depth int,
	settings *ConsolationSettings,
) {
	consolationDepth := settings.numConsolationRounds - depth

	finalsInBracket := len(m.Brackets) + depth + 1
	placesToPlayOut := settings.placesToPlayOut - 2*finalsInBracket

	if consolationDepth <= 0 && placesToPlayOut <= 0 {
		return
	}

	numRoundsToConsole := len(winnerBracket.MatchList.Rounds) - 1

	if consolationDepth <= 0 {
		numFinalsRequired := placesToFinals(placesToPlayOut)
		numRoundsToConsole = min(numRoundsToConsole, finalsToBrackets(numFinalsRequired))
	}

	startIndex := len(winnerBracket.MatchList.Rounds) - numRoundsToConsole
	roundsToConsole := winnerBracket.MatchList.Rounds[startIndex:]

	for _, r := range slices.Backward(roundsToConsole) {
		consolationBracket := m.createBracketFromRound(r, winnerBracket)
		if consolationBracket == nil {
			break
		}

		m.createConsolationBrackets(consolationBracket, depth+1, settings)

		winnerBracket.Consolations = append(winnerBracket.Consolations, consolationBracket)
		m.Brackets = append(m.Brackets, consolationBracket)
	}

	slices.Reverse(winnerBracket.Consolations)
}

func (m *SingleEliminationWithConsolationMatchMaker) createBracketFromRound(
	winnerRound *Round,
	winnerBracket *ConsolationBracket,
) *ConsolationBracket {
	winnerRankings := make([]*WinnerRanking, 0, 2*len(winnerRound.Matches))
	for _, m := range winnerRound.Matches {
		for s := range m.Slots {
			winnerRanking := s.placement.Ranking().(*WinnerRanking)
			winnerRankings = append(winnerRankings, winnerRanking)
		}
	}

	losers := make([]*Slot, 0, len(winnerRankings))
	for _, r := range winnerRankings {
		loserPlacement := NewPlacement(r, 1)
		loserSlot := NewPlacementSlot(loserPlacement)
		losers = append(losers, loserSlot)
	}

	if allBye(losers) {
		return nil
	}

	consolationEntries := NewSlotRanking(losers)
	consolationElimination := NewConsolationElimination(consolationEntries, m.RankingGraph)
	consolationBracket := newBracket(consolationElimination)

	for _, r := range winnerRankings {
		m.RankingGraph.AddEdge(r, consolationEntries)
	}

	roundIndex := slices.Index(winnerBracket.MatchList.Rounds, winnerRound)
	prevRound := winnerBracket.MatchList.Rounds[roundIndex-1]
	linkMatches(prevRound.Matches, consolationBracket.MatchList.Rounds[0].Matches, m.EliminationGraph)

	return consolationBracket
}

func (m *SingleEliminationWithConsolationMatchMaker) createMatchList() *MatchList {
	stack := make([]*ConsolationBracket, 0, 32)
	stack = append(stack, m.MainBracket)

	// Order the brackets by their highest achievable rank (main bracket, match for 3rd, bracket for 5th, ...)
	orderedBrackets := make([]*ConsolationBracket, 0, 32)
	for l := 1; l > 0; l = len(stack) {
		current := stack[l-1]
		stack = stack[:l-1]
		stack = append(stack, current.Consolations...)

		orderedBrackets = append(orderedBrackets, current)
	}

	// Group the finals/semi-finals/etc. of all brackets together
	maxNumRounds := len(m.MainBracket.MatchList.Rounds)
	groupedRounds := make([][]*Round, 0, maxNumRounds)
	for i := range maxNumRounds {
		size := 1 << i
		groupedRounds = append(groupedRounds, make([]*Round, 0, size))
	}
	for _, b := range orderedBrackets {
		numRounds := len(b.MatchList.Rounds)
		for i, r := range b.MatchList.Rounds {
			groupI := i + (maxNumRounds - numRounds)
			groupedRounds[groupI] = append(groupedRounds[groupI], r)
		}
	}

	// Make each group into a super round with the smaller rounds being nested
	// This way all finals/semi-finals/etc of all consolation levels are in one super round
	rounds := make([]*Round, 0, len(groupedRounds))
	matches := make([]*Match, 0, len(groupedRounds[0][0].Matches)*len(groupedRounds))
	for _, g := range groupedRounds {
		numMatches := len(g[0].Matches) * len(g)
		round := &Round{
			Matches:      make([]*Match, 0, numMatches),
			NestedRounds: g,
		}
		for _, r := range g {
			round.Matches = append(round.Matches, r.Matches...)
			matches = append(matches, r.Matches...)
		}

		rounds = append(rounds, round)
	}

	matchList := &MatchList{
		Matches: matches,
		Rounds:  rounds,
	}

	return matchList
}

func allBye(slots []*Slot) bool {
	for _, s := range slots {
		if !s.IsBye() {
			return false
		}
	}
	return true
}

// Returns the number of consolation elimination brackets with incrementing amount
// of rounds that produce the given amount of finals or more but not more brackets
// than neccessary.
// A bracket has 2^(rounds-1) finals thus each extra bracket adds double the amount
// of finals than its smaller predecessor.
func finalsToBrackets(numFinals int) int {
	// Function boils down to floor(log2(numFinals)+1)
	numBrackets := 0
	for 1<<numBrackets <= numFinals {
		numBrackets += 1
	}
	return numBrackets
}

// Returns how many finals need to be played to get the given number of places
// played out.
func placesToFinals(numPlaces int) int {
	// Integer math version of ceil(numPlaces / 2.0) (for non-negative)
	numFinals := numPlaces >> 1
	if numPlaces%2 != 0 {
		numFinals += 1
	}
	return numFinals
}

type SingleEliminationWithConsolation struct {
	BaseTournament
}

func NewSingleEliminationWithConsolation(
	entries Ranking,
	numConsolationRounds, placesToPlayOut int,
) *SingleEliminationWithConsolation {
	settings := &ConsolationSettings{numConsolationRounds: numConsolationRounds, placesToPlayOut: placesToPlayOut}

	matchMaker := &SingleEliminationWithConsolationMatchMaker{}
	matchList, rankingGraph, finalRanking, _ := matchMaker.MakeMatches(entries, settings)

	editingPolicy := &EliminationEditingPolicy{
		matchList:        matchList,
		eliminationGraph: matchMaker.EliminationGraph,
	}

	withdrawalPolicy := &EliminationWithdrawalPolicy{
		matchList:        matchList,
		eliminationGraph: matchMaker.EliminationGraph,
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

	consolationTournament := &SingleEliminationWithConsolation{BaseTournament: tournament}
	consolationTournament.Update(nil)

	return consolationTournament
}
