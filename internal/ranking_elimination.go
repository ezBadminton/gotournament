package internal

import "slices"

// The EliminationRanking ranks the players in an
// elimination tournament according to how far they
// reached. It is a tieable ranking with the players
// who lost out in the same round being tied on the
// same rank.
type EliminationRanking struct {
	BaseTieableRanking

	MatchList *MatchList
	Entries   Ranking
}

// Updates the return value of the GetRanks() method.
// Should be called whenever a result that influences the
// ranking becomes known.
func (r *EliminationRanking) UpdateRanks() {
	rounds := make([]*Round, 0, len(r.MatchList.Rounds))

	for _, r := range r.MatchList.Rounds {
		if len(r.NestedRounds) == 0 {
			rounds = append(rounds, r)
		} else {
			nested := make([]*Round, 0, len(r.NestedRounds))
			for _, r := range slices.Backward(r.NestedRounds) {
				nested = append(nested, r)
			}
			rounds = append(rounds, nested...)
		}
	}

	numRounds := len(rounds)

	ranks := make([][]*Slot, 0, 2*numRounds)

	for _, r := range slices.Backward(rounds) {
		roundRanks := rankRound(r)
		if len(roundRanks) > 0 {
			ranks = append(ranks, roundRanks...)
		}
	}

	ranks = append(ranks, r.Entries.GetRanks())

	ranks = RemoveDoubleRanks(ranks)

	r.ProcessUpdate(ranks)
}

func rankRound(round *Round) [][]*Slot {
	size := len(round.Matches)

	winners := make([]*Slot, 0, size)
	losers := make([]*Slot, 0, size)

	for _, m := range round.Matches {
		matchWinners, matchLosers := rankMatch(m)
		if matchWinners != nil {
			winners = append(winners, matchWinners)
		}
		losers = append(losers, matchLosers...)
	}

	ranks := make([][]*Slot, 0, 2)
	if len(winners) > 0 {
		ranks = append(ranks, winners)
	}
	if len(losers) > 0 {
		ranks = append(ranks, losers)
	}

	return ranks
}

func rankMatch(match *Match) (*Slot, []*Slot) {
	winner, _ := match.GetWinner()
	losers := make([]*Slot, 0, 2)

	if winner == nil {
		for s := range match.Slots {
			if s.Player() != nil {
				losers = append(losers, s)
			}
		}
	} else {
		loser := match.OtherSlot(winner)
		if loser != nil && loser.Player() != nil {
			losers = append(losers, loser)
		}
	}

	return winner, losers
}

func RemoveDoubleRanks(ranks [][]*Slot) [][]*Slot {
	found := make(map[Player]struct{})
	cleanedRanks := make([][]*Slot, 0, len(ranks))

	for _, r := range ranks {
		cleanedRank := make([]*Slot, 0, len(r))
		for _, s := range r {
			player := s.Player()
			_, ok := found[player]
			if !ok {
				cleanedRank = append(cleanedRank, s)
				found[player] = struct{}{}
			}
		}
		if len(cleanedRank) > 0 {
			cleanedRanks = append(cleanedRanks, cleanedRank)
		}
	}
	return cleanedRanks
}

func NewEliminationRanking(
	matchList *MatchList,
	entries Ranking,
	finalsRankings []Ranking,
	rankingGraph *RankingGraph,
) *EliminationRanking {
	baseRanking := NewBaseTieableRanking(0)
	ranking := &EliminationRanking{
		BaseTieableRanking: baseRanking,
		MatchList:          matchList,
		Entries:            entries,
	}
	rankingGraph.AddVertex(ranking)
	for _, r := range finalsRankings {
		rankingGraph.AddEdge(r, ranking)
	}
	return ranking
}
