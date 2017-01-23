// This file contains the server of the game.
// The server is responsible for making sure that the game is played by the rules.
package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
	"github.com/gorilla/websocket"
)

// game contains the connections and deck info.
type game struct {
	connectedPlayers int
	players          [2]*websocket.Conn
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

var (
	g = new(game)

//	requests = []string{"connect", "card", "close", "change", "end"}
)

const (
	player1     = 0
	player2     = 1
	connect     = "connect"
	yourTurn    = "It's your turn, pick a card index: "
	notYourTurn = "It's your opponent's turn, please wait."
)

// sendTo sends a message to player.
func sendTo(player int, message string) {
	g.players[player].WriteMessage(websocket.TextMessage, []byte(message))
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
func playerConnected(w http.ResponseWriter) {
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
	broadcast("Opponent left.")
	g.players[player1].Close()
	g.players[player2].Close()
}

func listenToPlayer(player int) {
	for {
		_, message, err := g.players[player].ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			unexpectedExit()
			return
		}
		m := string(message)
		fmt.Println(m)
		if len(m) == 2 && m[0] >= '1' && int(m[0]-'0') <= len(g.hands[player]) {
			sendTo(1-player, "Opponent's card: "+g.hands[player][m[0]-'1'])
		} else {
			sendTo(player, "wrong input, try again: ")
		}
	}
}

// handler handles the connections.
func handler(w http.ResponseWriter, r *http.Request) {
	if g.connectedPlayers < 2 {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  16,
			WriteBufferSize: 16,
		}
		conn, _ := upgrader.Upgrade(w, r, nil)

		_, p, _ := conn.ReadMessage()
		message := string(p)
		if message == connect {
			if g.connectedPlayers == 0 {
				g.players[player1] = conn
				go listenToPlayer(player1)
			} else {
				g.players[player2] = conn
				go listenToPlayer(player2)
			}
			g.connectedPlayers++
			playerConnected(w)
		}
	} else {
		w.Write([]byte("Already enough players in this game."))
	}
}

// main starts a server.
func main() {
	http.HandleFunc("/connect", handler)
	http.ListenAndServe(":8081", nil)
}
