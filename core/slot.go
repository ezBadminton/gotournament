package core

// A Slot is either a spot in a Ranking or one of two places
// in a Match.
//
// A Slot can represent one of 3 things:
//  - An actual player
//  - A not yet determined qualification called a Placement.
//    (e.g. the slots of a final match are the winners
//    of the semi-finals)
//  - A free win (bye) for the match opponent
//
// The Slot changes what it is representing depending
// on the state of the tournament (e.g. when the results of the
// semi-finals become known the final slots go from
// undetermined qualifications to actual players or
// when a player withdraws from the tournament their slot
// goes from actual player to bye).
type Slot struct {
	player    Player
	placement Placement
	bye       *Bye
}

// Returns whether this slot is an effective bye.
//
// Effective bye means it is also true when the
// slot inherits a bye slot via placement
func (s *Slot) IsBye() bool {
	if s.bye != nil {
		return true
	}

	if s.placement != nil && s.placement.Slot() != nil {
		return s.placement.Slot().IsBye()
	}

	return false
}

// Updates the return value of the Player method.
// This method is called when the ranking that this
// slot is dependant on updates. The dependency is stored
// in the ranking's list of dependant slots
// [Ranking.GetDependantSlots].
func (s *Slot) Update() {
	if s.placement == nil || s.bye != nil {
		return
	}
	slot := s.placement.Slot()
	if slot == nil {
		s.player = nil
		return
	}
	s.player = slot.player
}

func NewPlayerSlot(player Player) *Slot {
	return &Slot{player: player}
}

func NewPlacementSlot(placement Placement) *Slot {
	slot := &Slot{placement: placement}
	placement.Ranking().addDependantSlots(slot)
	return slot
}

func NewByeSlot(drawn bool) *Slot {
	bye := &Bye{Drawn: drawn}
	return &Slot{bye: bye}
}

// A Player is either a person or a team who is
// taking part in a tournament.
type Player interface {
	// Returns an ID that is unique among the players of
	// a tournament
	Id() string
}

// A Bye is a free win for a player.
type Bye struct {
	// This is true when the bye is due to a draw and false
	// when it's due to a player withdrawal
	Drawn bool
}
