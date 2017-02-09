package deck

import (
	"testing"
)

func TestShuffle(t *testing.T) {
	d := New()
	old := make([]string, Size)
	copy(old, d.Initial)
	d.Shuffle()
	for i := 0; i < Size; i++ {
		if old[i] != d.Initial[i] {
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
