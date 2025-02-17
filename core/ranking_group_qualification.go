package core

import (
	"cmp"
	"slices"
)

type GroupQualificationRanking struct {
	BaseRanking

	source     *GroupPhaseRanking
	placements []*BlockingPlacement
}

func (r *GroupQualificationRanking) updateRanks() {
	seeds := r.arrangeKnockOutSeeds()

	for i, placement := range r.placements {
		placement.blocking = !r.source.QualificationComplete
		placement.place = seeds[i]
	}
}

func (r *GroupQualificationRanking) arrangeKnockOutSeeds() []int {
	qualifications := r.createGroupQualifications()

	firstRoundSize := previousPowerOfTwo(len(qualifications))
	preRoundSize := len(qualifications) - firstRoundSize

	numFirstRoundSlots := len(qualifications) - preRoundSize*2
	firstQuals := qualifications[:numFirstRoundSlots]
	preQuals := slices.Clone(qualifications[numFirstRoundSlots:])

	preMatchups := make([]*qualificationMatchup, 0, firstRoundSize)

	// The first round qualifications get a bye in the pre round
	for _, qual := range firstQuals {
		preMatchups = append(preMatchups, &qualificationMatchup{a: qual, b: nil})
	}

	for range preRoundSize {
		a, b := getPreRoundMatchup(preQuals)
		preMatchups = append(preMatchups, &qualificationMatchup{a, b})
	}
	slices.SortFunc(preMatchups, compareMatchups)

	// Now pair the pre-round matchups into matchups of matchups
	// which will be the first full round

	firstNamed := preMatchups[:len(preMatchups)/2]

	pool := preMatchups[len(preMatchups)/2:]
	slices.SortFunc(pool, compareMatchupsInvertGroup)

	secondNamed := make([]*qualificationMatchup, len(preMatchups)/2)
	for i, m := range firstNamed {
		i = len(firstNamed) - 1 - i
		matchup := getLowestMatchup(pool, m)
		if matchup == nil {
			matchup = getLowestMatchup(pool, nil)
		}
		secondNamed[i] = matchup
	}

	firstMatchups := slices.Concat(firstNamed, secondNamed)

	seeds := make([]int, 0, len(firstMatchups)*2)
	for _, m := range firstMatchups {
		seeds = append(seeds, slices.Index(qualifications, m.a))
	}
	for _, m := range slices.Backward(firstMatchups) {
		if m.b != nil {
			seeds = append(seeds, slices.Index(qualifications, m.b))
		}
	}

	return seeds
}

// The group qualifications are palceholder structs
// for the slots from the final group phase ranking.
// Their stand-in simplifies the seeding of the
// qualified players for the knock out phase.
func (r *GroupQualificationRanking) createGroupQualifications() []*groupQualification {
	numQualifcations := r.source.RequiredUntiedRanks
	numGroups := len(r.source.groups)

	numUncontested := numQualifcations / numGroups

	qualifications := make([]*groupQualification, 0, (numUncontested+1)*numGroups)
	for place := range numUncontested {
		for group := range numGroups {
			qualifications = append(
				qualifications,
				&groupQualification{group: group, place: place},
			)
		}
	}

	numContested := numQualifcations % numGroups
	contestedSlots := r.source.Ranks()[len(qualifications) : len(qualifications)+numContested]

	for i, slot := range contestedSlots {
		group := i
		if slot.Player != nil {
			group = r.groupOfSlot(slot)
			if group == -1 {
				panic("Could not find slot in the groups")
			}
		}

		qualifications = append(
			qualifications,
			&groupQualification{group: group, place: numUncontested},
		)
	}

	return qualifications
}

func (r *GroupQualificationRanking) groupOfSlot(slot *Slot) int {
	for i, g := range r.source.groups {
		entrySlots := g.Entries.Ranks()
		if slices.Contains(entrySlots, slot) {
			return i
		}
	}

	return -1
}

