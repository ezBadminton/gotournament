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
	if ranking == nil {
		panic("Passed nil ranking to a placement")
	}
	return &BasePlacement{ranking: ranking, place: place}
}
