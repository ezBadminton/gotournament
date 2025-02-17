package core

import (
	"reflect"
	"testing"
	"time"
)

func TestDoubleEliminationRanking(t *testing.T) {
	players, err := PlayerSlice(16)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewDoubleElimination(entries)

	ranks := tournament.FinalRanking.TiedRanks()

	eq1 := len(ranks) == 1 && len(ranks[0]) == len(players)
	if !eq1 {
		t.Fatal("The final ranking does not contain all players who entered")
	}

	for _, m := range tournament.Matches {
		m.StartMatch()
		m.EndMatch(NewScore(1, 0))
	}
	tournament.Update(nil)

	ranks = tournament.FinalRanking.TiedRanks()

	eq1 = len(ranks) == 8
	eq2 := ranks[0][0].Player == players[0]
	eq3 := ranks[1][0].Player == players[1]
	eq4 := ranks[2][0].Player == players[3]
	eq5 := ranks[3][0].Player == players[2]
	eq6 := ranks[4][0].Player == players[6] && ranks[4][1].Player == players[7]
	eq7 := ranks[5][0].Player == players[5] && ranks[5][1].Player == players[4]
	eq8 := ranks[6][0].Player == players[15] && ranks[6][1].Player == players[12]
	eq9 := ranks[6][2].Player == players[14] && ranks[6][3].Player == players[13]
	eq10 := ranks[7][0].Player == players[8] && ranks[7][1].Player == players[11]
	eq11 := ranks[7][2].Player == players[9] && ranks[7][3].Player == players[10]
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 || !eq6 || !eq7 || !eq8 || !eq9 || !eq10 || !eq11 {
		t.Fatal("The final ranking is not as expected after all matches finished")
	}
}

func TestDoubleEliminationEditingPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewDoubleElimination(entries)

	editableMatches := tournament.EditableMatches()

	eq1 := len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The editable matches are not empty despite no matches being finished")
	}

	m1 := tournament.Matches[0]
	m2 := tournament.Rounds[1].Matches[0]
	m3 := tournament.Rounds[1].NestedRounds[1].Matches[0]

	m1.StartMatch()
	m1.EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	editableMatches = tournament.EditableMatches()

	eq1 = len(editableMatches) == 1 && editableMatches[0] == m1
	if !eq1 {
		t.Fatal("The finished match is not editable")
	}

	m2.StartMatch()
	tournament.Update(nil)
	editableMatches = tournament.EditableMatches()

	eq1 = len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The finished match is still editable despite the next match having started")
	}

	m2.StartTime = time.Time{}
	tournament.Update(nil)
	editableMatches = tournament.EditableMatches()

	eq1 = len(editableMatches) == 1 && editableMatches[0] == m1
	if !eq1 {
		t.Fatal("The finished match is not editable")
	}

	m3.StartMatch()
	tournament.Update(nil)
	editableMatches = tournament.EditableMatches()

	eq1 = len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The finished match is still editable despite the next match having started")
	}
}

func TestDoubleEliminationWithdrawalPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewDoubleElimination(entries)

	m1 := tournament.Matches[0]
	m2 := tournament.Rounds[1].Matches[0]
	m3 := tournament.Rounds[1].NestedRounds[1].Matches[0]

	withdrawnMatches := tournament.WithdrawPlayer(players[0])

	eq1 := len(withdrawnMatches) == 1 && withdrawnMatches[0] == m1
	if !eq1 {
		t.Fatal("The player did not withdraw from their first round match")
	}

	reenteredMatches := tournament.ReenterPlayer(players[0])
	eq1 = reflect.DeepEqual(reenteredMatches, withdrawnMatches)
	if !eq1 {
		t.Fatal("The player could not reenter into their first round match")
	}

	m1.StartMatch()
	m1.EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	withdrawnMatches = tournament.WithdrawPlayer(players[0])

	eq1 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == m2
	if !eq1 {
		t.Fatal("The player did not withdrawn from their winner bracket match")
	}

	withdrawnMatches = tournament.WithdrawPlayer(players[7])
	eq1 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == m3
	if !eq1 {
		t.Fatal("The player did not withdraw from their loser bracket match")
	}
}
