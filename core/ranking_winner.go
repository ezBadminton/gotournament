package core

import "slices"

// A WinnerRanking ranks the two participants of a
// Match into winner and loser.
type WinnerRanking struct {
	BaseRanking

	Match *Match
}

// Updates the return value of the GetRanks() method.
// Should be called whenever a result that influences the
// ranking becomes known.
func (r *WinnerRanking) updateRanks() {
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

	overrideDrawnBye := loser.Bye() != nil && loser.Bye().Drawn
	if overrideDrawnBye {
		loser = NewByeSlot(false)
	}

	blockWithdrawnPlayer := slices.Contains(r.Match.WithdrawnSlots(), loser)
	if blockWithdrawnPlayer {
		loser = NewByeSlot(false)
	}

	slots := []*Slot{winner, loser}

	r.Ranks = slots
}

// Creates a new WinnerRanking
func NewWinnerRanking(match *Match) *WinnerRanking {
	baseRanking := NewBaseRanking()
	ranking := &WinnerRanking{Match: match, BaseRanking: baseRanking}
	return ranking
}

// Adds this WinnerRanking as a dependant to the source rankings of the match's slots
// if they are both also WinnerRankings and the matches of those are keys in the allowedLinks
func (r *WinnerRanking) LinkRankingGraph(rankingGraph *RankingGraph, allowedLinks map[*Match]*WinnerRanking) {
	rankingGraph.AddVertex(r)

	placement1 := r.Match.Slot1.Placement()
	placement2 := r.Match.Slot2.Placement()

	if placement1 == nil || placement2 == nil {
		return
	}

	ranking1, ok1 := placement1.Ranking().(*WinnerRanking)
	ranking2, ok2 := placement2.Ranking().(*WinnerRanking)

	if !ok1 || !ok2 {
		return
	}

	_, filter1 := allowedLinks[ranking1.Match]

	if filter1 {
		rankingGraph.AddEdge(ranking1, r)
	}

	if ranking1 == ranking2 {
		return
	}

	_, filter2 := allowedLinks[ranking2.Match]

	if filter2 {
		rankingGraph.AddEdge(ranking2, r)
	}
}
