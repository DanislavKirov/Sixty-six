// Package deck provides api for creating and using a deck of 24 cards.
package deck

import (
	"errors"
	"math/rand"
	"time"
)

// Size is the number of cards in a deck.
const Size = 24

var (
	suits  = [4]string{"♣", "♦", "♥", "♠"}
	cards  = [6]string{"9", "J", "Q", "K", "X", "A"} // X == 10
	points = [6]int{0, 2, 3, 4, 10, 11}
)

// OrderedDeck is the initial full ordered deck.
// Points is a map connecting each card with its points.
var (
	OrderedDeck []string
	Points      map[byte]int
)

// init initializes the exported OrderedDeck and Points.
func init() {
	OrderedDeck = make([]string, Size)
	Points = make(map[byte]int)
	i := 0
	for idx, card := range cards {
		for _, suit := range suits {
			OrderedDeck[i] = card + suit
			i++
		}
		Points[card[0]] = points[idx]
	}
}

// Deck contains the original deck and the current one (after drawing cards).
type Deck struct {
	Initial []string
	Current []string
}

// New creates and returns an ordered deck of cards.
func New() *Deck {
	deck := new(Deck)

	deck.Initial = make([]string, Size)
	deck.Current = make([]string, Size)
	copy(deck.Initial, OrderedDeck)
	copy(deck.Current, deck.Initial)

	return deck
}

// Shuffle shuffles the Initial deck and makes a copy of it in Current.
func (d *Deck) Shuffle() {
	res := make([]string, Size)
	rand.Seed(time.Now().UTC().UnixNano())
	perm := rand.Perm(Size)
	for i, v := range perm {
		res[v] = d.Initial[i]
	}

	copy(d.Initial, res)
	if len(d.Current) < Size {
		d.Current = make([]string, Size)
	}
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

// DrawNcards returns the top n cards if n <= the size of the current deck.
func (d *Deck) DrawNcards(n int) ([]string, error) {
	if n > len(d.Current) {
		return nil, errors.New("Not enough cards in deck")
	}

	cards := d.Current[:n]
	d.Current = d.Current[n:]
	return cards, nil
}

// AreTheSameSuit returns true if card1 and card2 are from the same suit.
func AreTheSameSuit(card1, card2 string) bool {
	return card1[1:] == card2[1:]
}

// HasHigherRank returns true if the rank of card1 is higher than the rank of card2.
func HasHigherRank(card1, card2 string) bool {
	return Points[card1[0]] > Points[card2[0]]
}
