package core

import "maps"

// The crossGroupMatchMetricSource provides match metrics which are comparable between the
// groups of a GroupPhase. In case of groups with equal sizes the metrics are unaltered
// and the same as the baseMatchMetricSource would return.
//
// When the groups are of unequal sizes because the total amount of teams is not divisible
// by the number of groups, then the metrics are altered as follows:
//
//   - Each match that involves the last placed team of a group with more players
//     is counted as a walkover win for the opponent of the last placed team
//   - The last placed teams do not get their metrics altered
//   - Each team from the groups with fewer players gets an additional walkover win
//     (against a non-existent opponent) added to their metrics
//
// This way every team has the same amount of matches.
//
// The relevant tournament configuration for this to matter is this:
// When the number of qualifications is not divisble by the number of groups, it leads
// to one rank having to be compared across all groups to determine who qualifies
// (e.g. the best of all 3rd places).
// If the number of teams is additionally not divisible by the number of groups, then
// the problem arises that the teams played a different number of matches at the
// end of the group phase depending on the number of teams in their groups.
// That means their match metrics are not directly comparable.
// This algorithm is (until somebody finds a better way) the only way to fairly
// balance out the metrics.
//
// Tournament administrators would be be well-advised to not configure their group
// tournament in a way where neither the number of teams nor the number of qualifications
// is divisible by the number of groups, however this code enables them to do so if they choose.
type crossGroupMatchMetricSource struct {
	baseMatchMetricSource

	groups []*RoundRobin
}

func (s *crossGroupMatchMetricSource) CreateMetrics(
	players []Player,
) map[Player]*MatchMetrics {
	groupsFinished := true
	for _, group := range s.groups {
		requiredUntied := group.FinalRanking.RequiredUntiedRanks
		blockingTies := group.FinalRanking.BlockingTies(requiredUntied)
		if len(blockingTies) != 0 {
			groupsFinished = false
			break
		}
	}
	if !groupsFinished {
		return s.baseMatchMetricSource.CreateMetrics(players)
	}

	numGroups := len(s.groups)
	// The lower index group is always the one with the fewer players (if a difference exists at all)
	minGroupSize := len(s.groups[0].Entries.Ranks())
	maxGroupSize := len(s.groups[numGroups-1].Entries.Ranks())

	if minGroupSize == maxGroupSize {
		return s.baseMatchMetricSource.CreateMetrics(players)
	}

	largeGroups := make([]*RoundRobin, 0)
	smallGroups := make([]*RoundRobin, 0)
	for _, group := range s.groups {
		groupSize := len(group.Entries.Ranks())
		if groupSize == minGroupSize {
			smallGroups = append(smallGroups, group)
		} else {
			largeGroups = append(largeGroups, group)
		}
	}

	metrics := make(map[Player]*MatchMetrics)

	walkoverPoints := s.walkoverScore.Points1()
	walkoverSetWins := len(walkoverPoints)
	walkoverPointWins := 0
	for _, setPoints := range walkoverPoints {
		walkoverPointWins += setPoints
	}

	walkoverMetrics := &MatchMetrics{
		NumMatches: 1,
		Wins:       1,
		NumSets:    walkoverSetWins,
		SetWins:    walkoverSetWins,
		PointWins:  walkoverPointWins,
	}

	for _, group := range largeGroups {
		lastPlaced := getLastOfGroup(group)
		matchesWithLast, matchesWithoutLast := separateMatchesByPlayer(group.Matches, lastPlaced)
		metricsWithLast := make(map[Player]*MatchMetrics)
		metricsWithoutLast := make(map[Player]*MatchMetrics)
		s.extractMatchMetricsFromSlice(matchesWithLast, nil, metricsWithLast)
		s.extractMatchMetricsFromSlice(matchesWithoutLast, nil, metricsWithoutLast)
		for _, metrics := range metricsWithoutLast {
			metrics.Add(walkoverMetrics)
		}
		maps.DeleteFunc(metricsWithLast, func(player Player, _ *MatchMetrics) bool {
			return player.Id() != lastPlaced.Id()
		})
		maps.Copy(metrics, metricsWithLast)
		maps.Copy(metrics, metricsWithoutLast)
	}

	for _, group := range smallGroups {
		groupMetrics := make(map[Player]*MatchMetrics)
		s.extractMatchMetricsFromSlice(group.Matches, nil, groupMetrics)
		for _, metrics := range groupMetrics {
			metrics.Add(walkoverMetrics)
		}
		maps.Copy(metrics, groupMetrics)
	}

	return metrics
}

func getLastOfGroup(group *RoundRobin) Player {
	ranks := group.FinalRanking.Ranks()
	numRanks := len(ranks)
	lastRank := ranks[numRanks-1]
	return lastRank.Player
}

// Separates the given slice of matches into those with
// and those without the given player involved
func separateMatchesByPlayer(matches []*Match, player Player) ([]*Match, []*Match) {
	withPlayer := make([]*Match, 0)
	withoutPlayer := make([]*Match, 0)

	for _, match := range matches {
		if match.ContainsPlayer(player) {
			withPlayer = append(withPlayer, match)
		} else {
			withoutPlayer = append(withoutPlayer, match)
		}
	}

	return withPlayer, withoutPlayer
}
