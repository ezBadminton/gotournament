package internal

// A slice of matches and a slice of rounds containing all
// matches.
type MatchList struct {
	Matches []*Match
	Rounds  []*Round
}

func (l *MatchList) MatchesOfPlayer(player Player) []*Match {
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
func (l *MatchList) MatchesComplete() bool {
	for _, m := range l.Matches {
		if !m.HasBye() && !m.IsWalkover() && m.Score == nil {
			return false
		}
	}
	return true
}

// Returns true when any of the matches have started
func (l *MatchList) MatchesStarted() bool {
	for _, m := range l.Matches {
		if !m.StartTime.IsZero() {
			return true
		}
	}
	return false
}

// The WithdrawalPolicy dictates how a player can
// withdraw from a tournament and also if a player
// would be allowed to reenter.
type WithdrawalPolicy interface {
	// Withdraws the given player from the tournament.
	// The specific matches that the player was withdrawn from
	// are returned.
	WithdrawPlayer(player Player) []*Match

	// Attempts to reenter the player into the tournament.
	// On success the specific matches that the player
	// was reentered into are returned.
	ReenterPlayer(player Player) []*Match
}

type EditingPolicy interface {
	// Returns the comprehensive list of matches that are editable
	EditableMatches() []*Match

	// Updates the return value of EditableMatches
	UpdateEditableMatches()
}

// A Tournament is a chain of matches and rankings.
//
// Every tournament begins with a ranking of entries and
// ends with a final ranking. What comes in between depends
// on the tournament mode that is implemented.
type Tournament interface {
	// Update the tournament's slots and rankings.
	Update(start Ranking)

	GetMatchList() *MatchList
}

type BaseTournament[FinalRanking Ranking] struct {
	// The entries ranking which contains
	// the starting slots for all participants.
	Entries Ranking
	*MatchList
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
		ranking.UpdateRanks()
		for _, s := range ranking.DependantSlots() {
			s.Update()
		}
	}

	t.EditingPolicy.UpdateEditableMatches()
}

func (t *BaseTournament[_]) GetMatchList() *MatchList {
	return t.MatchList
}

func (t *BaseTournament[_]) Id() int {
	return t.id
}

func (t *BaseTournament[FinalRanking]) addTournamentData(
	matchList *MatchList,
	rankingGraph *RankingGraph,
	finalRanking FinalRanking,
) {
	t.MatchList = matchList
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

func NewBaseTournament[FinalRanking Ranking](entries Ranking) BaseTournament[FinalRanking] {
	tournament := BaseTournament[FinalRanking]{
		Entries: entries,
		id:      NextNodeId(),
	}
	return tournament
}
