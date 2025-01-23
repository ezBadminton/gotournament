package internal

import (
	"reflect"
	"slices"
	"testing"
)

func TestGroupKnockoutStructure(t *testing.T) {
	players, err := PlayerSlice(12)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		NewGroupKnockoutSingleElimination,
		4,
		8,
		NewScore(42, 0),
	)

	eq1 := len(tournament.knockOut.MatchList.Rounds) == 3
	if !eq1 {
		t.Fatal("Single elimination knockout has unexpected amount of rounds")
	}

	tournament.MatchList.Matches[4].StartMatch()
	tournament.MatchList.Matches[4].EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	ranks := tournament.FinalRanking.TiedRanks()

	eq1 = len(ranks) == 3
	eq2 := ranks[1][0].Player() == players[0]
	eq3 := ranks[2][0].Player() == players[8]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The final ranking did not update properly")
	}
}

func TestGroupKnockoutQualification(t *testing.T) {
	players, err := PlayerSlice(12)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		SingleEliminationWithConsolationBuilder(0, 3),
		4,
		8,
		NewScore(42, 0),
	)

	for _, m := range tournament.Matches[:24] {
		if !m.HasBye() {
			p1 := slices.Index(players, m.Slot1.Player())
			p2 := slices.Index(players, m.Slot2.Player())
			var score Score = NewScore(max(p1, p2)+1, 0)
			if p2 > p1 {
				score = score.Invert()
			}

			m.StartMatch()
			m.EndMatch(score)
		}
	}
	tournament.Update(nil)

	eq1 := tournament.groupPhase.FinalRanking.QualificationComplete
	if !eq1 {
		t.Fatal("The group phase is not marked as completed despite the matches being finished")
	}

	firstKnockOutRound := tournament.knockOut.Rounds[0]
	m := firstKnockOutRound.Matches[0]
	eq1 = m.Slot1.Player() == players[8] && m.Slot2.Player() == players[6]
	m = firstKnockOutRound.Matches[1]
	eq2 := m.Slot1.Player() == players[11] && m.Slot2.Player() == players[5]
	m = firstKnockOutRound.Matches[2]
	eq3 := m.Slot1.Player() == players[9] && m.Slot2.Player() == players[7]
	m = firstKnockOutRound.Matches[3]
	eq4 := m.Slot1.Player() == players[10] && m.Slot2.Player() == players[4]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The qualified players are not seeded correctly in the knockout")
	}
}

func TestGroupKnockoutUnbalancedQualifications(t *testing.T) {
	players, err := PlayerSlice(12)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		NewGroupKnockoutSingleElimination,
		3,
		6,
		NewScore(42, 0),
	)

	for _, m := range tournament.Matches[:18] {
		p1 := slices.Index(players, m.Slot1.Player())
		p2 := slices.Index(players, m.Slot2.Player())
		var score Score = NewScore(max(p1, p2)+1, 0)
		if p2 > p1 {
			score = score.Invert()
		}

		m.StartMatch()
		m.EndMatch(score)
	}
	tournament.Update(nil)

	knockOutMatches := tournament.knockOut.Matches

	m := knockOutMatches[0]
	eq1 := m.Slot1.Player() == players[11] && m.Slot2.IsBye()
	m = knockOutMatches[1]
	eq2 := m.Slot1.Player() == players[9] && m.Slot2.Player() == players[7]
	m = knockOutMatches[2]
	eq3 := m.Slot1.Player() == players[10] && m.Slot2.IsBye()
	m = knockOutMatches[3]
	eq4 := m.Slot1.Player() == players[6] && m.Slot2.Player() == players[8]
	m = knockOutMatches[4]
	eq5 := m.Slot1.Player() == players[11] && m.Slot2.Player() == nil
	m = knockOutMatches[5]
	eq6 := m.Slot1.Player() == players[10] && m.Slot2.Player() == nil
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 || !eq6 {
		t.Fatal("The qualified players are not seeded correctly in the knockout")
	}
}

