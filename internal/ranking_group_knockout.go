package internal

import "slices"

type GroupKnockoutRanking struct {
	BaseTieableRanking

	groupPhase *GroupPhase
	knockOut   *BaseTournament[*EliminationRanking]
}

func (r *GroupKnockoutRanking) updateRanks() {
	groups := r.groupPhase.Groups

	groupRanks := make([][][]*Slot, 0, len(groups))
	for _, g := range groups {
		ranking := g.FinalRanking
		groupRanks = append(groupRanks, ranking.TiedRanks())
	}

	combinedGroupRanks := make([][]*Slot, 8)
	rankFound := true
	for i := 0; rankFound; i += 1 {
		combinedGroupRank := make([]*Slot, 0, len(groups))
		for _, ranks := range groupRanks {
			if i >= len(ranks) {
				continue
			}
			combinedGroupRank = append(combinedGroupRank, ranks[i]...)
		}
		rankFound = len(combinedGroupRank) > 0
		combinedGroupRanks = append(combinedGroupRanks, combinedGroupRank)
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
	ranking.updateRanks()

	return ranking
}
