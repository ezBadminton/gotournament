package internal

import "slices"

type GroupPhaseRanking struct {
	BaseTieableRanking

	groups            []*RoundRobin
	crossGroupRanking TieableRanking

	GroupTies map[int][][]*Slot
}

func (r *GroupPhaseRanking) UpdateRanks() {
	for _, g := range r.groups {
		if !g.MatchList.MatchesComplete() {
			r.ProcessUpdate([][]*Slot{})
			return
		}
	}

	numGroups := len(r.groups)

	groupRankings := make([]*MatchMetricRanking, 0, numGroups)
	for _, g := range r.groups {
		groupRankings = append(groupRankings, g.FinalRanking.(*MatchMetricRanking))
	}

	contestedRank, numContested := r.contestedRank()
	r.collectGroupTies(groupRankings)
	tiesPresent := len(r.GroupTies) > 0
	maxNumRanks := len(r.groups[numGroups-1].Entries.GetRanks())

	ranks := make([][]*Slot, 0, maxNumRanks*numGroups)
	for i := range maxNumRanks {
		var rank [][]*Slot
		if i == contestedRank && !tiesPresent {
			rank = collectContestedRank(i, numContested, groupRankings, r.crossGroupRanking)
		} else {
			rank = collectRank(i, groupRankings)
		}
		ranks = append(ranks, rank...)
	}

	r.ProcessUpdate(ranks)
}

// The cross group ties are populated when there is a
// contested qualification between the occupants
// of one rank across different groups.
// They are always empty while the groups have
// blocking ties locally. Those are found in
// t.GroupTies.
func (r *GroupPhaseRanking) CrossGroupTies() [][]*Slot {
	return r.BlockingTies(r.RequiredUntiedRanks)
}

// When the number of qualifications is not divisible by
// the number of groups then there are qualifications
// that are contested among one rank
// across all groups. The index of that rank
// and the amount of contested qualifications is returned.
// Returns -1, 0 if no rank is contested.
func (r *GroupPhaseRanking) contestedRank() (int, int) {
	numQualifications := r.BaseTieableRanking.RequiredUntiedRanks
	numContested := numQualifications % len(r.groups)
	if numContested == 0 {
		return -1, 0
	}

	contestedIndex := numQualifications / len(r.groups)
	return contestedIndex, numContested
}

func collectRank(rank int, rankings []*MatchMetricRanking) [][]*Slot {
	ranks := make([][]*Slot, 0, len(rankings))

	for _, r := range rankings {
		slot := r.At(rank)
		if slot == nil {
			continue
		}
		metrics, ok := r.Metrics[slot.Player()]
		if ok && metrics.Withdrawn {
			slot = NewByeSlot(false)
		}
		ranks = append(ranks, []*Slot{slot})
	}

	return ranks
}

func collectContestedRank(
	rank, numContested int,
	rankings []*MatchMetricRanking,
	crossRanking TieableRanking,
) [][]*Slot {
	crossRanks := crossRanking.TiedRanks()
	contestants := make(map[int][]*Slot)

	for _, r := range rankings {
		slot := r.At(rank)
		if slot == nil {
			continue
		}
		crossRank := slices.IndexFunc(
			crossRanks,
			func(slots []*Slot) bool { return slices.Contains(slots, slot) },
		)
		if crossRank == -1 {
			panic("Slot was not found in the cross ranking")
		}
		metrics, ok := r.Metrics[slot.Player()]
		if ok && metrics.Withdrawn {
			slot = NewByeSlot(false)
		}

		_, ok = contestants[crossRank]
		if !ok {
			contestants[crossRank] = make([]*Slot, 0, 4)
		}

		contestants[crossRank] = append(contestants[crossRank], slot)
	}

	ranks := make([][]*Slot, 0, len(contestants))
	contestantKeys := make([]int, 0, len(contestants))
	for k := range contestants {
		contestantKeys = append(contestantKeys, k)
	}
	slices.Sort(contestantKeys)

	for _, k := range contestantKeys {
		tie := contestants[k]
		if len(tie) <= numContested {
			for _, slot := range tie {
				ranks = append(ranks, []*Slot{slot})
			}
		} else {
			ranks = append(ranks, tie)
		}
		numContested -= len(tie)
	}

	return ranks
}

func (r *GroupPhaseRanking) collectGroupTies(rankings []*MatchMetricRanking) {
	r.GroupTies = make(map[int][][]*Slot)

	for i, ranking := range rankings {
		blockingTies := ranking.BlockingTies(ranking.RequiredUntiedRanks)
		if len(blockingTies) > 0 {
			r.GroupTies[i] = blockingTies
		}
	}
}

func NewGroupPhaseRanking(
	groups []*RoundRobin,
	numQualifications int,
	crossGroupRanking TieableRanking,
	rankingGraph *RankingGraph,
) *GroupPhaseRanking {
	ranking := NewBaseTieableRanking(numQualifications)
	groupPhaseRanking := &GroupPhaseRanking{
		BaseTieableRanking: ranking,
		groups:             groups,
		crossGroupRanking:  crossGroupRanking,
	}

	rankingGraph.AddVertex(crossGroupRanking)
	rankingGraph.AddVertex(groupPhaseRanking)
	rankingGraph.AddEdge(crossGroupRanking, groupPhaseRanking)

	return groupPhaseRanking
}
