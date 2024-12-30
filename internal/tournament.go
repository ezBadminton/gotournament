package internal

// A slice of matches and a slice of rounds containing all
// matches.
type MatchList struct {
	Matches []*Match
	Rounds  []*Round
}

// A MatchMaker initializes a Tournament by defining the matches
// that are to be played.
type MatchMaker interface {
	// Creates a MatchList, a Ranking Graph and the final ranking
	//
	// The participating players are passed as a Ranking
	// of entries as well as some arbitrary tournament mode
	// specific settings.
	//
	// Given the same entries and settings, this method always
	// returns the same slice of rounds. Any RNG values are
	// seeded by a value from the settings.
	//
	// Can return an error when the ranking is empty or
	// invalid settings are passed.
	MakeMatches(entries Ranking, settings interface{}) (*MatchList, *RankingGraph, Ranking, error)
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
	Update()
}

// A Tournament is a chain of matches and rankings.
//
// Every tournament begins with a ranking of entries and
// ends with a final ranking. What comes in between depends
// on the tournament mode that is implemented.
type Tournament interface {
	// Update the tournament's slots and rankings.
	Update(start Ranking)
}

type BaseTournament struct {
	// The entries ranking which contains
	// the starting slots for all participants.
	Entries Ranking
	// The final ranking is the overall result
	// of the entire tournament.
	// It should contain a slot for every player
	// who is in the Entries.
	FinalRanking Ranking

	MatchMaker       MatchMaker
	RankingGraph     *RankingGraph
	MatchList        *MatchList
	WithdrawalPolicy WithdrawalPolicy
	EditingPolicy    EditingPolicy
}

func (t *BaseTournament) Update(start Ranking) {
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
}

func (t *BaseTournament) MatchesOfPlayer(player Player) []*Match {
	matches := make([]*Match, 0, 5)
	for _, m := range t.MatchList.Matches {
		if m.HasDrawnBye() {
			continue
		}

		if m.ContainsPlayer(player) {
			matches = append(matches, m)
		}
	}

	return matches
}
