package core

// The simplest possible ranking that just provides
// a list of directly player filled slots
type ConstantRanking struct {
	BaseRanking
}

// Updates the return value of the Ranks method.
// Should be called whenever a result that influences the
// ranking becomes known.
func (r *ConstantRanking) updateRanks() {
	// No implementation because this is a constant ranking
}

// Creates a *ConstantRanking from the given slice of players.
// The ranking will provide one Slot per player while
// keeping the order.
func NewConstantRanking(players []Player) *ConstantRanking {
	slots := make([]*Slot, 0, len(players))
	for _, p := range players {
		slots = append(slots, NewPlayerSlot(p))
	}
	baseRanking := NewBaseRanking()
	baseRanking.ranks = slots
	ranking := &ConstantRanking{BaseRanking: baseRanking}

	return ranking
}
