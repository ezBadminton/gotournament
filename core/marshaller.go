package core

import (
	"encoding/json"
	"maps"
)

func marshalRankingsAndSlots(entries Ranking, rankingGraph *RankingGraph) map[string]any {
	rankings := make(map[int][][]int)
	slots := make(map[int]string)

	for ranking := range rankingGraph.BreadthSearchIter(entries) {
		rankingSlots := ranking.Ranks()
		for _, slot := range rankingSlots {
			_, ok := slots[slot.Id]
			if ok {
				continue
			}
			slots[slot.Id] = marshalSlot(slot)
		}

		rankings[ranking.Id()] = marshalRanking(ranking)
	}

	result := map[string]any{
		"rankings": rankings,
		"slots":    slots,
	}

	return result
}

func marshalSlot(slot *Slot) string {
	if slot.IsBye() {
		if slot.Bye.Drawn {
			return "db"
		} else {
			return "b"
		}
	} else if p := slot.Player; p != nil {
		return p.Id()
	} else {
		return ""
	}
}

func marshalRanking(ranking Ranking) [][]int {
	switch r := ranking.(type) {
	case TieableRanking:
		return marshalTieableRanking(r)
	default:
		ranks := r.Ranks()
		slotIds := make([][]int, len(ranks))
		for i, slot := range ranks {
			slotIds[i] = []int{slot.Id}
		}
		return slotIds
	}
}

func marshalTieableRanking(ranking TieableRanking) [][]int {
	ranks := ranking.TiedRanks()
	slotIds := make([][]int, len(ranks))
	for i, rank := range ranks {
		rankSlotIds := make([]int, len(rank))
		for i, slot := range rank {
			rankSlotIds[i] = slot.Id
		}
		slotIds[i] = rankSlotIds
	}
	return slotIds
}

func marshalMetrics(ranking *MatchMetricRanking) []*MatchMetrics {
	slots := ranking.Ranks()
	metrics := make([]*MatchMetrics, 0, len(slots))
	for _, s := range slots {
		if s.Player == nil {
			continue
		}
		metrics = append(metrics, ranking.Metrics[s.Player])
	}
	return metrics
}

func marshalMatchList(matchList *matchList) map[string]any {
	rounds := make([][]map[string]any, len(matchList.Rounds))
	for i, round := range matchList.Rounds {
		roundMatches := make([]map[string]any, len(round.Matches))
		for i, match := range round.Matches {
			roundMatches[i] = marshalMatch(match)
		}
		rounds[i] = roundMatches
	}

	result := map[string]any{
		"rounds": rounds,
	}

	return result
}

func marshalMatch(match *Match) map[string]any {
	score := make([][]int, 0)
	if match.Score != nil {
		score = append(
			score,
			match.Score.Points1(),
			match.Score.Points2(),
		)
	}
	var locationId string
	if match.Location != nil {
		locationId = match.Location.Id()
	}
	result := map[string]any{
		"slot1": match.Slot1.Id,
		"slot2": match.Slot2.Id,
		"score": score,
		"start": match.StartTime.UnixMilli(),
		"end":   match.EndTime.UnixMilli(),
		"loc":   locationId,
	}
	return result
}

func marshalSingleElimination(tournament *SingleElimination) map[string]any {
	ranksAndSlots := marshalRankingsAndSlots(tournament.Entries, tournament.RankingGraph)
	matchList := marshalMatchList(tournament.matchList)
	result := map[string]any{
		"type": "SingleElimination",
	}

	maps.Copy(result, ranksAndSlots)
	maps.Copy(result, matchList)

	return result
}

func marshalRoundRobin(tournament *RoundRobin) map[string]any {
	ranksAndSlots := marshalRankingsAndSlots(tournament.Entries, tournament.RankingGraph)
	matchList := marshalMatchList(tournament.matchList)
	result := map[string]any{
		"type":    "RoundRobin",
		"metrics": marshalMetrics(tournament.FinalRanking),
	}

	maps.Copy(result, ranksAndSlots)
	maps.Copy(result, matchList)

	return result
}

