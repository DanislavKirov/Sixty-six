package deck_test

import (
	"testing"

	"github.com/DanislavKirov/Sixty-six/src/deck"
)

func TestShuffle(t *testing.T) {
	d := deck.New()
	old := make([]string, deck.Size)
	copy(old, d.Initial)
	d.Shuffle()
	for i := 0; i < deck.Size; i++ {
		if old[i] != d.Initial[i] {
			return
		}
	}
	t.Error("Didn't shuffle!")
}

func TestDrowCard(t *testing.T) {
	d := deck.New()

	for i := 0; i < deck.Size; i++ {
		_, err := d.DrawCard()
		if err != nil {
			t.Error("Empty deck!")
		}
	}

	_, err := d.DrawCard()
	if err == nil {
		t.Error("Deck should be empty!")
	}
}
