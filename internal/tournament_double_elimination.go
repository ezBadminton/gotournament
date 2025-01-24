package internal

import "slices"

type DoubleElimination struct {
	BaseTournament[*EliminationRanking]
	WinnerBracket    *SingleElimination
	EliminationGraph *EliminationGraph

	WinnerRankings map[*Match]*WinnerRanking

	loserRounds [][]*Match
	final       *Match
}

func (t *DoubleElimination) InitTournament(
	entries Ranking,
	rankingGraph *RankingGraph,
) {
	t.WinnerBracket = createSingleElimination(entries, true, rankingGraph)
	t.RankingGraph = t.WinnerBracket.RankingGraph
	t.EliminationGraph = t.WinnerBracket.EliminationGraph
	t.WinnerRankings = t.WinnerBracket.WinnerRankings

	numWinnerRounds := len(t.WinnerBracket.Rounds)

	t.loserRounds = make([][]*Match, 0, 2*(numWinnerRounds-1))
	for range numWinnerRounds - 1 {
		t.loserRounds = append(t.loserRounds, t.createMinorLoserRound())
		t.loserRounds = append(t.loserRounds, t.createMajorLoserRound())
	}

	t.final = t.createFinal()
	matchList := t.createMatchList()

	finalsRankins := []Ranking{t.WinnerRankings[t.final]}
	finalRanking := NewEliminationRanking(
		matchList,
		entries,
		finalsRankins,
		t.RankingGraph,
	)

	t.addTournamentData(matchList, t.RankingGraph, finalRanking)
}

func (t *DoubleElimination) createMinorLoserRound() []*Match {
	var lastMajor []*Match
	var targetRank int
	if len(t.loserRounds) == 0 {
		lastMajor = t.WinnerBracket.Rounds[0].Matches
		targetRank = 1
	} else {
		lastMajor = t.loserRounds[len(t.loserRounds)-1]
		targetRank = 0
	}

	slots := createWinnerRankingSlots(lastMajor, targetRank, t.RankingGraph, t.WinnerRankings)
	matches := CreatePairedMatches(slots)

	linkMatches(lastMajor, matches, t.EliminationGraph)

	return matches
}

func (t *DoubleElimination) createMajorLoserRound() []*Match {
	majorI := len(t.loserRounds) / 2
	winnerRound := t.WinnerBracket.Rounds[majorI+1].Matches
	lastMinor := t.loserRounds[len(t.loserRounds)-1]

	loserSlots := createWinnerRankingSlots(winnerRound, 1, t.RankingGraph, t.WinnerRankings)
	minorSlots := createWinnerRankingSlots(lastMinor, 0, t.RankingGraph, t.WinnerRankings)

	if majorI%2 == 0 {
		// Every second major round swap the upper and lower bracket halves
		// to prevent rematches
		swapHalves(loserSlots)
		winnerRound = slices.Clone(winnerRound)
		swapHalves(winnerRound)
	}

	matches := make([]*Match, 0, len(loserSlots))
	for i := range len(loserSlots) {
		match := NewMatch(loserSlots[i], minorSlots[i])
		matches = append(matches, match)

		t.EliminationGraph.AddVertex(match)
		t.EliminationGraph.AddEdge(winnerRound[i], match)
		t.EliminationGraph.AddEdge(lastMinor[i], match)
	}

	return matches
}

func (t *DoubleElimination) createFinal() *Match {
	upperFinal := t.WinnerBracket.Matches[len(t.WinnerBracket.Matches)-1]
	lowerFinal := t.loserRounds[len(t.loserRounds)-1][0]

	winners := createWinnerRankingSlots(
		[]*Match{upperFinal, lowerFinal},
		0,
		t.RankingGraph,
		t.WinnerRankings,
	)

	final := NewMatch(winners[0], winners[1])
	_ = createWinnerRankingSlots(
		[]*Match{final},
		0,
		t.RankingGraph,
		t.WinnerRankings,
	)

	return final
}

func (t *DoubleElimination) createMatchList() *MatchList {
	rounds := make([]*Round, 0, 2*len(t.WinnerBracket.Rounds))
	matches := make([]*Match, 0, 4*len(t.WinnerBracket.Rounds)-2)

	for i, r := range t.WinnerBracket.Rounds {
		loserRoundI := i - 1
		if loserRoundI >= 0 {
			minorMatches := t.loserRounds[2*loserRoundI]
			majorMatches := t.loserRounds[2*loserRoundI+1]

			winnerAndMinorRound := combineRounds(r, minorMatches)
			majorLoserRound := &Round{Matches: majorMatches}

			rounds = append(rounds, winnerAndMinorRound, majorLoserRound)
			matches = append(matches, winnerAndMinorRound.Matches...)
			matches = append(matches, majorLoserRound.Matches...)
		} else {
			rounds = append(rounds, r)
			matches = append(matches, r.Matches...)
		}
	}

	finalRound := &Round{Matches: []*Match{t.final}}
	rounds = append(rounds, finalRound)
	matches = append(matches, t.final)

	matchList := &MatchList{Matches: matches, Rounds: rounds}

	return matchList
}

func combineRounds(winnerRound *Round, minorLoserMatches []*Match) *Round {
	minorLoserRound := &Round{Matches: minorLoserMatches}

	matches := slices.Concat(winnerRound.Matches, minorLoserMatches)
	round := &Round{Matches: matches, NestedRounds: []*Round{winnerRound, minorLoserRound}}

	return round
}

func swapHalves[S ~[]E, E any](s S) {
	for i, j := 0, len(s)/2; j < len(s); i, j = i+1, j+1 {
		s[i], s[j] = s[j], s[i]
	}
}

func createDoubleElimination(entries Ranking, rankingGraph *RankingGraph) *DoubleElimination {
	doubleElimination := &DoubleElimination{
		BaseTournament: NewBaseTournament[*EliminationRanking](entries),
	}
	doubleElimination.InitTournament(entries, rankingGraph)

	matchList := doubleElimination.MatchList
	eliminationGraph := doubleElimination.EliminationGraph

	editingPolicy := &EliminationEditingPolicy{
		matchList:        matchList,
		eliminationGraph: eliminationGraph,
	}

	withdrawalPolicy := &EliminationWithdrawalPolicy{
		matchList:        matchList,
		eliminationGraph: eliminationGraph,
	}

	doubleElimination.addPolicies(editingPolicy, withdrawalPolicy)
	doubleElimination.Update(nil)

	return doubleElimination
}

func NewDoubleElimination(entries Ranking) *DoubleElimination {
	return createDoubleElimination(entries, nil)
}

func NewGroupKnockoutDoubleElimination(entries Ranking, rankingGraph *RankingGraph) *DoubleElimination {
	return createDoubleElimination(entries, rankingGraph)
}
