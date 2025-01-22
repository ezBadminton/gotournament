package internal

import (
	"reflect"
	"testing"
)

func TestRoundRobinCircleIndex(t *testing.T) {
	length := 14

	r1 := roundRobinCircleIndex(0, length, 0)
	r2 := roundRobinCircleIndex(0, length, 7)
	if r1 != 0 || r2 != 0 {
		t.Fatal("Index 0 did not stay fixed")
	}

	r1 = roundRobinCircleIndex(1, length, 0)
	r2 = roundRobinCircleIndex(5, length, 0)
	if r1 != 1 || r2 != 5 {
		t.Fatal("First round rotation did not preserve the original index")
	}

	r1 = roundRobinCircleIndex(1, length, 1)
	r2 = roundRobinCircleIndex(length-1, length, 1)
	if r1 != length-1 || r2 != length-2 {
		t.Fatal("Second round index was not rotated by one")
	}

	r1 = roundRobinCircleIndex(1, length, length-2)
	if r1 != 2 {
		t.Fatal("The last round did not rotate index 1 to index 2")
	}
}

func TestRoundRobin(t *testing.T) {
	players, err := PlayerSlice(4)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)

	tournament := NewRoundRobin(entries, 1, NewScore(21, 0))

	ml := tournament.MatchList
	finalRanking := tournament.FinalRanking

	eq1 := len(ml.Matches) == 6
	if !eq1 {
		t.Fatal("The 4-player round robin has an unexpected amount of matches")
	}

	ranks := finalRanking.TiedRanks()
	eq1 = len(ranks) == 1
	eq2 := len(ranks[0]) == 4
	if !eq1 || !eq2 {
		t.Fatal("The initial final ranking is not tied between all players")
	}

	p1 := players[0]
	p2 := players[1]
	p3 := players[2]
	p4 := players[3]

	// match 0 is p1 vs p4
	ml.Matches[0].StartMatch()
	ml.Matches[0].EndMatch(NewScore(10, 5))
	tournament.Update(nil)

	stats1 := finalRanking.Metrics[p1]
	stats4 := finalRanking.Metrics[p4]

	trueStats1 := &MatchMetrics{
		NumMatches:      1,
		Wins:            1,
		Losses:          0,
		NumSets:         1,
		SetWins:         1,
		SetLosses:       0,
		PointWins:       10,
		PointLosses:     5,
		SetDifference:   1,
		PointDifference: 5,
	}

	trueStats4 := &MatchMetrics{
		NumMatches:      1,
		Wins:            0,
		Losses:          1,
		NumSets:         1,
		SetWins:         0,
		SetLosses:       1,
		PointWins:       5,
		PointLosses:     10,
		SetDifference:   -1,
		PointDifference: -5,
	}

	eq1 = reflect.DeepEqual(stats1, trueStats1)
	eq2 = reflect.DeepEqual(stats4, trueStats4)
	if !eq1 || !eq2 {
		t.Fatal("The match metrics did not update correctly according to the frist match's result")
	}

	ranks = finalRanking.TiedRanks()
	eq1 = len(ranks) == 3
	eq2 = len(ranks[0]) == 1 && ranks[0][0].Player() == p1
	eq3 := len(ranks[1]) == 2 && ranks[1][0].Player() == p2 && ranks[1][1].Player() == p3
	eq4 := len(ranks[2]) == 1 && ranks[2][0].Player() == p4
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The final ranks did not update correctly according to the first match's result")
	}
}

func TestRoundRobinByes(t *testing.T) {
	players, err := PlayerSlice(3)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewRoundRobin(entries, 1, NewScore(21, 0))

	ml := tournament.MatchList
	finalRanking := tournament.FinalRanking

	eq1 := len(ml.Matches) == 6
	if !eq1 {
		t.Fatal("The 3-player round robin has an unexpected amount of matches")
	}

	eq1 = ml.Matches[0].HasDrawnBye()
	eq2 := ml.Matches[3].HasDrawnBye()
	eq3 := ml.Matches[5].HasDrawnBye()
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The expected drawn bye matches are not present")
	}

	ranks := finalRanking.TiedRanks()
	eq1 = len(ranks) == 1
	eq2 = len(ranks[0]) == 3
	if !eq1 || !eq2 {
		t.Fatal("The initial final ranking is not tied between all players")
	}

	stats := finalRanking.Metrics[players[0]]
	eq1 = stats.NumMatches == 0
	if !eq1 {
		t.Fatal("The match metrics do not have 0 number of matches")
	}
}

