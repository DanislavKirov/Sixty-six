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

// New creates and returns an ordered deck of cards.
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

// Shuffle takes a full deck and returns a shuffled one.
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

// DrawCard returns the top card of the deck if it has any.
func (d *Deck) DrawCard() (string, error) {
	card, err := d.DrawNcards(1)
	if err != nil {
		return "", err
	}
	return card[0], err
}

// DrawNcards returns the current deck and an error if n is bigger than the cards in the deck.
// Otherwise it returns the top n cards.
func (d *Deck) DrawNcards(n int) ([]string, error) {
	if n > len(d.Current) {
		d.Current = d.Current[len(d.Current)-1:]
		return d.Current, errors.New("Not enough cards in deck")
	}

	cards := d.Current[:n]
	d.Current = d.Current[n:]
	return cards, nil
}
