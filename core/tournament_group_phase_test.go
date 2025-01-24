package core

import (
	"reflect"
	"testing"
)

func TestGroupSeeding(t *testing.T) {
	players, err := PlayerSlice(12)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	rankingGraph := NewRankingGraph(entries)
	tournament := newGroupPhase(entries, 4, 3, NewScore(21, 0), rankingGraph)

	groups := tournament.Groups

	groupSlots := make([][]*Slot, 0, 4)
	for _, g := range groups {
		groupSlots = append(groupSlots, g.Entries.Ranks())
	}

	eq1 := groupSlots[0][0].player == players[0]
	eq2 := groupSlots[1][0].player == players[1]
	eq3 := groupSlots[2][0].player == players[2]
	eq4 := groupSlots[3][0].player == players[3]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The first four seeds did not get put into the first slots of the four groups.")
	}

	eq1 = groupSlots[0][1].player == players[7]
	eq2 = groupSlots[1][1].player == players[6]
	eq3 = groupSlots[2][1].player == players[5]
	eq4 = groupSlots[3][1].player == players[4]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The second four seeds did not get put into the second slots of the four groups in reverse.")
	}

	eq1 = groupSlots[0][2].player == players[8]
	eq2 = groupSlots[1][2].player == players[9]
	eq3 = groupSlots[2][2].player == players[10]
	eq4 = groupSlots[3][2].player == players[11]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The thrird four seeds did not get put into the thrird slots of the four groups.")
	}

	players, err = PlayerSlice(6)
	if err != nil {
		t.Fatal(err)
	}

	entries = NewConstantRanking(players)
	rankingGraph = NewRankingGraph(entries)
	tournament = newGroupPhase(entries, 4, 4, NewScore(21, 0), rankingGraph)

	groups = tournament.Groups

	groupSlots = make([][]*Slot, 0, 4)
	for _, g := range groups {
		groupSlots = append(groupSlots, g.Entries.Ranks())
	}

	eq1 = len(groupSlots[0]) == 1
	eq2 = len(groupSlots[1]) == 1
	eq3 = len(groupSlots[2]) == 2
	eq4 = len(groupSlots[3]) == 2
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The remaining entries were not put into the higher index groups")
	}

	eq1 = groupSlots[2][1].player == players[5]
	eq2 = groupSlots[3][1].player == players[4]
	if !eq1 || !eq2 {
		t.Fatal("The two remaining entries did not go into the groups in reverse order")
	}
}

func TestGroupPhaseRanking(t *testing.T) {
	players, err := PlayerSlice(9)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	rankingGraph := NewRankingGraph(entries)
	tournament := newGroupPhase(entries, 3, 6, NewScore(1, 0), rankingGraph)

	finalRanking := tournament.FinalRanking

	ranks := finalRanking.TiedRanks()

	eq1 := len(ranks) == len(players)
	eq2 := !finalRanking.QualificationComplete
	if !eq1 || !eq2 {
		t.Fatal("The final ranking is marked as complete despite the matches not having started")
	}

	for _, m := range tournament.MatchList.Matches {
		if !m.HasBye() {
			m.StartMatch()
			m.EndMatch(NewScore(1, 0))
		}
	}

	tournament.Update(nil)
	ranks = finalRanking.TiedRanks()

	eq1 = len(ranks) == len(players)
	eq2 = !finalRanking.QualificationComplete
	if !eq1 || !eq2 {
		t.Fatal("The final ranking is marked as complete despite ties being present")
	}

	eq1 = len(finalRanking.GroupTies) == 3
	if !eq1 {
		t.Fatal("The symmetric scores did not cause a tie in the groups")
	}

	eq1 = len(finalRanking.CrossGroupTies()) == 0
	eq2 = finalRanking.RequiredUntiedRanks == 6
	if !eq1 || !eq2 {
		t.Fatal("The cross group ties are not empty")
	}

	for _, m := range tournament.MatchList.Matches[6:9] {
		m.Score = NewScore(0, 1)
	}
	tournament.Update(nil)

	eq1 = len(finalRanking.GroupTies) == 0
	eq2 = len(finalRanking.CrossGroupTies()) == 0
	eq3 := finalRanking.QualificationComplete
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The final ranking is not free of ties despite the different player scores")
	}

	ranks = finalRanking.TiedRanks()
	eq1 = ranks[0][0].player == players[0]
	eq2 = ranks[1][0].player == players[1]
	eq3 = ranks[2][0].player == players[2]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The group winners do not occupy the top 3 final ranks")
	}

	eq1 = ranks[3][0].player == players[5]
	eq2 = ranks[4][0].player == players[4]
	eq3 = ranks[5][0].player == players[3]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The group 2nd places do not occupy the ranks 4-6")
	}

	eq1 = ranks[6][0].player == players[6]
	eq2 = ranks[7][0].player == players[7]
	eq3 = ranks[8][0].player == players[8]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The group 3rd places do not occupy the ranks 7-9")
	}
}

