package internal

// A [BalancedRanking] wraps another ranking and pads
// it with bye slots such that the number of ranks
// becomes a power of 2, facilitating a balanced
// elimination tournament tree.
type BalancedRanking struct {
	BaseRanking
	sourceRanking Ranking
}

// Updates the return value of the GetRanks() method.
// Should be called whenever a result that influences the
// ranking becomes known.
func (r *BalancedRanking) updateRanks() {
	sourceRanks := r.sourceRanking.GetRanks()
	sourceNumSlots := len(sourceRanks)
	numSlots := nextPowerOfTwo(sourceNumSlots)
	padding := numSlots - sourceNumSlots

	if padding == 0 {
		r.Ranks = sourceRanks
		return
	}

	slots := make([]*Slot, 0, numSlots)
	slots = append(slots, sourceRanks...)
	for range padding {
		slots = append(slots, NewByeSlot(true))
	}
	r.Ranks = slots
}

func NewBalancedRanking(source Ranking, rankingGraph *RankingGraph) *BalancedRanking {
	baseRanking := NewBaseRanking()
	ranking := &BalancedRanking{sourceRanking: source, BaseRanking: baseRanking}
	ranking.updateRanks()

	rankingGraph.AddVertex(ranking)
	rankingGraph.AddEdge(source, ranking)

	return ranking
}

// Returns the power of two that is immediately bigger than
// or equal to from
func nextPowerOfTwo(from int) int {
	powerOfTwo := 1
	for powerOfTwo < from {
		powerOfTwo *= 2
	}
	return powerOfTwo
}