func TestGroupKnockoutContestedQualifications(t *testing.T) {
	players, err := PlayerSlice(12)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		NewGroupKnockoutSingleElimination,
		3,
		5,
		NewScore(42, 0),
	)

	for _, m := range tournament.Matches[:18] {
		p1 := slices.Index(players, m.Slot1.Player())
		p2 := slices.Index(players, m.Slot2.Player())
		var score Score = NewScore(max(p1, p2)+1, 0)
		if p2 > p1 {
			score = score.Invert()
		}

		m.StartMatch()
		m.EndMatch(score)
	}
	tournament.Update(nil)

	knockOutEntrySlots := tournament.knockOut.Entries.GetRanks()
	eq1 := knockOutEntrySlots[0].Player() == players[11]
	eq2 := knockOutEntrySlots[1].Player() == players[10]
	eq3 := knockOutEntrySlots[2].Player() == players[9]
	eq4 := knockOutEntrySlots[3].Player() == players[8]
	eq5 := knockOutEntrySlots[4].Player() == players[7]
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 {
		t.Fatal("The qualified players are not the correct ones")
	}
}

func TestGroupKnockoutEditingPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		NewGroupKnockoutSingleElimination,
		2,
		4,
		NewScore(42, 0),
	)

	editableMatches := tournament.EditableMatches()
	eq1 := len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The editable matches are not empty despite no matches being finished")
	}

	groupMatches := tournament.groupPhase.Matches
	for i, m := range groupMatches {
		score := NewScore((i+1)*2, 0)
		m.StartMatch()
		m.EndMatch(score)
	}
	tournament.Update(nil)

	editableMatches = tournament.EditableMatches()
	eq1 = reflect.DeepEqual(editableMatches, groupMatches)
	if !eq1 {
		t.Fatal("The finished group matches are not editable")
	}

	knockoutMatches := tournament.knockOut.Matches
	knockoutMatches[0].StartMatch()
	tournament.Update(nil)

	editableMatches = tournament.EditableMatches()
	eq1 = len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The editiable matches are not cleared after the knockout matches began")
	}

	knockoutMatches[0].EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	editableMatches = tournament.EditableMatches()
	eq1 = len(editableMatches) == 1 && editableMatches[0] == knockoutMatches[0]
	if !eq1 {
		t.Fatal("The finished knockout match is not editable")
	}
}

func TestGroupKnockoutWithdrawalPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewGroupKnockout(
		entries,
		NewGroupKnockoutSingleElimination,
		2,
		4,
		NewScore(42, 0),
	)

	groupMatches := tournament.groupPhase.Matches

	withdrawnMatches := tournament.WithdrawPlayer(players[0])
	tournament.Update(nil)

	eq1 := reflect.DeepEqual(withdrawnMatches, []*Match{groupMatches[0], groupMatches[4], groupMatches[8]})
	if !eq1 {
		t.Fatal("The player was not withdrawn from their group phase matches")
	}

	for i, m := range groupMatches {
		if !m.IsWalkover() {
			score := NewScore((i+1)*2, 0)
			m.StartMatch()
			m.EndMatch(score)
		}
	}
	tournament.Update(nil)

	reenteredMatches := tournament.ReenterPlayer(players[0])
	tournament.Update(nil)
	eq1 = reflect.DeepEqual(reenteredMatches, withdrawnMatches)
	if !eq1 {
		t.Fatal("The player could not reenter their group phase matches despite the knockout not having started")
	}

	tournament.WithdrawPlayer(players[0])
	tournament.Update(nil)

	knockoutMatches := tournament.knockOut.Matches

	knockoutMatches[0].StartMatch()
	reenteredMatches = tournament.ReenterPlayer(players[0])
	tournament.Update(nil)
	eq1 = len(reenteredMatches) == 0
	if !eq1 {
		t.Fatal("The player was not prevented from reentering despite the knockout having started")
	}

	withdrawnMatches = tournament.WithdrawPlayer(players[2])
	eq1 = len(withdrawnMatches) == 0
	if !eq1 {
		t.Fatal("A player who is disqualified was able to withdraw after the knockout started")
	}

	withdrawnMatches = tournament.WithdrawPlayer(players[1])
	tournament.Update(nil)
	eq1 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == knockoutMatches[0]
	if !eq1 {
		t.Fatal("A qualified player was not able to withdraw from their knockout match")
	}
}