func TestCrossGroupTies(t *testing.T) {
	players, err := PlayerSlice(6)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	rankingGraph := NewRankingGraph(entries)
	tournament := newGroupPhase(entries, 3, 5, NewScore(1, 0), rankingGraph)

	ml := tournament.MatchList
	finalRanking := tournament.FinalRanking

	for _, m := range ml.Matches {
		m.StartMatch()
		m.EndMatch(NewScore(2, 0))
	}
	tournament.Update(nil)

	crossTies := finalRanking.CrossGroupTies()
	eq1 := len(finalRanking.GroupTies) == 0 && len(crossTies) == 1
	if !eq1 {
		t.Fatal("The equal scores did not cause a tie for the contested qualifications")
	}

	eq1 = len(crossTies[0]) == 3
	eq2 := crossTies[0][0].player == players[5]
	eq3 := crossTies[0][1].player == players[4]
	eq4 := crossTies[0][2].player == players[3]
	if !eq1 || !eq2 || !eq3 || !eq4 {
		t.Fatal("The cross tie does not contain the three 2nd placed group slots")
	}

	ml.Matches[1].Score = NewScore(3, 0)
	tournament.Update(nil)
	crossTies = finalRanking.CrossGroupTies()
	ranks := finalRanking.TiedRanks()

	eq1 = len(crossTies) == 0
	eq2 = ranks[3][0].player == players[5]
	eq3 = ranks[4][0].player == players[3]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("There is a blocking cross group tie despite the tie being the same size as the number of contested ranks")
	}

	ml.Matches[1].Score = NewScore(2, 1)
	tournament.Update(nil)
	crossTies = finalRanking.CrossGroupTies()

	eq1 = len(crossTies[0]) == 2
	eq2 = crossTies[0][0].player == players[5]
	eq3 = crossTies[0][1].player == players[3]
	if !eq1 || !eq2 || !eq3 {
		t.Fatal("The cross tie does not contain the two 2nd placed group slots who had fewer points")
	}

	finalRanking.AddTieBreaker(NewConstantRanking([]Player{players[5], players[3]}))
	finalRanking.updateRanks()
	crossTies = finalRanking.CrossGroupTies()

	eq1 = len(crossTies) == 0
	if !eq1 {
		t.Fatal("The tie breaker did not remove the cross group tie")
	}
}

func TestGroupPhaseWithdrawal(t *testing.T) {
	players, err := PlayerSlice(6)
	if err != nil {
		t.Fatal(err)
	}

	entries := NewConstantRanking(players)
	rankingGraph := NewRankingGraph(entries)
	walkoverScore := NewScore(42, 0)
	tournament := newGroupPhase(entries, 2, 6, walkoverScore, rankingGraph)

	wp := tournament.WithdrawalPolicy
	groupRankings := make([]*MatchMetricRanking, 0, 2)
	for _, g := range tournament.Groups {
		groupRankings = append(groupRankings, g.FinalRanking)
	}
	finalRanking := tournament.FinalRanking
	ml := tournament.MatchList

	matches := ml.Matches
	matches = []*Match{matches[2], matches[3], matches[5], matches[9]}

	withdrawnMatches := wp.WithdrawPlayer(players[0])

	eq1 := len(withdrawnMatches) == 2
	if !eq1 {
		t.Fatal("The withdrawn matches of the first seed player are not correct")
	}

	tournament.Update(nil)

	group1Metrics := groupRankings[0].Metrics
	p4Metrics := group1Metrics[players[3]]
	p5Metrics := group1Metrics[players[4]]

	trueMetrics := &MatchMetrics{
		NumMatches:      1,
		Wins:            1,
		NumSets:         1,
		SetWins:         1,
		PointWins:       walkoverScore.Points1()[0],
		SetDifference:   1,
		PointDifference: walkoverScore.Points1()[0],
	}

	eq1 = reflect.DeepEqual(p4Metrics, trueMetrics)
	eq2 := reflect.DeepEqual(p5Metrics, trueMetrics)
	if !eq1 || !eq2 {
		t.Fatal("The oppponents of the withdrawn player did not receive the walkover score")
	}

	for i, m := range matches {
		m.StartMatch()
		m.EndMatch(NewScore(i+1, 0))
	}
	tournament.Update(nil)

	ranks := finalRanking.TiedRanks()

	eq1 = len(ranks) == 6
	eq2 = ranks[0][0].player == players[3]
	eq3 := ranks[1][0].player == players[5]
	eq4 := ranks[2][0].player == players[4]
	eq5 := ranks[3][0].player == players[1]
	eq6 := ranks[5][0].player == players[2]
	if !eq1 || !eq2 || !eq3 || !eq4 || !eq5 || !eq6 {
		t.Fatal("The final ranking has an unexpected order")
	}

	eq1 = ranks[4][0].player == nil && ranks[4][0].IsBye()
	if !eq1 {
		t.Fatal("The withdrawn player was not excluded from the final rankings")
	}
}
