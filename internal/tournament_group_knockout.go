package internal

import "slices"

type KnockoutBuilder func(entires Ranking) *BaseTournament[*EliminationRanking]

type GroupKnockout struct {
	BaseTournament[*GroupKnockoutRanking]
	groupPhase *GroupPhase
	knockOut   *BaseTournament[*EliminationRanking]
}

func (t *GroupKnockout) InitTournament(
	entries Ranking,
	knockoutBuilder KnockoutBuilder,
	numGroups, numQualifications int,
	walkoverScore Score,
) {
	rankingGraph := NewRankingGraph(entries)

	t.groupPhase = NewGroupPhase(entries, numGroups, numQualifications, walkoverScore, rankingGraph)

	groupPhaseRanking := t.groupPhase.FinalRanking
	qualificationRanking := NewGroupQualificationRanking(groupPhaseRanking, rankingGraph)

	t.knockOut = knockoutBuilder(qualificationRanking)

	matchList := t.createMatchList()

	finalRanking := NewGroupKnockoutRanking(t.groupPhase, t.knockOut)
	rankingGraph.AddVertex(finalRanking)
	rankingGraph.AddEdge(t.knockOut.FinalRanking, finalRanking)

	t.addTournamentData(matchList, rankingGraph, finalRanking)
}

func (t *GroupKnockout) createMatchList() *MatchList {
	ml1 := t.groupPhase.MatchList
	ml2 := t.knockOut.MatchList

	matches := slices.Concat(ml1.Matches, ml2.Matches)
	rounds := slices.Concat(ml1.Rounds, ml2.Rounds)

	return &MatchList{Matches: matches, Rounds: rounds}
}
