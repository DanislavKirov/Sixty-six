package main

import "testing"

var g = new(Game)

func TestStart(t *testing.T) {
	g.Start()

	if len(g.hands[Player1]) != len(g.hands[Player2]) || len(g.hands[Player1]) != 6 {
		t.Error("Hands' size is wrong!")
	}

	for player := 0; player < 2; player++ {
		for _, card := range g.hands[player] {
			if card == g.trump {
				t.Error("Nobody should have the trump from the beginning!")
			}
		}
	}

	if g.DeckSize() != 12 {
		t.Error("Deck size not right!")
	}

	if g.PlayerInTurn() != Player2 || g.PlayerNotInTurn() != Player1 {
		t.Error("Not the right player in turn!")
	}
}
