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
func pickCard() int {
	for idx, card := range game.hands[Player2] {
		if isBetter(card, game.trick[Player1]) {
			return idx + 1
		}
	}
	return 0
}

// isBetter returns true if card1 wins.
func isBetter(card1, card2 string) bool {
	if (game.IsTrump(card1) && !game.IsTrump(card2)) ||
		(deck.AreTheSameSuit(card1, card2) && deck.HasHigherRank(card1, card2)) {
		return true
	}
	return false
}

func findLowestRank() int {
	idx, rank := 1, "A"[0]
	for i, card := range game.hands[Player2] {
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
		fmt.Println(err.Error())
		return
	}

	connection.Write([]byte(Connect))
	buff := make([]byte, 256)

	var cardIdx int
	for {
		size, err := connection.Read(buff)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err.Error())
			}
			return
		}
		message := string(buff)[:size]

		if strings.Contains(message, YourTurn) {
			if game.dealScore[Player2] >= 66 {
				connection.Write([]byte(Stop))
				continue
			}

			if game.trick[Player1] != NoCard {
				cardIdx = pickCard()
				if cardIdx == 0 {
					cardIdx = findLowestRank()
				}
			} else {
				cardIdx = rand.Intn(len(game.hands[Player2])) + 1
			}

			connection.Write([]byte(strconv.Itoa(cardIdx) + "\n"))
		} else if message == WrongInput || message == NotPossible {
			for idx, card := range game.hands[Player2] {
				if game.isGoodResponse(Player2, card) {
					connection.Write([]byte(strconv.Itoa(idx+1) + "\n"))
					break
				}
			}
		}
	}
}
