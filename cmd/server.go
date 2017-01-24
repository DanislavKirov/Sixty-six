// This file contains the server of the game.
// The server is responsible for making sure that the game is played by the rules.
package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
)

// game contains the connections and deck info.
type game struct {
	connectedPlayers int
	players          [2]net.Conn
	turn             int

	deck  *deck.Deck
	trump string
	hands [2][]string

	currentDealScore [2]int
	gameScore        [2]int
}

// deal deals the first cards.
func (g *game) deal() {
	g.hands[player1] = make([]string, 6)
	g.hands[player2] = make([]string, 6)
	hands, _ := g.deck.DrawNcards(13)
	copy(g.hands[player2][:3], hands[:3]) // player1 deals, so player2 gets the first cards
	copy(g.hands[player2][3:], hands[6:9])
	copy(g.hands[player1][:3], hands[3:6])
	copy(g.hands[player1][3:], hands[9:12])
	g.trump = hands[12]
	g.turn = player2
}

// sendTo sends a message to player.
func sendTo(player int, message string) {
	g.players[player].Write([]byte(message))
}

// broadcast sends a message to both players.
func broadcast(message string) {
	sendTo(player1, message)
	sendTo(player2, message)
}

func sendHands() {
	hand := strings.Join(g.hands[player1], " ")
	sendTo(player1, "Your hand: "+hand)
	hand = strings.Join(g.hands[player2], " ")
	sendTo(player2, "Your hand: "+hand)
}

func sendTurn() {
	sendTo(g.turn, yourTurn)
	sendTo(1-g.turn, notYourTurn)
}

// playerConnected sends suitable message to the players when someone connects.
func playerConnected() {
	if g.connectedPlayers == 1 {
		sendTo(player1, "Waiting for the other player to connect.")
	} else {
		broadcast("The game starts now.")
		startGame()
	}
}

// startGame creates a deck and deals the first cards.
func startGame() {
	g.deck = deck.New()
	g.deck.Shuffle()
	g.deal()
	sendHands()
	broadcast("Trump: " + g.trump)
	sendTurn()
}

func unexpectedExit() {
	g.players[player1].Close()
	if g.connectedPlayers == 2 {
		broadcast("Opponent left.")
		g.players[player2].Close()
	}
}

func listenToPlayer(player int) {
	p := make([]byte, 64)
	for {
		_, err := g.players[player].Read(p)
		if err != nil {
			fmt.Println(err.Error())
			unexpectedExit()
			return
		}
		m := string(p)
		cardIdx := int(m[0]-'0') - 1
		if cardIdx >= 0 && cardIdx <= len(g.hands[player]) {
			sendTo(1-player, "Opponent's card: "+g.hands[player][cardIdx])
		} else {
			sendTo(player, "wrong input, try again: ")
		}
	}
}

var (
	g      = new(game)
	server net.Listener
)

// startServer starts a server.
func startServer() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	server = l
	//	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		p := make([]byte, 64)
		_, e := conn.Read(p)
		if e != nil {
			log.Fatal(err)
		}
		if string(p[:7]) == Connect {
			if g.connectedPlayers == 0 {
				g.players[player1] = conn
				go listenToPlayer(player1)
			} else if g.connectedPlayers == 1 {
				g.players[player2] = conn
				go listenToPlayer(player2)
			} else {
				conn.Write([]byte("Already enough players."))
				break
			}
			g.connectedPlayers += 1
			playerConnected()
		}
	}
}
