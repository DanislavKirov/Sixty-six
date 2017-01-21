// This file contains the server of the game.
// The server is responsible for making sure that the game is played by the rules.
package main

import (
	"net/http"
	"strings"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
	"github.com/gorilla/websocket"
)

// game contains the connections and deck info.
type game struct {
	connectedPlayers int
	players          [2]*websocket.Conn
	deck             *deck.Deck
	hands            [2]string
	trump            string
}

// deal deals the first cards.
func (g *game) deal() {
	hands, _ := g.deck.DrawNcards(12)
	g.hands[player1] = strings.Join(hands[:3], ", ") + ", " + strings.Join(hands[6:9], ", ")
	g.hands[player2] = strings.Join(hands[3:6], ", ") + ", " + strings.Join(hands[9:], ", ")
	g.trump, _ = g.deck.DrawCard()
}

var (
	g        = new(game)
	requests = []string{"connect", "card", "close", "change", "end"}
)

const (
	player1 = 0
	player2 = 1
)

// sendTo sends a message to player.
func sendTo(player int, message string) {
	g.players[player].WriteMessage(websocket.TextMessage, []byte(message))
}

// playerConnected sends suitable message to the players when someone connects.
func playerConnected(w http.ResponseWriter) {
	if g.connectedPlayers == 1 {
		sendTo(player1, "Waiting for the other player to connect.")
	} else {
		message := "The game is about to begin."
		sendTo(player1, message)
		sendTo(player2, message)
		startGame()
	}
}

// startGame creates a deck and deals the first cards.
func startGame() {
	g.deck = deck.New()
	g.deck.Shuffle()
	g.deal()
	sendTo(player1, g.hands[player1])
	sendTo(player2, g.hands[player2])
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
		if message == requests[0] {
			if g.connectedPlayers == 0 {
				g.players[player1] = conn
			} else {
				g.players[player2] = conn
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
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8081", nil)
}
