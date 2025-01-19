package internal

import (
	"cmp"
	"slices"
	"strings"
)

// A Ranking orders a set of Slots according to an implementation specific metric.
type Ranking interface {
	// Returns the current ranks
	GetRanks() []*Slot

	// Returns the occupant of the ith place in the Ranking.
	// Returns nil if the place is unoccupied or out of bounds.
	At(i int) *Slot

	// Updates the return value of the GetRanks() method.
	// Should be called whenever a result that influences the
	// ranking becomes known.
	UpdateRanks()

	// All slots that resolve their qualification from this
	// ranking are added here.
	AddDependantSlots(slots ...*Slot)

	// Returns all dependant slots
	DependantSlots() []*Slot

	GraphNode
}

type BaseRanking struct {
	Ranks          []*Slot
	id             int
	dependantSlots []*Slot
}

func (r *BaseRanking) GetRanks() []*Slot {
	return r.Ranks
}

func (r *BaseRanking) At(i int) *Slot {
	if i >= len(r.Ranks) || i < 0 {
		return nil
	}
	return r.Ranks[i]
}

func (r *BaseRanking) UpdateRanks() {}

func (r *BaseRanking) AddDependantSlots(slots ...*Slot) {
	r.dependantSlots = append(r.dependantSlots, slots...)
}

func (r *BaseRanking) DependantSlots() []*Slot {
	return r.dependantSlots
}

func (r *BaseRanking) Id() int {
	return r.id
}

func NewBaseRanking() BaseRanking {
	id := NextNodeId()
	return BaseRanking{id: id}
}

// Creates a BaseRanking with the given slots as the ranks
func NewSlotRanking(slots []*Slot) *BaseRanking {
	ranking := NewBaseRanking()
	ranking.Ranks = slots
	return &ranking
}

type TieableRanking interface {
	Ranking

	// Returns a slice of slices of slots.
	//
	// A slice with multiple slots in it means the rank
	// is tied between them.
	TiedRanks() [][]*Slot

	// Returns the same ranks as TiedRanks but
	// without the tie breakers applied
	UnbrokenTiedRanks() [][]*Slot

	// Tie breakers which are a set of
	// rankings. All ties that contain the same set
	// of players as one of the tie breakers are
	// broken using the order of the tie breaker ranking.
	AddTieBreaker(tieBreaker Ranking)
	RemoveTieBreaker(tieBreaker Ranking)

	// Returns the ties that are in the
	// given top n of ranks
	BlockingTies(topN int) [][]*Slot

	// Returns the same ties as BlockingTies
	// but without tie breakers applied
	BlockingUnbrokenTies(topN int) [][]*Slot
}

type BaseTieableRanking struct {
	BaseRanking

	tiedRanks         [][]*Slot
	unbrokenTiedRanks [][]*Slot
	tieBreakers       map[string]Ranking

	RequiredUntiedRanks int
}

// Returns a slice of slices of slots.
//
// A slice with multiple slots in it means the rank
// is tied between them.
func (r *BaseTieableRanking) TiedRanks() [][]*Slot {
	return r.tiedRanks
}

// Returns the same ranks as TiedRanks but
// without the tie breakers applied
func (r *BaseTieableRanking) UnbrokenTiedRanks() [][]*Slot {
	return r.unbrokenTiedRanks
}

func (r *BaseTieableRanking) AddTieBreaker(tieBreaker Ranking) {
	tieHash := TieHash(tieBreaker.GetRanks())
	r.tieBreakers[tieHash] = tieBreaker
}

func (r *BaseTieableRanking) RemoveTieBreaker(tieBreaker Ranking) {
	tieHash := TieHash(tieBreaker.GetRanks())
	delete(r.tieBreakers, tieHash)
}

// Creates a hash of the given tie by sorting and concatenating
// the IDs of the players in the slots.
// Equal hashes mean the same players are in the ties
func TieHash(tie []*Slot) string {
	playerIds := make([]string, 0, len(tie))
	for _, s := range tie {
		playerIds = append(playerIds, s.Player().Id())
	}

	slices.Sort(playerIds)

	var sb strings.Builder
	for i, id := range playerIds {
		sb.WriteString(id)
		if i < len(playerIds)-1 {
			sb.WriteRune('\n')
		}
	}

	tieHash := sb.String()
	return tieHash
}

