package internal

import (
	"slices"
)

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

type SingleEliminationWithConsolation struct {
	BaseTournament[*EliminationRanking]
	MainBracket      *ConsolationBracket
	Brackets         []*ConsolationBracket
	EliminationGraph *EliminationGraph
}

func (t *SingleEliminationWithConsolation) initTournament(
	entries Ranking,
	numConsolationRounds, placesToPlayOut int,
	rankingGraph *RankingGraph,
) {
	t.Brackets = make([]*ConsolationBracket, 0, 16)

	mainElimination, err := createSingleElimination(entries, true, rankingGraph)
	if err != nil {
		panic("could not create main elimination")
	}
	t.MainBracket = newBracket(mainElimination)
	t.RankingGraph = mainElimination.RankingGraph
	t.EliminationGraph = mainElimination.EliminationGraph

	t.createConsolationBrackets(t.MainBracket, 0, numConsolationRounds, placesToPlayOut)

	t.Brackets = append(t.Brackets, t.MainBracket)
	slices.Reverse(t.Brackets)

	matchList := t.createMatchList()

	finalsRankings := make([]Ranking, 0, len(t.Brackets))
	for _, b := range t.Brackets {
		finals := b.MatchList.Matches[len(b.MatchList.Matches)-1]
		finalsRanking := b.WinnerRankings[finals]
		finalsRankings = append(finalsRankings, finalsRanking)
	}

	finalRanking := NewEliminationRanking(matchList, entries, finalsRankings, t.RankingGraph)

	t.addTournamentData(matchList, t.RankingGraph, finalRanking)
}

func (t *SingleEliminationWithConsolation) createConsolationBrackets(
	winnerBracket *ConsolationBracket,
	depth int,
	numConsolationRounds, placesToPlayOut int,
) {
	consolationDepth := numConsolationRounds - depth

	finalsInBracket := len(t.Brackets) + depth + 1
	playOutDepth := placesToPlayOut - 2*finalsInBracket

	if consolationDepth <= 0 && playOutDepth <= 0 {
		return
	}

	numRoundsToConsole := len(winnerBracket.MatchList.Rounds) - 1

	if consolationDepth <= 0 {
		numFinalsRequired := placesToFinals(playOutDepth)
		numRoundsToConsole = min(numRoundsToConsole, finalsToBrackets(numFinalsRequired))
	}

	startIndex := len(winnerBracket.MatchList.Rounds) - numRoundsToConsole
	roundsToConsole := winnerBracket.MatchList.Rounds[startIndex:]

	for _, r := range slices.Backward(roundsToConsole) {
		consolationBracket := t.createBracketFromRound(r, winnerBracket)
		if consolationBracket == nil {
			break
		}

		t.createConsolationBrackets(
			consolationBracket,
			depth+1,
			numConsolationRounds,
			placesToPlayOut,
		)

		winnerBracket.Consolations = append(winnerBracket.Consolations, consolationBracket)
		t.Brackets = append(t.Brackets, consolationBracket)
	}

	slices.Reverse(winnerBracket.Consolations)
}

func (t *SingleEliminationWithConsolation) createBracketFromRound(
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
	consolationElimination, err := newConsolationElimination(consolationEntries, t.RankingGraph)
	if err != nil {
		panic("could not create consolation bracket")
	}
	consolationBracket := newBracket(consolationElimination)

	for _, r := range winnerRankings {
		t.RankingGraph.AddEdge(r, consolationEntries)
	}

	roundIndex := slices.Index(winnerBracket.MatchList.Rounds, winnerRound)
	prevRound := winnerBracket.MatchList.Rounds[roundIndex-1]
	linkMatches(prevRound.Matches, consolationBracket.MatchList.Rounds[0].Matches, t.EliminationGraph)

	return consolationBracket
}

func (t *SingleEliminationWithConsolation) createMatchList() *MatchList {
	stack := make([]*ConsolationBracket, 0, 32)
	stack = append(stack, t.MainBracket)

	// Order the brackets by their highest achievable rank (main bracket, match for 3rd, bracket for 5th, ...)
	orderedBrackets := make([]*ConsolationBracket, 0, 32)
	for l := 1; l > 0; l = len(stack) {
		current := stack[l-1]
		stack = stack[:l-1]
		stack = append(stack, current.Consolations...)

		orderedBrackets = append(orderedBrackets, current)
	}

	// Group the finals/semi-finals/etc. of all brackets together
	maxNumRounds := len(t.MainBracket.MatchList.Rounds)
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

func createSingleEliminationWithConsolation(
	entries Ranking,
	numConsolationRounds, placesToPlayOut int,
	rankingGraph *RankingGraph,
) (*SingleEliminationWithConsolation, error) {
	if len(entries.GetRanks()) < 2 {
		return nil, ErrTooFewEntries
	}

	consolationTournament := &SingleEliminationWithConsolation{
		BaseTournament: newBaseTournament[*EliminationRanking](entries),
	}
	consolationTournament.initTournament(
		entries,
		numConsolationRounds,
		placesToPlayOut,
		rankingGraph,
	)

	matchList := consolationTournament.MatchList
	eliminationGraph := consolationTournament.EliminationGraph

	editingPolicy := &EliminationEditingPolicy{
		matchList:        matchList,
		eliminationGraph: eliminationGraph,
	}

	withdrawalPolicy := &EliminationWithdrawalPolicy{
		matchList:        matchList,
		eliminationGraph: eliminationGraph,
	}

	consolationTournament.addPolicies(editingPolicy, withdrawalPolicy)

	consolationTournament.Update(nil)

	return consolationTournament, nil
}

func NewSingleEliminationWithConsolation(
	entries Ranking,
	numConsolationRounds, placesToPlayOut int,
) (*SingleEliminationWithConsolation, error) {
	return createSingleEliminationWithConsolation(
		entries,
		numConsolationRounds,
		placesToPlayOut,
		nil,
	)
}

func SingleEliminationWithConsolationBuilder(
	numConsolationRounds, placesToPlayOut int,
) KnockoutBuilder {
	builder := func(entries Ranking, rankingGraph *RankingGraph) (*BaseTournament[*EliminationRanking], error) {
		tournament, err := createSingleEliminationWithConsolation(
			entries,
			numConsolationRounds,
			placesToPlayOut,
			rankingGraph,
		)

		return &tournament.BaseTournament, err
	}

	return builder
}
