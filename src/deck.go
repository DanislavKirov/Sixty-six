// Package deck provides api for creating and using a deck of cards.
package deck

import (
	"errors"
	"math/rand"
	"time"
)

var (
	suits  = [4]string{"Clubs", "Diamonds", "Hearts", "Spades"}
	cards  = [6]string{"Nine", "Jack", "Queen", "King", "Ten", "Ace"}
	points = [6]int{0, 2, 3, 4, 10, 11}
)

// Size is the number of cards in a deck.
const Size = 24

// Deck contains the original deck and the current one (after drawing cards).
type Deck struct {
	Initial []string
	Current []string
}

// New creates and returns a shuffled deck of cards.
func New() *Deck {
	deck := new(Deck)
	deck.Initial = make([]string, Size)
	i := 0
	for _, card := range cards {
		for _, suit := range suits {
			deck.Initial[i] = card + " of " + suit
			i++
		}
	}

	deck.Current = make([]string, Size)
	copy(deck.Current, deck.Initial)

	return deck
}

// Shuffle takes a deck and returns a shuffled one with the same cards in it.
func (d *Deck) Shuffle() {
	res := make([]string, Size)
	rand.Seed(time.Now().UTC().UnixNano())
	perm := rand.Perm(Size)
	for i, v := range perm {
		res[v] = d.Initial[i]
	}

	copy(d.Initial, res)
	copy(d.Current, res)
}

// DrawCard takes a deck and returns the top card if it has any.
func (d *Deck) DrawCard() (string, error) {
	if len(d.Current) == 0 {
		return "", errors.New("Empty deck")
	}

	card := d.Current[0]
	d.Current = d.Current[1:]

	return card, nil
}
