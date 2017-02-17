package main

import "testing"

func TestPickCard(t *testing.T) {
	test := new(game)
	test.trump = "K♥"
	test.closedBy = Player1
	test.trick[Player1] = "Q♥"
	test.hands[Player2] = []string{"J♥", "Q♦", "A♥", "A♦"}
	if test.pickCard() != 3 {
		t.Error("Pick card error!")
	}

	test.trump = "K♦"
	test.closedBy = Nobody
	test.trick[Player1] = "Q♥"
	test.hands[Player2] = []string{"J♥", "Q♠", "9♥", "A♠"}
	if test.pickCard() != 0 {
		t.Error("Pick card error!", test.pickCard())
	}
}

func TestFindLowestRank(t *testing.T) {
	test := new(game)
	test.hands[Player2] = []string{"J♥", "Q♠", "9♥", "A♦"}
	if test.findLowestRank() != 3 {
		t.Error("Lowest rank error!")
	}
}
