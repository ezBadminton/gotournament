package internal

import "slices"

type GroupKnockoutRanking struct {
	BaseTieableRanking

	groupPhase *GroupPhase
	knockOut   *BaseTournament[*EliminationRanking]
}

func (r *GroupKnockoutRanking) UpdateRanks() {
	groups := r.groupPhase.Groups

	groupRanks := make([][][]*Slot, 0, len(groups))
	for _, g := range groups {
		ranking := g.FinalRanking
		groupRanks = append(groupRanks, ranking.TiedRanks())
	}
	numRanks := 0
	for _, ranks := range groupRanks {
		numRanks += len(ranks)
	}

	combinedGroupRanks := make([][]*Slot, numRanks)
	rankFound := true
	for i := 0; rankFound; i += 1 {
		rankFound = false
		for _, ranks := range groupRanks {
			if i >= len(ranks) {
				continue
			}
			rankFound = true
			combinedGroupRanks = append(combinedGroupRanks, ranks[i])
		}
	}

	knockOutRanks := r.knockOut.FinalRanking.TiedRanks()

	ranks := slices.Concat(combinedGroupRanks, knockOutRanks)
	ranks = RemoveDoubleRanks(ranks)

	r.ProcessUpdate(ranks)
}

func NewGroupKnockoutRanking(groupPhase *GroupPhase, knockOut *BaseTournament[*EliminationRanking]) *GroupKnockoutRanking {
	baseRanking := NewBaseTieableRanking(0)
	ranking := &GroupKnockoutRanking{
		BaseTieableRanking: baseRanking,
		groupPhase:         groupPhase,
		knockOut:           knockOut,
	}
	ranking.UpdateRanks()

	return ranking
}
