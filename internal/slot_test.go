package internal

import (
	"errors"
)

type TestPlayer struct {
	id string
}

func (p *TestPlayer) Id() string {
	return p.id
}

var _ Player = &TestPlayer{}

var testIds string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var nextTestId int = 0

func NewTestPlayer() (*TestPlayer, error) {
	if nextTestId >= len(testIds) {
		return nil, errors.New("Max number of test players exceeded")
	}
	id := string(testIds[nextTestId])
	nextTestId += 1

	return &TestPlayer{id: id}, nil
}

func PlayerSlice(num int) ([]Player, error) {
	nextTestId = 0
	players := make([]Player, 0, num)
	for range num {
		player, err := NewTestPlayer()
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}
