package internal

import "slices"

type KnockoutBuilder func(entries Ranking, rankingGraph *RankingGraph) *BaseTournament[*EliminationRanking]

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

	t.knockOut = knockoutBuilder(qualificationRanking, rankingGraph)

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

type GroupKnockoutEditingPolicy struct {
	editableMatches []*Match
	groupPhase      *GroupPhase
	knockOut        *BaseTournament[*EliminationRanking]
}

func (e *GroupKnockoutEditingPolicy) UpdateEditableMatches() {
	knockOutStarted := e.knockOut.MatchList.MatchesStarted()
	if knockOutStarted {
		e.knockOut.UpdateEditableMatches()
		e.editableMatches = e.knockOut.EditableMatches()
	} else {
		e.groupPhase.UpdateEditableMatches()
		e.editableMatches = e.groupPhase.EditableMatches()
	}
}

func (e *GroupKnockoutEditingPolicy) EditableMatches() []*Match {
	return e.editableMatches
}

type GroupKnockoutWithdrawalPolicy struct {
	groupPhase *GroupPhase
	knockOut   *BaseTournament[*EliminationRanking]
}

func (w *GroupKnockoutWithdrawalPolicy) WithdrawPlayer(player Player) []*Match {
	knockOutStarted := w.knockOut.MatchList.MatchesStarted()
	if knockOutStarted {
		return w.knockOut.WithdrawPlayer(player)
	} else {
		return w.groupPhase.WithdrawPlayer(player)
	}
}

func (w *GroupKnockoutWithdrawalPolicy) ReenterPlayer(player Player) []*Match {
	knockOutStarted := w.knockOut.MatchList.MatchesStarted()
	if knockOutStarted {
		return w.knockOut.ReenterPlayer(player)
	} else {
		return w.groupPhase.ReenterPlayer(player)
	}
}

func NewGroupKnockout(
	entries Ranking,
	knockoutBuilder KnockoutBuilder,
	numGroups, numQualifications int,
	walkoverScore Score,
) *GroupKnockout {
	groupKnockout := &GroupKnockout{
		BaseTournament: NewBaseTournament[*GroupKnockoutRanking](entries),
	}
	groupKnockout.InitTournament(
		entries,
		knockoutBuilder,
		numGroups,
		numQualifications,
		walkoverScore,
	)

	editingPolicy := &GroupKnockoutEditingPolicy{
		groupPhase: groupKnockout.groupPhase,
		knockOut:   groupKnockout.knockOut,
	}

	withdrawalPolicy := &GroupKnockoutWithdrawalPolicy{
		groupPhase: groupKnockout.groupPhase,
		knockOut:   groupKnockout.knockOut,
	}

	groupKnockout.addPolicies(editingPolicy, withdrawalPolicy)

	groupKnockout.Update(nil)

	return groupKnockout
}
