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