func TestRoundRobinWithdrawal(t *testing.T) {
	players, err := PlayerSlice(3)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewRoundRobin(entries, 1, NewScore(21, 0))

	finalRanking := tournament.FinalRanking

	wp := tournament.WithdrawalPolicy

	p1 := players[0]
	p2 := players[1]
	p3 := players[2]

	withdrawnMatches := wp.WithdrawPlayer(p1)
	eq1 := len(withdrawnMatches) == 2
	if !eq1 {
		t.Fatal("The withdrawn matches are not as expected")
	}

	tournament.Update(nil)
	ranks := finalRanking.TiedRanks()
	eq1 = len(ranks) == 2
	eq2 := len(ranks[0]) == 2 && ranks[0][0].Player() == p2 && ranks[0][1].Player() == p3
	eq3 := len(ranks[1]) == 1 && ranks[1][0].Player() == p1
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The final ranking did not put the non-withdrawn players to the top and the withdrawn player to the bottom")
	}

	stats1 := finalRanking.Metrics[p1]
	stats2 := finalRanking.Metrics[p2]
	stats3 := finalRanking.Metrics[p3]

	eq1 = stats1.NumMatches == 2 && stats1.Losses == 2
	eq2 = stats1.PointDifference == -42
	eq3 = reflect.DeepEqual(stats2, stats3)
	eq4 := stats2.NumMatches == 1 && stats2.Wins == 1
	eq5 := stats2.PointDifference == 21
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 {
		t.Fatal("The match metrics did not change correctly according to the walkover match results")
	}

	reenteredMatches := wp.ReenterPlayer(p1)
	eq1 = reflect.DeepEqual(withdrawnMatches, reenteredMatches)
	if !eq1 {
		t.Fatal("The withdrawn player did not reenter into the withdrawn matches")
	}

	tournament.Update(nil)
	ranks = finalRanking.TiedRanks()

	eq1 = len(ranks) == 1
	eq2 = len(ranks[0]) == 3
	if !eq1 || !eq2 {
		t.Fatal("The final ranking did not reset to the initial tied state after the reentering")
	}
}

func TestRoundRobinEditing(t *testing.T) {
	players, err := PlayerSlice(3)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	tournament := NewRoundRobin(entries, 1, NewScore(21, 0))

	ml := tournament.MatchList
	ep := tournament.EditingPolicy

	editableMatches := ep.EditableMatches()
	eq1 := len(editableMatches) == 0
	if !eq1 {
		t.Fatal("There are editable matches despite no matches being completed")
	}

	ml.Matches[1].StartMatch()
	ml.Matches[1].EndMatch(NewScore(1, 0))

	ep.Update()
	editableMatches = ep.EditableMatches()

	eq1 = len(editableMatches) == 1 && editableMatches[0] == ml.Matches[1]
	if !eq1 {
		t.Fatal("The completed match is not editable")
	}
}

