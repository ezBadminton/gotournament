package core

import (
	"fmt"
	"slices"
	"testing"
)

// Run through a 4-player tournament with no special cases
func TestSmallSingleElimination(t *testing.T) {
	players, err := PlayerSlice(4)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleElimination(entries)
	winnerRankings := tournament.WinnerRankings

	semi1 := tournament.MatchList.Matches[0]
	semi2 := tournament.MatchList.Matches[1]

	// Highest vs lowest seed
	eq1 := semi1.Slot1.player == players[0]
	eq2 := semi1.Slot2.player == players[3]
	// Second highest vs second lowest
	eq3 := semi2.Slot1.player == players[1]
	eq4 := semi2.Slot2.player == players[2]

	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The players were assigned the wrong slots according to their seeds.")
	}

	finalRanking := tournament.FinalRanking
	finalRanking.updateRanks()
	if len(finalRanking.TiedRanks()) != 1 {
		t.Fatal("The final ranking does not have one tied rank for all players")
	}

	if len(finalRanking.TiedRanks()[0]) != 4 {
		t.Fatal("The final ranking's tied rank does not contain all players")
	}

	semi1.StartMatch()
	semi1.EndMatch(NewScore(1, 0))
	tournament.Update(winnerRankings[semi1])

	if len(finalRanking.TiedRanks()) != 2 {
		t.Fatal("The final ranking did not update according to semi1 match result")
	}

	eq1 = finalRanking.TiedRanks()[0][0].player == players[0]
	if !eq1 {
		t.Fatal("The final ranking did not put the semi1 winner to the top")
	}

	final := tournament.MatchList.Matches[2]

	eq1 = final.Slot1.player == players[0]
	if !eq1 {
		t.Fatal("The semi1 winner did not advance into the final slot")
	}

	eq1 = final.Slot2.player == nil
	if !eq1 {
		t.Fatal("The second final slot is erroneously occupied")
	}

	semi2.StartMatch()
	semi2.EndMatch(NewScore(0, 1))
	tournament.Update(winnerRankings[semi2])

	eq1 = final.Slot2.player == players[2]
	if !eq1 {
		t.Fatal("The semi2 winner did not advance into the final slot")
	}

	eq1 = finalRanking.TiedRanks()[0][0].player == players[0]
	eq2 = finalRanking.TiedRanks()[0][1].player == players[2]
	eq3 = finalRanking.TiedRanks()[1][0].player == players[3]
	eq4 = finalRanking.TiedRanks()[1][1].player == players[1]
	if !eq1 || !eq2 {
		t.Fatal("The semi winners did not get placed tied in the top rank of the final ranking")
	}
	if !eq3 || !eq4 {
		t.Fatal("The semi losers did not get placed tied in the bottom rank of the final ranking")
	}

	final.StartMatch()
	final.EndMatch(NewScore(1, 0))
	tournament.Update(winnerRankings[final])

	eq1 = finalRanking.TiedRanks()[0][0].player == players[0]
	eq2 = finalRanking.TiedRanks()[1][0].player == players[2]
	eq3 = finalRanking.TiedRanks()[2][0].player == players[3]
	eq4 = finalRanking.TiedRanks()[2][1].player == players[1]
	if !eq1 {
		t.Fatal("The finals winner did not get placed at the top of the final ranking")
	}
	if !eq2 {
		t.Fatal("The finals loser did not get placed at the second place of the final ranking")
	}
	if !eq3 || !eq4 {
		t.Fatal("The semi losers did not get placed tied at the bottom rank of the final ranking")
	}

	fmt.Println("Done")
}

// Test a 6-player tournament that would generate two bye-slots to get
// to a balanced 8 starter slots
func TestSingleEliminationUnbalanced(t *testing.T) {
	players, err := PlayerSlice(6)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleElimination(entries)

	quarter1 := tournament.MatchList.Matches[0]
	quarter2 := tournament.MatchList.Matches[1]
	quarter3 := tournament.MatchList.Matches[2]
	quarter4 := tournament.MatchList.Matches[3]

	semi1 := tournament.MatchList.Matches[4]
	semi2 := tournament.MatchList.Matches[5]

	eq1 := quarter1.HasDrawnBye()
	eq2 := quarter3.HasDrawnBye()
	eq3 := quarter2.HasDrawnBye()
	eq4 := quarter4.HasDrawnBye()
	if !eq1 || !eq2 {
		t.Fatal("The two highest players did not get a bye in the first round")
	}
	if eq3 || eq4 {
		t.Fatal("An additional first round match has a drawn bye slot while it should not")
	}

	eq1 = semi1.Slot1.player == players[0]
	eq2 = semi2.Slot1.player == players[1]
	if !eq1 || !eq2 {
		t.Fatal("The two top seeded players did not move to their semi final slots")
	}

	eq1 = semi1.Slot2.player == nil
	eq2 = semi2.Slot2.player == nil
	if !eq1 || !eq2 {
		t.Fatal("The second semi final slots are not empty")
	}
}