func getPreRoundMatchup(qualifications []*groupQualification) (*groupQualification, *groupQualification) {
	var highest *groupQualification
	for i, q := range qualifications {
		if q != nil {
			highest = q
			qualifications[i] = nil
			break
		}
	}

	lowest := getLowestQualification(qualifications, highest.group)
	if lowest == nil {
		lowest = getLowestQualification(qualifications, -1)
	}

	return highest, lowest
}

func getLowestQualification(qualifications []*groupQualification, groupConstraint int) *groupQualification {
	var lowest *groupQualification
	for i, q := range slices.Backward(qualifications) {
		if q != nil && q.group != groupConstraint {
			lowest = q
			qualifications[i] = nil
			break
		}
	}
	return lowest
}

func getLowestMatchup(pool []*qualificationMatchup, groupConstraint *qualificationMatchup) *qualificationMatchup {
	var lowest *qualificationMatchup
	for i, m := range slices.Backward(pool) {
		if m != nil && !overlappingGroups(m, groupConstraint) {
			lowest = m
			pool[i] = nil
			break
		}
	}
	return lowest
}

type groupQualification struct {
	group, place int
}

func compareQualifications(a, b *groupQualification) int {
	placeComparison := cmp.Compare(a.place, b.place)
	if placeComparison != 0 {
		return placeComparison
	}

	groupComparison := cmp.Compare(a.group, b.group)
	return groupComparison
}

func compareQualificationsInvertGroup(a, b *groupQualification) int {
	placeComparison := cmp.Compare(a.place, b.place)
	if placeComparison != 0 {
		return placeComparison
	}

	groupComparison := -1 * cmp.Compare(a.group, b.group)
	return groupComparison
}

type qualificationMatchup struct {
	a, b *groupQualification
}

func (m *qualificationMatchup) getHigherPlaced() *groupQualification {
	if m.b == nil {
		return m.a
	}

	comparison := compareQualifications(m.a, m.b)
	if comparison == -1 {
		return m.a
	} else {
		return m.b
	}
}

func compareMatchups(a, b *qualificationMatchup) int {
	return compareQualifications(a.getHigherPlaced(), b.getHigherPlaced())
}

func compareMatchupsInvertGroup(a, b *qualificationMatchup) int {
	return compareQualificationsInvertGroup(a.getHigherPlaced(), b.getHigherPlaced())
}

func overlappingGroups(a, b *qualificationMatchup) bool {
	if b == nil {
		return false
	}

	switch {
	case a.a.group == b.a.group:
		return true
	case b.b != nil && a.a.group == b.b.group:
		return true
	case a.b != nil && a.b.group == b.a.group:
		return true
	case a.b != nil && b.b != nil && a.b.group == b.b.group:
		return true
	}
	return false
}

// Returns the power of two that is immediately smaller
// or equal to from.
func previousPowerOfTwo(from int) int {
	result := 1
	for from > 1 {
		from >>= 1
		result <<= 1
	}
	return result
}

func NewGroupQualificationRanking(source *GroupPhaseRanking, rankingGraph *RankingGraph) *GroupQualificationRanking {
	numQualifcations := source.RequiredUntiedRanks

	placements := make([]*BlockingPlacement, 0, numQualifcations)
	slots := make([]*Slot, 0, numQualifcations)
	for range numQualifcations {
		placements = append(placements, NewBlockingPlacement(source, 0, true))
	}
	for _, p := range placements {
		slots = append(slots, NewPlacementSlot(p))
	}

	baseRanking := NewSlotRanking(slots)
	ranking := &GroupQualificationRanking{
		BaseRanking: *baseRanking,
		source:      source,
		placements:  placements,
	}
	ranking.addDependantSlots(slots...)
	ranking.updateRanks()

	rankingGraph.AddVertex(ranking)
	rankingGraph.AddEdge(source, ranking)

	return ranking
}