func TestRoundRobinTies(t *testing.T) {
	players, err := PlayerSlice(3)
	if err != nil {
		t.Fatal(err)
	}

	p1 := players[0]
	p2 := players[1]
	p3 := players[2]

	entries := NewConstantRanking(players)
	tournament := NewRoundRobin(entries, 1, NewScore(1, 0))

	// 2-1 2-0 0-2
	matches := tournament.MatchList.Matches
	finalRanking := tournament.FinalRanking

	for _, m := range []*Match{matches[1], matches[2], matches[4]} {
		m.StartMatch()
		m.EndMatch(NewScore(1, 0))
	}
	tournament.Update(nil)

	ranks := finalRanking.TiedRanks()
	eq1 := len(ranks) == 1
	eq2 := len(ranks[0]) == 3
	if !eq1 || !eq2 {
		t.Fatal("The final ranking is not tied after every player had one win in the round robin")
	}

	finalRanking.AddTieBreaker(entries)
	finalRanking.UpdateRanks()
	ranks = finalRanking.TiedRanks()

	eq1 = len(ranks) == 3
	eq2 = len(ranks[0]) == 1 && ranks[0][0].Player() == p1
	eq3 := len(ranks[1]) == 1 && ranks[1][0].Player() == p2
	eq4 := len(ranks[2]) == 1 && ranks[2][0].Player() == p3
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The tie breaker did not break the tie in the right order")
	}

	tournament = NewRoundRobin(entries, 2, NewScore(1, 0))
	matches = tournament.MatchList.Matches
	finalRanking = tournament.FinalRanking

	// p1 and p2 win against p3 and p1 & p2
	// each win one of their matches against each other
	matches[1].StartMatch()
	matches[1].EndMatch(NewScore(1, 0))
	matches[2].StartMatch()
	matches[2].EndMatch(NewScore(0, 1))
	matches[4].StartMatch()
	matches[4].EndMatch(NewScore(1, 0))
	matches[7].StartMatch()
	matches[7].EndMatch(NewScore(0, 1))
	matches[8].StartMatch()
	matches[8].EndMatch(NewScore(1, 0))
	matches[10].StartMatch()
	matches[10].EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	ranks = finalRanking.TiedRanks()

	eq1 = len(ranks) == 2
	eq2 = len(ranks[0]) == 2
	eq3 = ranks[0][0].Player() == p1 && ranks[0][1].Player() == p2
	eq4 = len(ranks[1]) == 1 && ranks[1][0].Player() == p3
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The two winning players are not tied for first and the losing player is not at the bottom of the ranks")
	}

	// Give more points to p2
	matches[10].Score = NewScore(2, 0)
	tournament.Update(nil)

	ranks = finalRanking.TiedRanks()
	eq1 = len(ranks) == 3
	eq2 = len(ranks[0]) == 1 && ranks[0][0].Player() == p2
	eq3 = len(ranks[1]) == 1 && ranks[1][0].Player() == p1
	eq4 = len(ranks[2]) == 1 && ranks[2][0].Player() == p3
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The additional point did not put the player to the top of the ranks")
	}

	// Give more sets to p1
	matches[4].Score.(*TestScore).numSets = 2
	tournament.Update(nil)

	ranks = finalRanking.TiedRanks()
	eq1 = len(ranks) == 3
	eq2 = len(ranks[0]) == 1 && ranks[0][0].Player() == p1
	eq3 = len(ranks[1]) == 1 && ranks[1][0].Player() == p2
	eq4 = len(ranks[2]) == 1 && ranks[2][0].Player() == p3
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The additional set did not put the player to the top of the ranks")
	}

	players, err = PlayerSlice(4)
	if err != nil {
		t.Fatal(err)
	}
	entries = NewConstantRanking(players)
	tournament = NewRoundRobin(entries, 2, NewScore(1, 0))
	matches = tournament.MatchList.Matches
	finalRanking = tournament.FinalRanking

	p1 = players[0]
	p2 = players[1]

	// Have p1 and p2 with one win and one loss each
	// p1 wins the direct encounter
	matches[0].StartMatch()
	matches[0].EndMatch(NewScore(0, 1))
	matches[1].StartMatch()
	matches[1].EndMatch(NewScore(1, 0))
	matches[4].StartMatch()
	matches[4].EndMatch(NewScore(1, 0))
	tournament.Update(nil)

	ranks = finalRanking.TiedRanks()

	eq1 = reflect.DeepEqual(finalRanking.Metrics[p1], finalRanking.Metrics[p2])
	eq2 = len(ranks) == 4
	eq3 = ranks[1][0].Player() == p1
	eq4 = ranks[2][0].Player() == p2
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The players with equal match metrics were not ranked by the result of their direct encounter")
	}
}
