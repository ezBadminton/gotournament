package internal

import (
	"math"
	"reflect"
	"slices"
	"testing"
)

func TestFinalsToBrackets(t *testing.T) {
	eq1 := finalsToBrackets(0) == 0
	eq2 := finalsToBrackets(1) == 1
	if !eq1 || !eq2 {
		t.Fatal("The amount of brackets for the low end number of finals is incorrect")
	}

	for i := 2; i < 300; i += 1 {
		b := finalsToBrackets(i)
		eq1 = int(math.Floor(math.Log2(float64(i))+1.0)) == b
		if !eq1 {
			t.Fatalf("The number of brackets for %v finals was unexpected: %v", i, b)
		}
	}
}

func TestPlacesToFinals(t *testing.T) {
	for i := range 500 {
		eq1 := placesToFinals(i) == int(math.Ceil(float64(i)/2.0))
		if !eq1 {
			t.Fatal("The placesToFinals function returned an unexpected result")
		}
	}
}

func TestConsolationBrackets(t *testing.T) {
	players, err := PlayerSlice(16)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleEliminationWithConsolation(entries, 0, 16)

	brackets := tournament.Brackets

	eq1 := len(brackets) == 8
	if !eq1 {
		t.Fatal("The tournament with 16 places to play out does not have the expected amount of brackets")
	}
	eq1 = len(tournament.MatchList.Rounds[3].NestedRounds) == 8
	if !eq1 {
		t.Fatal("The tournament does not have 8 finals")
	}

	mainBracket := brackets[0]

	eq1 = len(mainBracket.MatchList.Rounds) == len(mainBracket.Consolations[0].MatchList.Rounds)+1
	eq2 := len(mainBracket.MatchList.Rounds) == len(mainBracket.Consolations[1].MatchList.Rounds)+2
	eq3 := len(mainBracket.MatchList.Rounds) == len(mainBracket.Consolations[2].MatchList.Rounds)+3
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The consolation brackets do not have a descending amount of rounds")
	}

	tournament, _ = NewSingleEliminationWithConsolation(entries, 0, 15)
	brackets = tournament.Brackets

	eq1 = len(brackets) == 8
	if !eq1 {
		t.Fatal("The tournament with 15 places to play out does not have the expected amount of brackets")
	}

	tournament, _ = NewSingleEliminationWithConsolation(entries, 0, 14)
	brackets = tournament.Brackets

	eq1 = len(brackets) == 7
	if !eq1 {
		t.Fatal("The tournament with 14 places to play out does not have the expected amount of brackets")
	}
	eq1 = len(tournament.MatchList.Rounds[3].NestedRounds) == 7
	if !eq1 {
		t.Fatal("The tournament does not have 7 finals")
	}

	tournament, _ = NewSingleEliminationWithConsolation(entries, 1, 0)
	brackets = tournament.Brackets

	eq1 = len(brackets) == 4
	if !eq1 {
		t.Fatal("The tournament with 1 consolation round does not have the expected amount of brackets")
	}

	tournament, _ = NewSingleEliminationWithConsolation(entries, 1, 8)
	brackets = tournament.Brackets

	eq1 = len(brackets) == 5
	if !eq1 {
		t.Fatal("The tournament with 1 consolation round and 8 places to play out does not have the expected amount of brackets")
	}

	players, err = PlayerSlice(6)
	if err != nil {
		t.Fatal(err)
	}

	entries = NewConstantRanking(players)
	tournament, _ = NewSingleEliminationWithConsolation(entries, 0, 8)
	brackets = tournament.Brackets

	eq1 = len(brackets) == 3
	if !eq1 {
		t.Fatal("The tournament with two drawn byes has an all-bye bracket")
	}
}

func TestConsolationGraphs(t *testing.T) {
	players, err := PlayerSlice(16)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleEliminationWithConsolation(entries, 0, 16)

	brackets := tournament.Brackets
	mainBracket := brackets[0]

	rankingGraph := mainBracket.RankingGraph

	firstConsolation := mainBracket.Consolations[0]
	secondConsolation := firstConsolation.Consolations[0]

	firstWinnerRanking := mainBracket.WinnerRankings[mainBracket.MatchList.Matches[0]]
	firstConsolationWinnerRanking := firstConsolation.WinnerRankings[firstConsolation.MatchList.Matches[0]]

	firstMatchDependants := rankingGraph.GetDependants(firstWinnerRanking)
	firstConsolationDependants := rankingGraph.GetDependants(firstConsolationWinnerRanking)

	eq1 := slices.Contains(firstMatchDependants, firstConsolation.Entries)
	eq2 := slices.Contains(firstConsolationDependants, secondConsolation.Entries)
	eq3 := len(firstMatchDependants) == 2 && len(firstConsolationDependants) == 2
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The brackets are not properly connected in the ranking graph. The losers do not go to the entries ranking of the consolations.")
	}

	eliminationGraph := tournament.EliminationGraph

	firstMatch := mainBracket.MatchList.Matches[0]
	secondRoundMatch := mainBracket.MatchList.Rounds[1].Matches[0]
	firstConsolationMatch := firstConsolation.MatchList.Matches[0]

	firstMatchQualifications := eliminationGraph.GetDependants(firstMatch)

	eq1 = slices.Contains(firstMatchQualifications, secondRoundMatch)
	eq2 = slices.Contains(firstMatchQualifications, firstConsolationMatch)
	eq3 = len(firstMatchQualifications) == 2
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The matches are not properly connected in the qualification graph. The winner/loser no not go to the next round/consolation")
	}

}

