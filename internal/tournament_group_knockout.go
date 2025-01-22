package internal

import "slices"

type KnockoutBuilder func(entires Ranking) *BaseTournament

type GroupKnockoutSettings struct {
	GroupPhaseSettings

	knockoutBuilder KnockoutBuilder
}

type GroupKnockoutMatchMaker struct {
	groupPhase *GroupPhase
	knockOut   *BaseTournament
}

func (m *GroupKnockoutMatchMaker) MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error) {
	gkSettings := settings.(*GroupKnockoutSettings)
	numQualifications := gkSettings.NumQualifications
	numGroups := gkSettings.NumGroups
	walkoverScore := gkSettings.WalkoverScore

	rankingGraph := NewRankingGraph(entries)

	m.groupPhase = NewGroupPhase(entries, numGroups, numQualifications, walkoverScore, rankingGraph)

	groupPhaseRanking := m.groupPhase.FinalRanking.(*GroupPhaseRanking)
	qualificationRanking := NewGroupQualificationRanking(groupPhaseRanking, rankingGraph)

	m.knockOut = gkSettings.knockoutBuilder(qualificationRanking)

	matchList := m.createMatchList()

	finalRanking := NewGroupKnockoutRanking(m.groupPhase, m.knockOut)
	rankingGraph.AddVertex(finalRanking)
	rankingGraph.AddEdge(m.knockOut.FinalRanking, finalRanking)

	return matchList, rankingGraph, finalRanking, nil
}

func (m *GroupKnockoutMatchMaker) createMatchList() *MatchList {
	ml1 := m.groupPhase.MatchList
	ml2 := m.knockOut.MatchList

	matches := slices.Concat(ml1.Matches, ml2.Matches)
	rounds := slices.Concat(ml1.Rounds, ml2.Rounds)

	return &MatchList{Matches: matches, Rounds: rounds}
}
