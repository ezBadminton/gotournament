package core

import (
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"
)

var (
	ErrBothBye        = errors.New("both bye")
	ErrBothWalkover   = errors.New("both walkover")
	ErrByeAndWalkover = errors.New("bye and walkover")
	ErrNoScore        = errors.New("no score")
	ErrEqualScore     = errors.New("equal score")
)

// A match with two slots for the opponents.
//
// It also has information about the result
// of the match and some meta data.
type Match struct {
	// The first opponent slot
	Slot1 *Slot
	// The second opponent slot
	Slot2 *Slot

	// An iterator that goes over the two slots
	Slots iter.Seq[*Slot]

	// Score of the match or
	// nil when the match is not completed
	Score Score

	// The location where this match is played
	Location Location

	// The time when the match was started
	// If this is not zero and the EndTime
	// is zero it means the match is in progress
	StartTime time.Time

	// The time when the match concluded
	// When this is not zero the StartTime
	// and Score should also not be zero/nil
	EndTime time.Time

	// A list of players who withdrew from this
	// match
	WithdrawnPlayers []Player

	// Id for graph node hashing
	id int
}

func (m *Match) GetWinner() (*Slot, error) {
	bye1 := m.Slot1.IsBye()
	bye2 := m.Slot2.IsBye()

	if bye1 && bye2 {
		return nil, ErrBothBye
	}

	withdrawn := m.WithdrawnSlots()

	if len(withdrawn) == 1 {
		notWithdrawn := m.OtherSlot(withdrawn[0])
		if notWithdrawn.IsBye() {
			return nil, ErrByeAndWalkover
		} else {
			return notWithdrawn, nil
		}
	} else if len(withdrawn) == 2 {
		return nil, ErrBothWalkover
	}

	if !bye1 && bye2 {
		return m.Slot1, nil
	}
	if !bye2 && bye1 {
		return m.Slot2, nil
	}

	if m.Score == nil {
		return nil, ErrNoScore
	}

	winnerIndex, err := m.Score.GetWinner()
	if err != nil {
		return nil, ErrEqualScore
	}

	if winnerIndex == 0 {
		return m.Slot1, nil
	}
	if winnerIndex == 1 {
		return m.Slot2, nil
	}

	panic("Someting went wrong while getting the match's winner")
}

func (m *Match) OtherSlot(slot *Slot) *Slot {
	if slot == m.Slot1 {
		return m.Slot2
	}
	if slot == m.Slot2 {
		return m.Slot1
	}

	panic("Slot is not in the Match")
}

// Returns the slots that are occupied by withdrawn players
func (m *Match) WithdrawnSlots() []*Slot {
	withdrawn := make([]*Slot, 0, 2)
	if len(m.WithdrawnPlayers) == 0 {
		return withdrawn
	}

	withdrawn1, withdrawn2 := false, false
	for _, p := range m.WithdrawnPlayers {
		if p.Id() == m.Slot1.Player.Id() {
			withdrawn1 = true
		}
		if p.Id() == m.Slot2.Player.Id() {
			withdrawn2 = true
		}
	}

	if withdrawn1 {
		withdrawn = append(withdrawn, m.Slot1)
	}
	if withdrawn2 {
		withdrawn = append(withdrawn, m.Slot2)
	}

	return withdrawn
}

func (m *Match) IsWalkover() bool {
	return len(m.WithdrawnSlots()) > 0
}

func (m *Match) HasBye() bool {
	return m.Slot1.IsBye() || m.Slot2.IsBye()
}

func (m *Match) HasDrawnBye() bool {
	bye1 := m.Slot1.Bye
	bye2 := m.Slot2.Bye

	return (bye1 != nil && bye1.Drawn) || (bye2 != nil && bye2.Drawn)
}

func (m *Match) ContainsPlayer(player Player) bool {
	id := player.Id()
	return m.Slot1.Player.Id() == id || m.Slot2.Player.Id() == id
}

// Returns true when the given player has withdrawn
// and is occupying one of the slots
func (m *Match) IsPlayerWithdrawn(player Player) bool {
	withdrawnSlots := m.WithdrawnSlots()
	for _, s := range withdrawnSlots {
		if s.Player == player {
			return true
		}
	}
	return false
}

func (m *Match) StartMatch() error {
	if !m.StartTime.IsZero() {
		return errors.New("Match already started")
	}
	m.StartTime = time.Now()
	return nil
}

func (m *Match) EndMatch(score Score) error {
	if m.StartTime.IsZero() {
		return errors.New("Match cannot end before it started")
	}
	if !m.EndTime.IsZero() {
		return errors.New("Match already ended")
	}
	m.Score = score
	m.EndTime = time.Now()
	return nil
}

func (m *Match) Id() int {
	return m.id
}

func (m *Match) String() string {
	var sb strings.Builder
	p1 := m.Slot1.Player
	if p1 == nil {
		sb.WriteString("[Empty]")
	} else {
		sb.WriteString(p1.Id())
	}
	sb.WriteString(" vs. ")
	p2 := m.Slot2.Player
	if p2 == nil {
		sb.WriteString("[Empty]")
	} else {
		sb.WriteString(p2.Id())
	}

	if m.Score != nil {
		p1, p2 := m.Score.Points1(), m.Score.Points2()
		sb.WriteRune('\t')
		for i := range len(p1) {
			setString := fmt.Sprintf("%v - %v ", p1[i], p2[i])
			sb.WriteString(setString)
		}
	}

	return sb.String()
}

func NewMatch(slot1, slot2 *Slot) *Match {
	id := NextId()

	iterator := func(yield func(s *Slot) bool) {
		if !yield(slot1) {
			return
		}
		yield(slot2)
	}

	match := &Match{
		Slot1: slot1,
		Slot2: slot2,
		Slots: iterator,
		id:    id,
	}
	return match
}

// Returns true if one or more of the given matches have started
func MatchesStarted(matches ...*Match) bool {
	for _, m := range matches {
		if !m.StartTime.IsZero() {
			return true
		}
	}
	return false
}

// The result of a match.
//
// The scores are slices to be able to model
// competitions where a match consists of
// multiple sets (e.g. Tennis)
type Score interface {
	// Points of first opponent
	Points1() []int

	// Points of second opponent
	Points2() []int

	// Returns either 0 or 1 whether the
	// first opponent won or the second.
	// Errors when no winner is determined.
	GetWinner() (int, error)

	// Returns a new Score that has Points1
	// and Points2 flipped
	Invert() Score
}

// A Round is a list of matches that can be played in
// parallel during a tournament.
// The matches of a round depend on the completion
// of all previous rounds.
type Round struct {
	// The matches that are played in this round
	Matches []*Match

	// Other Rounds that this Round is composed of
	// Is empty when no underlying rounds exist
	NestedRounds []*Round
}

// A Location is a court or a field
// where a match is played on
type Location interface {
	Id() string
}
