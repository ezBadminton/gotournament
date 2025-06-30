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

	firstRoundSize := nextPowerOfTwo(len(qualifications))

	pool := slices.Clone(qualifications)
	for len(pool) < firstRoundSize {
		pool = append(
			pool,
			&groupQualification{
				isBye: true,
				group: -1,
				place: -1,
			},
		)
	}

	matchups := [][]*groupQualification{pool}

	for len(matchups[0]) != 2 {
		nextMatchups := make([][]*groupQualification, 0, 2*len(matchups))
		for _, subPool := range matchups {
			upper, lower := r.splitQualifications(subPool)
			nextMatchups = append(nextMatchups, upper, lower)
		}
		matchups = nextMatchups
	}

	numRounds := getNumRounds(len(pool))
	seedMatchups := arrangeSeeds(numRounds)
	seeds := make([]int, len(qualifications))

	for i, matchup := range matchups {
		seedMatchup := seedMatchups[i]
		seeds[seedMatchup.seed1] = slices.Index(qualifications, matchup[0])
		if !matchup[1].isBye {
			seeds[seedMatchup.seed2] = slices.Index(qualifications, matchup[1])
		}
	}

	return seeds
}

// Splits the given qualifications into an upper and lower bracket.
// The qualifications slice is assumed to already be ordered by a predetermined seeding.
// The split tries to avoid placing qualifications from the same group into the same bracket
// and balances the same-group assignments between groups if it is unavoidable.
func (r *GroupQualificationRanking) splitQualifications(qualifications []*groupQualification) ([]*groupQualification, []*groupQualification) {
	for _, qual := range qualifications {
		// The inPool flag is set to false after r.pickSeedWithGroupConstraint picks a qualification out
		qual.inPool = true
	}

	originalSeeds := make(map[*groupQualification]int, len(qualifications))
	for i, qual := range qualifications {
		originalSeeds[qual] = i
	}

	numQuals := len(qualifications)
	numRounds := getNumRounds(numQuals)
	seedMatchups := arrangeSeeds(numRounds)

	upperSeedMatchups := seedMatchups[:len(seedMatchups)/2]
	lowerSeedMatchups := seedMatchups[len(seedMatchups)/2:]

	upperSeeds := make([]int, len(seedMatchups))
	lowerSeeds := make([]int, len(seedMatchups))

	for i, matchup := range upperSeedMatchups {
		upperSeeds[i] = matchup.seed1
		upperSeeds[len(seedMatchups)-1-i] = matchup.seed2
	}
	for i, matchup := range lowerSeedMatchups {
		lowerSeeds[i] = matchup.seed1
		lowerSeeds[len(seedMatchups)-1-i] = matchup.seed2
	}

	// Counters for how many qualifications from each group are in upper/lower bracket
	upperGroups := make(map[int]int)
	lowerGroups := make(map[int]int)

	upper := make([]*groupQualification, 0, numQuals/2)
	lower := make([]*groupQualification, 0, numQuals/2)

	for _, seed := range upperSeeds {
		qualification := r.pickSeedWithGroupConstraint(qualifications, seed, upperGroups)
		group := qualification.group
		if group >= 0 {
			upperGroups[group] = upperGroups[group] + 1
		}
		upper = append(upper, qualification)
	}

	for _, seed := range lowerSeeds {
		qualification := r.pickSeedWithGroupConstraint(qualifications, seed, lowerGroups)
		group := qualification.group
		if group >= 0 {
			lowerGroups[group] = lowerGroups[group] + 1
		}
		lower = append(lower, qualification)
	}

	slices.SortFunc(upper, func(a, b *groupQualification) int { return cmp.Compare(originalSeeds[a], originalSeeds[b]) })
	slices.SortFunc(lower, func(a, b *groupQualification) int { return cmp.Compare(originalSeeds[a], originalSeeds[b]) })

	return upper, lower
}

// Returns the groupQualification that is at the seed index of the pool unless that
// qualification is already out of the pool or the group constraint is violated.
// Finds an alternative candidate for the seed by looking at qualifications of the same place
// and eventually considers the whole pool, breaking the group constraint only if necessary.
func (r *GroupQualificationRanking) pickSeedWithGroupConstraint(
	pool []*groupQualification,
	seed int,
	groupConstraint map[int]int,
) *groupQualification {
	directCandidate := pool[seed]
	if directCandidate.inPool && groupConstraint[directCandidate.group] == 0 {
		directCandidate.inPool = false
		return directCandidate
	}

	for _, alternativeCandidate := range pool {
		inPool := alternativeCandidate.inPool
		samePlace := alternativeCandidate.place == directCandidate.place
		noGroupConflict := groupConstraint[alternativeCandidate.group] == 0
		if inPool && samePlace && noGroupConflict {
			alternativeCandidate.inPool = false
			return alternativeCandidate
		}
	}

	alternativeCandidates := make([]*groupQualification, 0, len(pool))
	for _, qualification := range pool {
		if qualification.inPool && !qualification.isBye {
			alternativeCandidates = append(alternativeCandidates, qualification)
		}
	}
	slices.SortStableFunc(alternativeCandidates, func(a, b *groupQualification) int {
		if a == directCandidate {
			return -1
		}
		if b == directCandidate {
			return 1
		}
		if a.place == b.place {
			return 0
		}

		aDistance := a.place - directCandidate.place
		bDistance := b.place - directCandidate.place

		aAbsDistance := abs(aDistance)
		bAbsDistance := abs(bDistance)

		if aAbsDistance == bAbsDistance {
			// Prioritize lower placed alternatives
			if aDistance > 0 {
				return -1
			} else {
				return 1
			}
		} else if aAbsDistance < bAbsDistance {
			return -1
		} else {
			return 1
		}
	})

	slices.SortStableFunc(alternativeCandidates, func(a, b *groupQualification) int {
		return cmp.Compare(groupConstraint[a.group], groupConstraint[b.group])
	})

	if len(alternativeCandidates) > 0 {
		alternativeCandidates[0].inPool = false
		return alternativeCandidates[0]
	}

	return nil
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

type groupQualification struct {
	group, place int
	isBye        bool
	inPool       bool
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