// Returns the ties that are in the
// given top n of ranks
func (r *BaseTieableRanking) BlockingTies(topN int) [][]*Slot {
	return topNBlockingTies(r.tiedRanks, topN)
}

// Returns the same ties as BlockingTies
// but without tie breakers applied
func (r *BaseTieableRanking) BlockingUnbrokenTies(topN int) [][]*Slot {
	return topNBlockingTies(r.unbrokenTiedRanks, topN)
}

func topNBlockingTies(ties [][]*Slot, topN int) [][]*Slot {
	blockingTies := make([][]*Slot, 0, topN-1)
	rankIndex := 0

	for _, t := range ties {
		if rankIndex >= topN {
			break
		}
		if len(t) > 1 {
			blockingTies = append(blockingTies, t)
		}
		rankIndex += len(t)
	}

	return blockingTies
}

// Attempts to find a tie breaker for the given
// tie and use it.
//
// On success returns a slice of slices with only one
// slot per nested slice.
// If no applicable tie breaker is present the returned slice
// contains only one nested slice with all slots in it.
func (r *BaseTieableRanking) TryTieBreak(tie []*Slot) [][]*Slot {
	if len(tie) == 1 {
		return [][]*Slot{tie}
	}

	tieHash := TieHash(tie)
	tieBreaker, exists := r.tieBreakers[tieHash]

	if !exists {
		return [][]*Slot{tie}
	}

	breakerRanks := tieBreaker.GetRanks()
	breakerIds := make([]string, 0, len(breakerRanks))
	for _, s := range breakerRanks {
		breakerIds = append(breakerIds, s.Player().Id())
	}

	slices.SortFunc(tie, func(a, b *Slot) int {
		indexA := slices.Index(breakerIds, a.Player().Id())
		indexB := slices.Index(breakerIds, b.Player().Id())
		return cmp.Compare(indexA, indexB)
	})

	brokenRanks := make([][]*Slot, 0, len(tie))
	for _, s := range tie {
		brokenTie := []*Slot{s}
		brokenRanks = append(brokenRanks, brokenTie)
	}

	return brokenRanks
}

// Embedders of the BaseTieableRanking should call this in their
// implementation of the UpdateRanks method to persist the update result
func (r *BaseTieableRanking) ProcessUpdate(updatedTiedRanks [][]*Slot) {
	r.unbrokenTiedRanks = updatedTiedRanks
	r.tiedRanks = r.applyTieBreakers(updatedTiedRanks)
	r.Ranks = flattenTiedRanks(r.tiedRanks)
}

func (r *BaseTieableRanking) applyTieBreakers(tiedRanks [][]*Slot) [][]*Slot {
	if len(r.tieBreakers) == 0 {
		return tiedRanks
	}

	tieBrokenRanks := make([][]*Slot, 0, len(tiedRanks)+5)
	for _, t := range tiedRanks {
		brokenTie := r.TryTieBreak(t)
		tieBrokenRanks = append(tieBrokenRanks, brokenTie...)
	}

	return tieBrokenRanks
}

func (r *BaseTieableRanking) String() string {
	var sb strings.Builder

	for _, r := range r.TiedRanks() {
		for _, s := range r {
			player := s.Player()
			if player == nil {
				sb.WriteString("Empty slot\n")
			} else {
				sb.WriteString(player.Id())
				sb.WriteRune('\n')
			}
		}
		sb.WriteString("---")
		sb.WriteRune('\n')
	}

	return sb.String()
}

func flattenTiedRanks(tiedRanks [][]*Slot) []*Slot {
	numRanks := 0
	for _, t := range tiedRanks {
		numRanks += len(t)
	}

	ranks := make([]*Slot, 0, numRanks)
	for _, t := range tiedRanks {
		ranks = append(ranks, t...)
	}

	return ranks
}

func NewBaseTieableRanking(requiredUntiedRanks int) BaseTieableRanking {
	ranking := BaseTieableRanking{
		BaseRanking:         NewBaseRanking(),
		tieBreakers:         make(map[string]Ranking),
		RequiredUntiedRanks: requiredUntiedRanks,
	}
	return ranking
}
