package core

import (
	"maps"
)

type TournamentMarshaller struct {
	idMap map[int]string
}

func newTournamentMarshaller(tournament MatchLister, getMatchId func(i int) string) TournamentMarshaller {
	idMap := make(map[int]string)
	matches := tournament.MatchList().Matches
	for i, m := range matches {
		idMap[m.id] = getMatchId(i)
	}
	return TournamentMarshaller{idMap}
}

func (m *TournamentMarshaller) marshalEntriesAndFinal(entries, finalRanking Ranking) map[string]any {
	result := map[string]any{
		"entries":      m.marshalRanking(entries),
		"finalRanking": m.marshalRanking(finalRanking),
	}
	return result
}

func (m *TournamentMarshaller) marshalSlot(slot *Slot) map[string]any {
	var occupant string
	if slot.IsBye() {
		if slot.Bye != nil && slot.Bye.Drawn {
			occupant = "db"
		} else {
			occupant = "b"
		}
	} else if p := slot.Player; p != nil {
		occupant = p.Id()
	} else {
		occupant = ""
	}
	return map[string]any{
		"id":       slot.Id,
		"occupant": occupant,
	}
}

func (m *TournamentMarshaller) marshalRanking(ranking Ranking) [][]map[string]any {
	switch r := ranking.(type) {
	case TieableRanking:
		return m.marshalTieableRanking(r)
	default:
		ranks := r.Ranks()
		slotIds := make([][]map[string]any, len(ranks))
		for i, slot := range ranks {
			slotIds[i] = []map[string]any{m.marshalSlot(slot)}
		}
		return slotIds
	}
}

func (m *TournamentMarshaller) marshalTieableRanking(ranking TieableRanking) [][]map[string]any {
	ranks := ranking.TiedRanks()
	slotIds := make([][]map[string]any, len(ranks))
	for i, rank := range ranks {
		rankSlotIds := make([]map[string]any, len(rank))
		for i, slot := range rank {
			rankSlotIds[i] = m.marshalSlot(slot)
		}
		slotIds[i] = rankSlotIds
	}
	return slotIds
}

func (m *TournamentMarshaller) marshalTies(ties [][]*Slot) [][]map[string]any {
	marshalledTies := make([][]map[string]any, len(ties))
	for i, tie := range ties {
		slots := make([]map[string]any, len(tie))
		for i, slot := range tie {
			slots[i] = m.marshalSlot(slot)
		}
		marshalledTies[i] = slots
	}
	return marshalledTies
}

