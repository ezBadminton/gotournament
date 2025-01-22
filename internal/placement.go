package internal

type Placement interface {
	Slot() *Slot
	Ranking() Ranking
}

// A simple index into a Ranking
type BasePlacement struct {
	ranking Ranking
	place   int
}

// Returns the current Slot at the Placement
func (p *BasePlacement) Slot() *Slot {
	return p.ranking.At(p.place)
}

func (p *BasePlacement) Ranking() Ranking {
	return p.ranking
}

func NewPlacement(ranking Ranking, place int) *BasePlacement {
	return &BasePlacement{ranking: ranking, place: place}
}

// A simple index into a Ranking
// When blocking is true, the Slot() method always returns nil.
// Otherwise behaves like BasePlacement.
type BlockingPlacement struct {
	ranking  Ranking
	place    int
	blocking bool
}

// Returns the current Slot at the Placement
func (p *BlockingPlacement) Slot() *Slot {
	if p.blocking {
		return nil
	}
	return p.ranking.At(p.place)
}

func (p *BlockingPlacement) UnblockedSlot() *Slot {
	return p.ranking.At(p.place)
}

func (p *BlockingPlacement) Ranking() Ranking {
	return p.ranking
}

func NewBlockingPlacement(ranking Ranking, place int, blocking bool) *BlockingPlacement {
	return &BlockingPlacement{ranking: ranking, place: place, blocking: blocking}
}