func marshalSingleEliminationWithConsolation(tournament *SingleEliminationWithConsolation) map[string]any {
	ranksAndSlots := marshalRankingsAndSlots(tournament.Entries, tournament.RankingGraph)
	mainBracket := marshalConsolationBracket(tournament.MainBracket)
	result := map[string]any{
		"type":        "SingleEliminationWithConsolation",
		"mainBracket": mainBracket,
	}

	maps.Copy(result, ranksAndSlots)

	return result
}

func marshalConsolationBracket(bracket *ConsolationBracket) map[string]any {
	matchList := marshalMatchList(bracket.matchList)
	nested := make([]map[string]any, len(bracket.Consolations))
	for i, bracket := range bracket.Consolations {
		nested[i] = marshalConsolationBracket(bracket)
	}
	result := map[string]any{
		"matches":      matchList,
		"consolations": nested,
	}
	return result
}

func marshalDoubleElimination(tournamet *DoubleElimination) map[string]any {
	ranksAndSlots := marshalRankingsAndSlots(tournamet.Entries, tournamet.RankingGraph)
	matchList := marshalMatchList(tournamet.matchList)
	result := map[string]any{
		"type": "DoubleElimination",
	}

	maps.Copy(result, ranksAndSlots)
	maps.Copy(result, matchList)

	return result
}

func marshalGroupPhase(tournament *GroupPhase) map[string]any {
	groupMatchLists := make([]any, 0)
	groupMetrics := make([][]*MatchMetrics, len(tournament.Groups))
	for _, g := range tournament.Groups {
		matchList := marshalMatchList(g.matchList)
		groupMatchLists = append(groupMatchLists, matchList["rounds"])

		metrics := marshalMetrics(g.FinalRanking)
		groupMetrics = append(groupMetrics, metrics)
	}

	result := map[string]any{
		"type":         "GroupPhase",
		"groupRounds":  groupMatchLists,
		"groupMetrics": groupMetrics,
	}

	return result
}

func marshalGroupKnockout(tournament *GroupKnockout) map[string]any {
	ranksAndSlots := marshalRankingsAndSlots(tournament.Entries, tournament.RankingGraph)
	groupPhase := marshalGroupPhase(tournament.GroupPhase)
	var koPhase map[string]any
	switch ko := tournament.KnockOutTournament.(type) {
	case *SingleElimination:
		matchList := marshalMatchList(ko.matchList)
		koPhase = map[string]any{
			"type": "SingleElimination",
		}
		maps.Copy(koPhase, matchList)
	case *SingleEliminationWithConsolation:
		mainBracket := marshalConsolationBracket(ko.MainBracket)
		koPhase = map[string]any{
			"type":        "SingleEliminationWithConsolation",
			"mainBracket": mainBracket,
		}
	case *DoubleElimination:
		matchList := marshalMatchList(ko.matchList)
		koPhase = map[string]any{
			"type": "DoubleElimination",
		}
		maps.Copy(koPhase, matchList)
	default:
		panic("group knockout marshaller: unknown ko phase tournament type")
	}

	result := map[string]any{
		"type":       "GroupKnockout",
		"groupPhase": groupPhase,
		"koPhase":    koPhase,
	}

	maps.Copy(result, ranksAndSlots)

	return result
}

func (t *SingleElimination) MarshalJSON() ([]byte, error) {
	anymap := marshalSingleElimination(t)
	return json.Marshal(anymap)
}

func (t *SingleEliminationWithConsolation) MarshalJSON() ([]byte, error) {
	anymap := marshalSingleEliminationWithConsolation(t)
	return json.Marshal(anymap)
}

func (t *RoundRobin) MarshalJSON() ([]byte, error) {
	anymap := marshalRoundRobin(t)
	return json.Marshal(anymap)
}

func (t *GroupKnockout) MarshalJSON() ([]byte, error) {
	anymap := marshalGroupKnockout(t)
	return json.Marshal(anymap)
}

func (t *DoubleElimination) MarshalJSON() ([]byte, error) {
	anymap := marshalDoubleElimination(t)
	return json.Marshal(anymap)
}
