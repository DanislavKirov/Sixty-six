package deck

import "testing"

func TestShuffle(t *testing.T) {
	d := New()
	d.Shuffle()

	d.DrawCard()
	d.Shuffle()

	if len(d.Current) != Size {
		t.Error("Size error!")
	}

	for i := 0; i < Size; i++ {
		if OrderedDeck[i] != d.Initial[i] {
			return
		}
	}
	t.Error("Didn't shuffle!")
}

func TestDrowCard(t *testing.T) {
	d := New()
	d.Shuffle()

	for i := 0; i < Size; i++ {
		_, err := d.DrawCard()
		if err != nil {
			t.Error("Empty deck!")
		}
		if len(d.Initial) == len(d.Current) {
			t.Error("Initial deck must not change!")
		}
	}

	_, err := d.DrawCard()
	if err == nil {
		t.Error("Deck should be empty!")
	}
}

func TestDrowNcards(t *testing.T) {
	deck := New()
	deck.Shuffle()

	cards, err := deck.DrawNcards(Size / 2)
	if err != nil || len(cards) != Size/2 {
		t.Error("Deck should have enough cards!")
	}

	cards, err = deck.DrawNcards(Size/2 + 1)
	if err == nil || len(cards) != 0 {
		t.Error("Deck shouldn't have enough cards!")
	}

	cards, err = deck.DrawNcards(Size / 2)
	if err != nil || len(cards) != Size/2 {
		t.Error("Deck should have enough cards!")
	}

	cards, err = deck.DrawNcards(1)
	if err == nil || len(cards) != 0 {
		t.Error("Deck shouldn't have enough cards!")
	}
}

func TestHasHigherRank(t *testing.T) {
	if HasHigherRank(OrderedDeck[0], OrderedDeck[4]) || // idx: 0,1,2,3 -> 9s
		!HasHigherRank(OrderedDeck[8], OrderedDeck[4]) { // idx: 4->J, 8->Q
		t.Error("Rank error!")
	}
}

func TestAreTheSameSuit(t *testing.T) {
	if AreTheSameSuit(OrderedDeck[0], OrderedDeck[1]) ||
		!AreTheSameSuit(OrderedDeck[0], OrderedDeck[4]) {
		t.Error("Suit error!")
	}
}