// Test the withdrawal of players
func TestSingleEliminationWithdrawalPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleElimination(entries)

	wp := tournament.WithdrawalPolicy
	ml := tournament.MatchList

	p1 := players[0]
	p2 := players[7]
	//p3 := players[3]
	p4 := players[4]

	withdrawnMatches := wp.WithdrawPlayer(p1)

	eq1 := tournament.MatchList.Matches[0] == withdrawnMatches[0]
	if !eq1 {
		t.Fatal("The first seed player did not withdraw from their first round match")
	}

	tournament.Update(nil)

	eq1 = ml.Rounds[1].Matches[0].Slot1.player == p2
	if !eq1 {
		t.Fatal("The opponent of the withdrawn player did not advance to the next round")
	}

	withdrawnMatches = wp.WithdrawPlayer(p2)
	eq1 = ml.Matches[0] == withdrawnMatches[0]
	if !eq1 {
		t.Fatal("The last seed player did not withdraw from their first round match")
	}

	tournament.Update(nil)

	eq1 = ml.Rounds[1].Matches[0].Slot1.IsBye()
	if !eq1 {
		t.Fatal("The double withdrawal did not make the next round slot a bye")
	}

	reenteredMatches := wp.ReenterPlayer(p2)
	eq1 = ml.Matches[0] == reenteredMatches[0]
	if !eq1 {
		t.Fatal("The last seed player did not reenter into their first round match")
	}

	tournament.Update(nil)

	eq1 = ml.Rounds[1].Matches[0].Slot1.player == p2
	if !eq1 {
		t.Fatal("The reentered player did not advance to the next round")
	}

	reenteredMatches = wp.ReenterPlayer(p2)
	eq1 = len(reenteredMatches) == 0
	if !eq1 {
		t.Fatal("The player reentered twice")
	}

	reenteredMatches = wp.ReenterPlayer(p1)
	tournament.Update(nil)

	eq1 = ml.Matches[0] == reenteredMatches[0]
	eq2 := ml.Rounds[1].Matches[0].Slot1.player == nil
	eq3 := !ml.Rounds[1].Matches[0].HasBye()
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The second reenter did not put the tournament into the original starting state")
	}

	ml.Matches[0].StartMatch()
	ml.Matches[0].EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	withdrawnMatches = wp.WithdrawPlayer(p1)
	tournament.Update(nil)

	eq1 = ml.Rounds[1].Matches[0] == withdrawnMatches[0]
	if !eq1 {
		t.Fatal("The first seed player did not withdraw from the second round match after winning the first round")
	}

	ml.Matches[1].StartMatch()
	ml.Matches[1].EndMatch(NewScore(0, 1))
	tournament.Update(nil)

	eq1 = ml.Rounds[2].Matches[0].Slot1.player == p4
	if !eq1 {
		t.Fatal("The winner of the first round did not go through to the third round with the walkover in the second")
	}

	// Test withdrawal from unbalanced tournament
	players, err = PlayerSlice(7)
	if err != nil {
		t.Fatal(err)
	}

	entries = NewConstantRanking(players)
	tournament, _ = NewSingleElimination(entries)

	wp = tournament.WithdrawalPolicy
	ml = tournament.MatchList

	p1 = players[0]
	p2 = players[3]
	p3 := players[4]
	p4 = players[1]

	withdrawnMatches = wp.WithdrawPlayer(p1)
	tournament.Update(nil)

	eq1 = ml.Rounds[1].Matches[0] == withdrawnMatches[0]
	if !eq1 {
		t.Fatal("The first seed player withdrew from a drawn bye match.")
	}

	match1 := wp.WithdrawPlayer(p2)[0]
	tournament.Update(nil)
	match2 := wp.WithdrawPlayer(p3)[0]
	tournament.Update(nil)

	eq1 = match1 == match2
	eq2 = ml.Rounds[1].Matches[0].Slot2.IsBye()
	eq3 = ml.Rounds[2].Matches[0].Slot1.IsBye()
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The triple withdrawal did not propagate the resulting bye to the final slot")
	}

	ml.Matches[2].StartMatch()
	ml.Matches[2].EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ml.Matches[3].StartMatch()
	ml.Matches[3].EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ml.Matches[5].StartMatch()
	ml.Matches[5].EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ml.Matches[6].StartMatch()
	ml.Matches[6].EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	withdrawnMatches = wp.WithdrawPlayer(p4)
	eq1 = len(withdrawnMatches) == 0
	if !eq1 {
		t.Fatal("The tournament winner was able to withraw after the tournament was completed")
	}
}

func TestSingleEliminationEditingPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleElimination(entries)

	ml := tournament.MatchList

	ep := tournament.EditingPolicy
	ep.updateEditableMatches()

	editableMatches := ep.EditableMatches()
	eq1 := len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The list of editable matches is not empty on the fresh tournament")
	}

	ml.Matches[0].StartMatch()
	ml.Matches[0].EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ep.updateEditableMatches()

	editableMatches = ep.EditableMatches()
	eq1 = ml.Matches[0] == editableMatches[0]
	if !eq1 {
		t.Fatal("The finished match is not editable")
	}

	for _, m := range ml.Rounds[0].Matches[1:] {
		m.StartMatch()
		m.EndMatch(NewScore(1, 0))
	}
	tournament.Update(nil)
	ep.updateEditableMatches()

	editableMatches = ep.EditableMatches()
	eq1 = slices.Equal(editableMatches, ml.Rounds[0].Matches)
	if !eq1 {
		t.Fatal("The first round matches are not all editable after they ended")
	}

	ml.Rounds[1].Matches[0].StartMatch()
	tournament.Update(nil)
	ep.updateEditableMatches()

	editableMatches = ep.EditableMatches()
	eq1 = slices.Contains(editableMatches, ml.Rounds[0].Matches[0])
	eq2 := slices.Contains(editableMatches, ml.Rounds[0].Matches[1])
	if eq1 || eq2 {
		t.Fatal("The two predecessor matches of the started match are still editable")
	}
}
