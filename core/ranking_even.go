package core

// An even Ranking appends a bye slot to its ranks
// but only if the source ranking has an uneven number
// of slots so it is guaranteed to have an even number.
type EvenRanking struct {
	BaseRanking
	sourceRanking Ranking
}

func (r *EvenRanking) updateRanks() {
	sourceRanks := r.sourceRanking.Ranks()

	if len(sourceRanks)%2 != 0 {
		byeSlot := NewByeSlot(true)
		sourceRanks = append(sourceRanks, byeSlot)
	}

	r.ranks = sourceRanks
}

func NewEvenRanking(source Ranking, rankingGraph *RankingGraph) *EvenRanking {
	baseRanking := NewBaseRanking()
	ranking := &EvenRanking{BaseRanking: baseRanking, sourceRanking: source}
	ranking.updateRanks()

	rankingGraph.AddVertex(ranking)
	rankingGraph.AddEdge(source, ranking)

	return ranking
}
