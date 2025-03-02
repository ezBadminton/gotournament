package core

import "slices"

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

	// Lists the matches that a player would be withdrawn from
	// if WithdrawPlayer was called
	ListWithdrawMatches(player Player) []*Match

	// Lists the matches that a player would reenter into
	// if ReenterPlayer was called
	ListReenterMatches(player Player) []*Match
}

func withdrawFromMatches(player Player, withdrawMatches []*Match) {
	for _, m := range withdrawMatches {
		m.WithdrawnPlayers = append(m.WithdrawnPlayers, player)
	}
}

func reenterIntoMatches(player Player, reenterMatches []*Match) {
	for _, m := range reenterMatches {
		m.WithdrawnPlayers = slices.DeleteFunc(m.WithdrawnPlayers, func(p Player) bool { return p.Id() == player.Id() })
	}
}
