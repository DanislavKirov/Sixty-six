package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
)

// pickCard returns index of card which can win the trick.
// It returns 0 if bot can't win with any card.
func (g *game) pickCard() int {
	for idx, card := range g.hands[Player2] {
		if g.isBetter(card, g.trick[Player1]) {
			return idx + 1
		}
	}
	return 0
}

// isBetter returns true if card1 wins.
func (g *game) isBetter(card1, card2 string) bool {
	if (g.isTrump(card1) && !g.isTrump(card2)) ||
		(deck.AreTheSameSuit(card1, card2) && deck.HasHigherRank(card1, card2)) {
		return true
	}
	return false
}

// findLowestRank returns the index of the lowest rank card.
func (g *game) findLowestRank() int {
	idx, rank := 1, "A"[0]
	for i, card := range g.hands[Player2] {
		if deck.Points[card[Rank]] < deck.Points[rank] {
			rank = card[Rank]
			idx = i
		}
	}

	return idx + 1
}

// startBot creates and manages the bot.
func startBot(ip string) {
	connection, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection.Write([]byte(Connect))
	buff := make([]byte, 256)

	var cardIdx int
	for {
		size, err := connection.Read(buff)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
		message := string(buff)[:size]

		if strings.Contains(message, YourTurn) {
			if g.dealScore[Player2] >= 66 {
				connection.Write([]byte(Stop))
				continue
			}

			if g.trick[Player1] != NoCard {
				cardIdx = g.pickCard()
				if cardIdx == 0 {
					cardIdx = g.findLowestRank()
				}
			} else {
				cardIdx = rand.Intn(len(g.hands[Player2])) + 1
			}

			connection.Write([]byte(strconv.Itoa(cardIdx) + "\n"))
		} else if message == WrongInput || message == NotPossible {
			for idx, card := range g.hands[Player2] {
				if g.isGoodResponse(Player2, card) {
					connection.Write([]byte(strconv.Itoa(idx+1) + "\n"))
					break
				}
			}
		}
	}
}
