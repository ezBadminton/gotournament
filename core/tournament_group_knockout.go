package core

import (
	"errors"
	"slices"
)

var (
	ErrTooFewGroups  = errors.New("the number of groups has to be at least 1")
	ErrTooManyGroups = errors.New("the number of groups is too large for the amount of entries")
	ErrTooFewQuals   = errors.New("the number of qualifications has to be at least 2")
)

type KnockoutBuilder func(entries Ranking, rankingGraph *RankingGraph) (*BaseTournament[*EliminationRanking], error)

type GroupKnockout struct {
	BaseTournament[*GroupKnockoutRanking]
	groupPhase *GroupPhase
	knockOut   *BaseTournament[*EliminationRanking]
}

func (t *GroupKnockout) initTournament(
	entries Ranking,
	knockoutBuilder KnockoutBuilder,
	numGroups, numQualifications int,
	walkoverScore Score,
) error {
	numEntries := len(entries.GetRanks())

	if numEntries < 2 {
		return ErrTooFewEntries
	}
	if numGroups < 1 {
		return ErrTooFewGroups
	}
	if 2*numGroups > numEntries {
		return ErrTooManyGroups
	}
	if numQualifications < 2 {
		return ErrTooFewQuals
	}

	rankingGraph := NewRankingGraph(entries)

	t.groupPhase = newGroupPhase(entries, numGroups, numQualifications, walkoverScore, rankingGraph)

	groupPhaseRanking := t.groupPhase.FinalRanking
	qualificationRanking := NewGroupQualificationRanking(groupPhaseRanking, rankingGraph)

	knockOut, err := knockoutBuilder(qualificationRanking, rankingGraph)
	if err != nil {
		return err
	}
	t.knockOut = knockOut

	matchList := t.createMatchList()

	finalRanking := NewGroupKnockoutRanking(t.groupPhase, t.knockOut)
	rankingGraph.AddVertex(finalRanking)
	rankingGraph.AddEdge(t.knockOut.FinalRanking, finalRanking)

	t.addTournamentData(matchList, rankingGraph, finalRanking)

	return nil
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

func (e *GroupKnockoutEditingPolicy) updateEditableMatches() {
	knockOutStarted := e.knockOut.MatchList.MatchesStarted()
	if knockOutStarted {
		e.knockOut.updateEditableMatches()
		e.editableMatches = e.knockOut.EditableMatches()
	} else {
		e.groupPhase.updateEditableMatches()
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
) (*GroupKnockout, error) {
	groupKnockout := &GroupKnockout{
		BaseTournament: newBaseTournament[*GroupKnockoutRanking](entries),
	}
	err := groupKnockout.initTournament(
		entries,
		knockoutBuilder,
		numGroups,
		numQualifications,
		walkoverScore,
	)
	if err != nil {
		return nil, err
	}

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

	return groupKnockout, nil
}