func TestConsolationRanking(t *testing.T) {
	players, err := PlayerSlice(4)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleEliminationWithConsolation(entries, 0, 4)

	matchList := tournament.MatchList
	finalRanking := tournament.FinalRanking

	semi1 := matchList.Matches[0]
	semi2 := matchList.Matches[1]
	final := matchList.Matches[2]
	loserFinal := matchList.Matches[3]

	semi1.StartMatch()
	semi2.StartMatch()
	loserFinal.StartMatch()
	semi1.EndMatch(NewScore(1, 0))
	semi2.EndMatch(NewScore(1, 0))
	loserFinal.EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ranks := finalRanking.TiedRanks()

	eq1 := len(ranks) == 3
	eq2 := len(ranks[0]) == 2 && len(ranks[1]) == 1 && len(ranks[2]) == 1
	eq3 := ranks[0][0].Player() == players[0] && ranks[0][1].Player() == players[1]
	eq4 := ranks[1][0].Player() == players[3] && ranks[2][0].Player() == players[2]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The finalists are not tied for first and/or the players in the loser final are not ranked correctly")
	}

	final.StartMatch()
	final.EndMatch(NewScore(1, 0))
	tournament.Update(nil)
	ranks = finalRanking.TiedRanks()

	eq1 = len(ranks) == 4
	eq2 = ranks[0][0].Player() == players[0]
	eq3 = ranks[1][0].Player() == players[1]
	eq4 = ranks[2][0].Player() == players[3]
	eq5 := ranks[3][0].Player() == players[2]
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 {
		t.Fatal("The final ranking of the tournament is unexpected")
	}
}

func TestConsolationWithdrawal(t *testing.T) {
	players, err := PlayerSlice(7)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleEliminationWithConsolation(entries, 0, 8)

	ml := tournament.MatchList
	brackets := tournament.Brackets

	semi1 := ml.Rounds[1].Matches[0]

	eq1 := semi1.Slot1.Player() == players[0]
	if !eq1 {
		t.Fatal("The first seed player did not advance to the next round")
	}

	firstConsolationMatch := brackets[1].MatchList.Matches[0]
	matchFor7th := brackets[2].MatchList.Matches[0]

	eq1 = firstConsolationMatch.Slot1.IsBye()
	eq2 := matchFor7th.Slot1.IsBye()
	if !eq1 || !eq2 {
		t.Fatal("The drawn bye did not propagate into the consolation brackets")
	}

	wp := tournament.WithdrawalPolicy
	withdrawnMatches := wp.WithdrawPlayer(players[0])

	eq1 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == semi1
	if !eq1 {
		t.Fatal("The first seed player did not withdraw from their current match")
	}

	tournament.Update(nil)
	matchFor3rd := brackets[3].MatchList.Matches[0]

	eq1 = matchFor3rd.Slot1.IsBye()
	if !eq1 {
		t.Fatal("The withdrawal did not cause the match for 3rd place to get a bye")
	}

	quarter2 := ml.Matches[1]
	withdrawnMatches = wp.WithdrawPlayer(players[3])
	eq1 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == quarter2
	withdrawnMatches = wp.WithdrawPlayer(players[4])
	eq2 = len(withdrawnMatches) == 1 && withdrawnMatches[0] == quarter2
	if !eq1 || !eq2 {
		t.Fatal("The withdrawal of the two players in the quarter-final did not return the quarter final match")
	}

	tournament.Update(nil)
	final := brackets[0].MatchList.Matches[6]
	matchFor5th := brackets[1].MatchList.Matches[2]
	eq1 = final.Slot1.IsBye()
	eq2 = matchFor3rd.Slot1.IsBye()
	eq3 := firstConsolationMatch.Slot1.IsBye()
	eq4 := matchFor7th.Slot1.IsBye()
	eq5 := matchFor5th.Slot1.IsBye()
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 {
		t.Fatal("The byes of the withdrawn players did not propagate to the consolation brackets correctly")
	}
}

func TestConsolationEditingPolicy(t *testing.T) {
	players, err := PlayerSlice(8)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament, _ := NewSingleEliminationWithConsolation(entries, 0, 8)

	ml := tournament.MatchList
	ep := tournament.EditingPolicy

	editableMatches := ep.EditableMatches()

	eq1 := len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The slice of editable matches is not empty at the start of the tournament")
	}

	quarter1 := ml.Matches[0]
	quarter2 := ml.Matches[1]
	semi1 := ml.Rounds[1].Matches[0]

	quarter1.StartMatch()
	quarter2.StartMatch()
	quarter1.EndMatch(NewScore(1, 0))
	quarter2.EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	editableMatches = ep.EditableMatches()

	eq1 = len(editableMatches) == 2 && reflect.DeepEqual([]*Match{quarter1, quarter2}, editableMatches)
	if !eq1 {
		t.Fatal("The two completed matches did not become editable")
	}

	semi1.StartMatch()
	ep.updateEditableMatches()

	editableMatches = ep.EditableMatches()

	eq1 = len(editableMatches) == 0
	if !eq1 {
		t.Fatal("The start of the semi final did not make the quarter finals undeditable")
	}

	semi1.EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	editableMatches = ep.EditableMatches()

	eq1 = len(editableMatches) == 1 && editableMatches[0] == semi1
	if !eq1 {
		t.Fatal("The completion of the semi final did not make it editable")
	}
}
