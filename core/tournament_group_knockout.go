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

type KnockOutTournament interface {
	getBase() *BaseTournament[*EliminationRanking]
}

type KnockoutBuilder func(entries Ranking, rankingGraph *RankingGraph) (KnockOutTournament, error)

type GroupKnockout struct {
	BaseTournament[*GroupKnockoutRanking]
	GroupPhase         *GroupPhase
	KnockOut           *BaseTournament[*EliminationRanking]
	KnockOutTournament KnockOutTournament
}

func (t *GroupKnockout) initTournament(
	entries Ranking,
	knockoutBuilder KnockoutBuilder,
	numGroups, numQualifications int,
	walkoverScore Score,
) error {
	numEntries := len(entries.Ranks())

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

	t.GroupPhase = newGroupPhase(entries, numGroups, numQualifications, walkoverScore, rankingGraph)

	groupPhaseRanking := t.GroupPhase.FinalRanking
	qualificationRanking := NewGroupQualificationRanking(groupPhaseRanking, rankingGraph)

	knockOutTournament, err := knockoutBuilder(qualificationRanking, rankingGraph)
	if err != nil {
		return err
	}
	t.KnockOutTournament = knockOutTournament
	t.KnockOut = knockOutTournament.getBase()

	matchList := t.createMatchList()

	finalRanking := NewGroupKnockoutRanking(t.GroupPhase, t.KnockOut)
	rankingGraph.AddVertex(finalRanking)
	rankingGraph.AddEdge(t.KnockOut.FinalRanking, finalRanking)

	t.addTournamentData(matchList, rankingGraph, finalRanking)

	return nil
}

func (t *GroupKnockout) createMatchList() *matchList {
	ml1 := t.GroupPhase.matchList
	ml2 := t.KnockOut.matchList

	matches := slices.Concat(ml1.Matches, ml2.Matches)
	rounds := slices.Concat(ml1.Rounds, ml2.Rounds)

	return &matchList{Matches: matches, Rounds: rounds}
}

type GroupKnockoutEditingPolicy struct {
	editableMatches []*Match
	groupPhase      *GroupPhase
	knockOut        *BaseTournament[*EliminationRanking]
}

func (e *GroupKnockoutEditingPolicy) UpdateEditableMatches() {
	knockOutStarted := e.knockOut.matchList.MatchesStarted()
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

// Withdraws the given player from the tournament.
// The specific matches that the player was withdrawn from
// are returned.
func (w *GroupKnockoutWithdrawalPolicy) WithdrawPlayer(player Player) []*Match {
	withdrawMatches := w.ListWithdrawMatches(player)
	withdrawFromMatches(player, withdrawMatches)
	return withdrawMatches
}

// Attempts to reenter the player into the tournament.
// On success the specific matches that the player
// was reentered into are returned.
func (w *GroupKnockoutWithdrawalPolicy) ReenterPlayer(player Player) []*Match {
	reenterMatches := w.ListReenterMatches(player)
	reenterIntoMatches(player, reenterMatches)
	return reenterMatches
}

func (w *GroupKnockoutWithdrawalPolicy) ListWithdrawMatches(player Player) []*Match {
	knockOutStarted := w.knockOut.matchList.MatchesStarted()
	if knockOutStarted {
		return w.knockOut.ListWithdrawMatches(player)
	} else {
		return w.groupPhase.ListWithdrawMatches(player)
	}
}

func (w *GroupKnockoutWithdrawalPolicy) ListReenterMatches(player Player) []*Match {
	knockOutStarted := w.knockOut.matchList.MatchesStarted()
	if knockOutStarted {
		return w.knockOut.ListReenterMatches(player)
	} else {
		return w.groupPhase.ListReenterMatches(player)
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
		groupPhase: groupKnockout.GroupPhase,
		knockOut:   groupKnockout.KnockOut,
	}

	withdrawalPolicy := &GroupKnockoutWithdrawalPolicy{
		groupPhase: groupKnockout.GroupPhase,
		knockOut:   groupKnockout.KnockOut,
	}

	groupKnockout.addPolicies(editingPolicy, withdrawalPolicy)

	groupKnockout.Update(nil)

	return groupKnockout, nil
}
