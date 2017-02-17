package main

import "testing"

var (
	test  = new(game)
	hand  = []string{"Q♥", "9♥", "K♥", "X♠"}
	trump = "A♥"
)

func TestStart(t *testing.T) {
	test.start()

	if len(test.hands[Player1]) != len(test.hands[Player2]) || len(test.hands[Player1]) != 6 {
		t.Error("Hands' size is wrong!")
	}

	for player := 0; player < 2; player++ {
		for _, card := range test.hands[player] {
			if card == test.trump {
				t.Error("Nobody should have the trump from the beginning!")
			}
		}
	}

	if len(test.deck.Current)+1 != 12 {
		t.Error("Deck size not right!")
	}

	if test.playerInTurn != Player2 || test.playerNotInTurn() != Player1 {
		t.Error("Not the right player in turn!")
	}
}

func TestCheckForMarriage(t *testing.T) {
	test.hands[Player1] = hand
	test.trump = trump

	if ok, pts := test.checkForMarriage(Player1, "K♥"); !ok || pts != 40 {
		t.Error("Marriage error!")
	}

	if ok, pts := test.checkForMarriage(Player1, "9♥"); ok || pts != 0 {
		t.Error("Marriage error!")
	}
}

func TestIsPossibleExchange(t *testing.T) {
	test.playerInTurn = Player2
	test.trick[Player1] = NoCard
	test.hands[Player2] = hand
	test.trump = trump

	if ok, _ := test.isPossibleExchange(Player2); ok {
		t.Error("Exchange error!")
	}

	test.hasTrickWon[Player2] = true
	if ok, _ := test.isPossibleExchange(Player2); !ok {
		t.Error("Exchange error!")
	}

	test.hands[Player2][1] = "J♠"
	if ok, _ := test.isPossibleExchange(Player2); ok {
		t.Error("Exchange error!")
	}
}

func TestHasSameSuit(t *testing.T) {
	test.hands[Player1] = hand
	if !test.hasSameSuit(Player1, "J♥") {
		t.Error("Same suit error!")
	}
	if test.hasSameSuit(Player1, "J♦") {
		t.Error("Same suit error!")
	}
}

func TestHasSameSuitHigher(t *testing.T) {
	test.hands[Player1] = hand
	if !test.hasSameSuitHigher(Player1, "J♥") {
		t.Error("Same suit error!")
	}
	if test.hasSameSuitHigher(Player1, "J♦") {
		t.Error("Same suit error!")
	}
}

func TestHasTrump(t *testing.T) {
	test.hands[Player1] = hand
	if !test.hasTrump(Player1) {
		t.Error("Trump error!")
	}

	test.trump = "J♦"
	if test.hasTrump(Player1) {
		t.Error("Trump error!")
	}
}

func TestFindWinner(t *testing.T) {
	test.playerInTurn = Player2
	test.trump = "J♠"
	test.trick[Player1] = "J♥"
	test.trick[Player2] = "J♦"
	if test.findWinner() != Player1 {
		t.Error("Trick win error!")
	}

	test.trump = "Q♦"
	if test.findWinner() != Player2 {
		t.Error("Trick win error!")
	}

	test.trick[Player2] = "Q♥"
	if test.findWinner() != Player2 {
		t.Error("Trick win error!")
	}
}

func TestEndDeal(t *testing.T) {
	test.dealScore[Player1] = 67
	test.dealScore[Player2] = 47
	test.hasTrickWon[Player1] = true
	test.hasTrickWon[Player2] = true
	test.closedBy = Nobody

	if player, pts := test.endDeal(Nobody); player != Player1 || pts != 1 {
		t.Error("End deal error!")
	}

	test.dealScore[Player1] = 65
	test.dealScore[Player2] = 47
	test.hasTrickWon[Player1] = true
	test.hasTrickWon[Player2] = true
	if player, pts := test.endDeal(Player1); player != Player2 || pts != 2 {
		t.Error("End deal error!")
	}

	test.dealScore[Player1] = 65
	test.dealScore[Player2] = 0
	test.hasTrickWon[Player1] = true
	if player, pts := test.endDeal(Nobody); player != Player1 || pts != 3 {
		t.Error("End deal error!")
	}
}

func TestDraw(t *testing.T) {
	copy(test.hands[Player1], hand)
	copy(test.hands[Player2], hand)
	test.emptyCardSlots[Player1] = 4
	test.emptyCardSlots[Player2] = 4

	test.deck.Shuffle()
	test.draw()
	if len(test.hands[Player1]) != len(test.hands[Player2]) ||
		len(test.deck.Current) != 22 || test.hands[Player1][4] == test.hands[Player2][4] {
		t.Error("Eror in drawing.")
	}

	test.trump = test.deck.Current[0]
	test.deck.Current = test.deck.Current[21:]
	test.playerInTurn = Player1
	test.emptyCardSlots[Player1] = 5
	test.emptyCardSlots[Player2] = 5
	test.draw()
	if len(test.deck.Current) != 0 || test.hands[Player2][5] != test.trump {
		t.Error("Eror in drawing.")
	}

	test.draw()
	if len(test.deck.Current) != 0 || len(test.hands[Player2]) != 5 {
		t.Error("Eror in drawing.")
	}
}
