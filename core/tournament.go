package core

import "errors"

var (
	ErrTooFewEntries = errors.New("not enough entries for this tournament mode")
)

// A slice of matches and a slice of rounds containing all
// matches.
type matchList struct {
	Matches []*Match
	Rounds  []*Round
}

func (l *matchList) MatchesOfPlayer(player Player) []*Match {
	matches := make([]*Match, 0, 5)
	for _, m := range l.Matches {
		if m.HasDrawnBye() {
			continue
		}

		if m.ContainsPlayer(player) {
			matches = append(matches, m)
		}
	}

	return matches
}

// Returns true when all matches in the list are complete
func (l *matchList) MatchesComplete() bool {
	for _, m := range l.Matches {
		if !m.HasBye() && !m.IsWalkover() && m.Score == nil {
			return false
		}
	}
	return true
}

// Returns true when any of the matches have started
func (l *matchList) MatchesStarted() bool {
	for _, m := range l.Matches {
		if !m.StartTime.IsZero() {
			return true
		}
	}
	return false
}

type EditingPolicy interface {
	// Returns the comprehensive list of matches that are editable
	EditableMatches() []*Match

	// Updates the return value of EditableMatches
	updateEditableMatches()
}

type RankingUpdater interface {
	// Updates all rankings and slots going
	// from the start Ranking in the
	// dependecy graph
	Update(start Ranking)
}

type MatchLister interface {
	MatchList() *matchList
}

type BaseTournament[FinalRanking Ranking] struct {
	// The entries ranking which contains
	// the starting slots for all participants.
	Entries Ranking
	*matchList
	*RankingGraph
	FinalRanking FinalRanking

	WithdrawalPolicy
	EditingPolicy

	id int
}

func (t *BaseTournament[_]) Update(start Ranking) {
	if start == nil {
		start = t.Entries
	}

	bfs := t.RankingGraph.BreadthSearchIter(start)
	for ranking := range bfs {
		ranking.updateRanks()
		for _, s := range ranking.dependantSlots() {
			s.Update()
		}
	}

	t.EditingPolicy.updateEditableMatches()
}

func (t *BaseTournament[_]) Id() int {
	return t.id
}

func (t *BaseTournament[_]) MatchList() *matchList {
	return t.matchList
}

func (t *BaseTournament[FinalRanking]) addTournamentData(
	matchList *matchList,
	rankingGraph *RankingGraph,
	finalRanking FinalRanking,
) {
	t.matchList = matchList
	t.RankingGraph = rankingGraph
	t.FinalRanking = finalRanking
}

func (t *BaseTournament[_]) addPolicies(
	editingPolicy EditingPolicy,
	withdrawalPolicy WithdrawalPolicy,
) {
	t.EditingPolicy = editingPolicy
	t.WithdrawalPolicy = withdrawalPolicy
}

func newBaseTournament[FinalRanking Ranking](entries Ranking) BaseTournament[FinalRanking] {
	tournament := BaseTournament[FinalRanking]{
		Entries: entries,
		id:      NextId(),
	}
	return tournament
}