func (m *TournamentMarshaller) marshalMetrics(ranking *MatchMetricRanking) []*MatchMetrics {
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

func (m *TournamentMarshaller) marshalMatchList(matchList *matchList) map[string]any {
	rounds := make([][]*Match, len(matchList.Rounds))
	for i, r := range matchList.Rounds {
		rounds[i] = r.Matches
	}
	return m.marshalRoundList(rounds)
}

func (m *TournamentMarshaller) marshalRoundList(rounds [][]*Match) map[string]any {
	matches := make([][]string, len(rounds))
	for i, round := range rounds {
		roundMatches := make([]string, len(round))
		for i, match := range round {
			roundMatches[i] = m.idMap[match.id]
		}
		matches[i] = roundMatches
	}

	result := map[string]any{
		"rounds": matches,
	}

	return result
}

func (m *TournamentMarshaller) marshalMatch(match *Match) map[string]any {
	result := map[string]any{
		"slot1":    m.marshalSlot(match.Slot1),
		"slot2":    m.marshalSlot(match.Slot2),
		"walkover": match.IsWalkover(),
	}
	winner, err := match.GetWinner()
	if err == nil && winner.Player != nil {
		result["winner"] = winner.Player.Id()
	}
	return result
}

type editingPolicyAndMatchList interface {
	EditingPolicy
	MatchLister
}

func (m *TournamentMarshaller) marshalEditableMatches(editingPolicy editingPolicyAndMatchList) map[string]any {
	editable := editingPolicy.EditableMatches()
	editableIds := make([]string, len(editable))
	for i, match := range editable {
		editableIds[i] = m.idMap[match.id]
	}
	result := map[string]any{
		"editable": editableIds,
	}
	return result
}

func (m *TournamentMarshaller) marshalSingleElimination(tournament *SingleElimination) map[string]any {
	ranks := m.marshalEntriesAndFinal(tournament.Entries, tournament.FinalRanking)
	matchList := m.marshalMatchList(tournament.matchList)
	editable := m.marshalEditableMatches(tournament)
	result := map[string]any{
		"type": "SingleElimination",
	}

	maps.Copy(result, matchList)
	maps.Copy(result, editable)
	maps.Copy(result, ranks)

	return result
}

func (m *TournamentMarshaller) marshalRoundRobin(tournament *RoundRobin) map[string]any {
	ranks := m.marshalEntriesAndFinal(tournament.Entries, tournament.FinalRanking)
	matchList := m.marshalMatchList(tournament.matchList)
	editable := m.marshalEditableMatches(tournament)
	numUntied := tournament.FinalRanking.RequiredUntiedRanks
	ties := m.marshalTies(tournament.FinalRanking.BlockingTies(numUntied))
	unbrokenTies := m.marshalTies(tournament.FinalRanking.BlockingUnbrokenTies(numUntied))
	result := map[string]any{
		"type":         "RoundRobin",
		"metrics":      m.marshalMetrics(tournament.FinalRanking),
		"ties":         ties,
		"unbrokenTies": unbrokenTies,
	}

	maps.Copy(result, matchList)
	maps.Copy(result, editable)
	maps.Copy(result, ranks)

	return result
}

func (m *TournamentMarshaller) marshalSingleEliminationWithConsolation(tournament *SingleEliminationWithConsolation) map[string]any {
	ranks := m.marshalEntriesAndFinal(tournament.Entries, tournament.FinalRanking)
	mainBracket := m.marshalConsolationBracket(tournament.MainBracket)
	editable := m.marshalEditableMatches(tournament)
	result := map[string]any{
		"type":        "SingleEliminationWithConsolation",
		"mainBracket": mainBracket,
	}

	maps.Copy(result, editable)
	maps.Copy(result, ranks)

	return result
}

func (m *TournamentMarshaller) marshalConsolationBracket(bracket *ConsolationBracket) map[string]any {
	matchList := m.marshalMatchList(bracket.matchList)
	nested := make([]map[string]any, len(bracket.Consolations))
	for i, bracket := range bracket.Consolations {
		nested[i] = m.marshalConsolationBracket(bracket)
	}
	result := map[string]any{
		"consolations": nested,
	}
	maps.Copy(result, matchList)

	return result
}

func (m *TournamentMarshaller) marshalDoubleElimination(tournament *DoubleElimination) map[string]any {
	ranks := m.marshalEntriesAndFinal(tournament.Entries, tournament.FinalRanking)
	winnerMatchList := m.marshalMatchList(tournament.WinnerBracket.matchList)
	loserMatchList := m.marshalRoundList(tournament.loserRounds)
	editable := m.marshalEditableMatches(tournament)
	result := map[string]any{
		"type":         "DoubleElimination",
		"winnerRounds": winnerMatchList["rounds"],
		"loserRounds":  loserMatchList["rounds"],
		"final":        m.idMap[tournament.final.id],
	}

	maps.Copy(result, editable)
	maps.Copy(result, ranks)

	return result
}

func (m *TournamentMarshaller) marshalGroupPhase(tournament *GroupPhase) map[string]any {
	groups := make([]any, 0)
	for _, g := range tournament.Groups {
		matchList := m.marshalMatchList(g.matchList)
		ranks := m.marshalEntriesAndFinal(g.Entries, g.FinalRanking)
		numUntied := g.FinalRanking.RequiredUntiedRanks
		ties := m.marshalTies(g.FinalRanking.BlockingTies(numUntied))
		unbrokenTies := m.marshalTies(g.FinalRanking.BlockingUnbrokenTies(numUntied))
		groupResult := map[string]any{
			"type":         "GroupRoundRobin",
			"metrics":      m.marshalMetrics(g.FinalRanking),
			"ties":         ties,
			"unbrokenTies": unbrokenTies,
		}
		maps.Copy(groupResult, matchList)
		maps.Copy(groupResult, ranks)

		groups = append(groups, groupResult)
	}

	ties := tournament.FinalRanking.CrossGroupTies()
	unbrokenTies := tournament.FinalRanking.BlockingUnbrokenTies(tournament.FinalRanking.RequiredUntiedRanks)
	crossGroupTies := m.marshalTies(ties)
	unbrokenCrossGroupTies := m.marshalTies(unbrokenTies)
	crossTiedRank, _ := tournament.FinalRanking.contestedRank()

	result := map[string]any{
		"type":                   "GroupPhase",
		"groups":                 groups,
		"crossGroupTies":         crossGroupTies,
		"unbrokenCrossGroupTies": unbrokenCrossGroupTies,
		"crossTiedRank":          crossTiedRank,
	}

	return result
}

func (m *TournamentMarshaller) marshalGroupKnockout(tournament *GroupKnockout) map[string]any {
	ranks := m.marshalEntriesAndFinal(tournament.Entries, tournament.FinalRanking)
	groupPhase := m.marshalGroupPhase(tournament.GroupPhase)
	editable := m.marshalEditableMatches(tournament)
	var koPhase map[string]any
	switch ko := tournament.KnockOutTournament.(type) {
	case *SingleElimination:
		matchList := m.marshalMatchList(ko.matchList)
		ranks := m.marshalEntriesAndFinal(ko.Entries, ko.FinalRanking)
		koPhase = map[string]any{
			"type": "SingleElimination",
		}
		maps.Copy(koPhase, matchList)
		maps.Copy(koPhase, ranks)
	case *SingleEliminationWithConsolation:
		mainBracket := m.marshalConsolationBracket(ko.MainBracket)
		ranks := m.marshalEntriesAndFinal(ko.Entries, ko.FinalRanking)
		koPhase = map[string]any{
			"type":        "SingleEliminationWithConsolation",
			"mainBracket": mainBracket,
		}
		maps.Copy(koPhase, ranks)
	case *DoubleElimination:
		matchList := m.marshalMatchList(ko.matchList)
		ranks := m.marshalEntriesAndFinal(ko.Entries, ko.FinalRanking)
		koPhase = map[string]any{
			"type": "DoubleElimination",
		}
		maps.Copy(koPhase, matchList)
		maps.Copy(koPhase, ranks)
	default:
		panic("group knockout marshaller: unknown ko phase tournament type")
	}

	result := map[string]any{
		"type":       "GroupKnockout",
		"groupPhase": groupPhase,
		"koPhase":    koPhase,
		"koStarted":  tournament.KnockOut.matchList.MatchesStarted(),
	}

	maps.Copy(result, editable)
	maps.Copy(result, ranks)

	return result
}

func (m *Match) ToMap() map[string]any {
	marshaller := TournamentMarshaller{}
	return marshaller.marshalMatch(m)
}

func (t *SingleElimination) ToMap(getMatchId func(int) string) map[string]any {
	marshaller := newTournamentMarshaller(t, getMatchId)
	return marshaller.marshalSingleElimination(t)
}

func (t *SingleEliminationWithConsolation) ToMap(getMatchId func(int) string) map[string]any {
	marshaller := newTournamentMarshaller(t, getMatchId)
	return marshaller.marshalSingleEliminationWithConsolation(t)
}

func (t *RoundRobin) ToMap(getMatchId func(int) string) map[string]any {
	marshaller := newTournamentMarshaller(t, getMatchId)
	return marshaller.marshalRoundRobin(t)
}

func (t *GroupKnockout) ToMap(getMatchId func(int) string) map[string]any {
	marshaller := newTournamentMarshaller(t, getMatchId)
	return marshaller.marshalGroupKnockout(t)
}

func (t *DoubleElimination) ToMap(getMatchId func(int) string) map[string]any {
	marshaller := newTournamentMarshaller(t, getMatchId)
	return marshaller.marshalDoubleElimination(t)
}
