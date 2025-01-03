package internal

// A WinnerRanking ranks the two participants of a
// Match into winner and loser.
type WinnerRanking struct {
	BaseRanking

	Match *Match
}

// Updates the return value of the GetRanks() method.
// Should be called whenever a result that influences the
// ranking becomes known.
func (r *WinnerRanking) UpdateRanks() {
	winner, err := r.Match.GetWinner()
	if err == ErrBothBye || err == ErrBothWalkover || err == ErrByeAndWalkover {
		byeSlot := NewByeSlot(false)
		r.Ranks = []*Slot{byeSlot, byeSlot}
		return
	}
	if winner == nil {
		r.Ranks = make([]*Slot, 0)
		return
	}

	loser := r.Match.OtherSlot(winner)
	if loser.Bye() != nil && loser.Bye().Drawn {
		loser = NewByeSlot(false)
	}

	slots := []*Slot{winner, loser}

	r.Ranks = slots
}

// Creates a new WinnerRanking
// and adds the new ranking as a dependant to the source
// rankings of the match slots
func NewWinnerRanking(match *Match, rankingGraph *RankingGraph) *WinnerRanking {
	baseRanking := NewBaseRanking()
	ranking := &WinnerRanking{Match: match, BaseRanking: baseRanking}
	linkRankingGraph(match, ranking, rankingGraph)
	return ranking
}

func linkRankingGraph(match *Match, ranking Ranking, rankingGraph *RankingGraph) {
	rankingGraph.AddVertex(ranking)

	placement1 := match.Slot1.Placement()
	placement2 := match.Slot2.Placement()

	if placement1 == nil || placement2 == nil {
		return
	}

	rankingGraph.AddEdge(placement1.Ranking(), ranking)

	if placement1.Ranking() == placement2.Ranking() {
		return
	}

	rankingGraph.AddEdge(placement2.Ranking(), ranking)
}
