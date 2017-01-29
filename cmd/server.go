// This file contains the server of the game.
// The server is responsible for making sure that the game is played by the rules.
package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
)

// game contains the players' connections and deck info.
type game struct {
	connectedPlayers int
	players          [2]net.Conn
	playerInTurn     int

	deck           *deck.Deck
	hands          [2][]string
	trump          string
	isClosed       bool
	trick          [2]string
	hasTrickWon    [2]bool
	marriages      [2]int
	emptyCardSlots [2]int

	dealScore [2]int
	gameScore [2]int
}

// deal deals the first cards as if Player1 is the dealer.
func (g *game) deal() {
	g.hands[Player1] = make([]string, 6)
	g.hands[Player2] = make([]string, 6)
	hands, _ := g.deck.DrawNcards(13)
	copy(g.hands[Player2][:3], hands[:3])
	copy(g.hands[Player2][3:], hands[6:9])
	copy(g.hands[Player1][:3], hands[3:6])
	copy(g.hands[Player1][3:], hands[9:12])
	g.trump = hands[12]
	g.playerInTurn = Player2
}

// getTrumpMsg returns suitable for sending string containing the trump.
func (g *game) getTrumpMsg() string {
	return Trump + replaceTens(g.trump) + "\n"
}

// getHandMsg returns suitable for sending string containing player's hand.
func (g *game) getHandMsg(player int) string {
	return YourHand + replaceTens(strings.Join(g.hands[player], " ")) + "\n"
}

func (g *game) getPointsMsg(player int) string {
	return "Your points: " + strconv.Itoa(g.dealScore[player]) + "\n"
}

// getPlayerNotInTurn returns the player who is waiting.
func (g *game) getPlayerNotInTurn() int {
	return 1 - g.playerInTurn
}

// sendTurnInfo sends info about hands, trump and turns to each player.
func (g *game) sendTurnInfo() {
	info := "\n" + g.getHandMsg(g.playerInTurn) + g.getTrumpMsg() + g.getPointsMsg(g.playerInTurn) + YourTurn
	sendTo(g.playerInTurn, info)
	info = "\n" + g.getHandMsg(g.getPlayerNotInTurn()) + g.getTrumpMsg() + g.getPointsMsg(g.getPlayerNotInTurn()) + NotYourTurn
	sendTo(g.getPlayerNotInTurn(), info)
}

func replaceTens(cards string) string {
	return strings.Replace(cards, "X", "10", -1)
}

// sendTo sends message to player.
func sendTo(player int, message string) {
	g.players[player].Write([]byte(message))
}

// broadcast sends message to both players.
func broadcast(message string) {
	sendTo(Player1, message)
	sendTo(Player2, message)
}

// playerConnected sends suitable message to the players when someone connects.
// Also starts the game when Player2 connects.
func playerConnected() {
	if g.connectedPlayers == 1 {
		sendTo(Player1, Waiting)
	} else {
		broadcast(Start)
		startGame()
	}
}

// startGame creates a deck and deals the first cards.
func startGame() {
	g.deck = deck.New()
	g.deck.Shuffle()
	g.deal()
	g.playerInTurn = Player2
	g.sendTurnInfo()
}

// unexpectedExit informs players if someone quits and closes the connections.
func unexpectedExit() {
	if g.connectedPlayers == 2 {
		broadcast(OpponentLeft)
		g.players[Player2].Close()
	} else {
		sendTo(Player1, OpponentLeft)
	}
	g.players[Player1].Close()
	server.Close()
}

func isTrump(card string) bool {
	return card[Suit:] == g.trump[Suit:]
}

func areTheSameSuit(card1, card2 string) bool {
	return card1[Suit:] == card2[Suit:]
}

func getTheOtherPlayer(player int) int {
	return 1 - player
}

func checkForMarriageWith(rank byte, card string, player int) string {
	for _, c := range g.hands[player] {
		if c != NoCard && c[Rank] == rank && areTheSameSuit(c, card) {
			if isTrump(card) {
				return "40"
			}
			return "20"
		}
	}
	return ""
}

func checkForMarriage(card string, player int) string {
	if card[Rank] == 'Q' {
		return checkForMarriageWith('K', card, player)
	} else if card[Rank] == 'K' {
		return checkForMarriageWith('Q', card, player)
	}
	return ""
}

func isPossibleExchange(player int) (bool, int) {
	if g.trump[Rank] == '9' || !g.hasTrickWon[player] || g.isClosed || len(g.deck.Current) == 0 {
		return false, -1
	}

	for idx, card := range g.hands[player] {
		if isTrump(card) && card[Rank] == '9' {
			return true, idx
		}
	}

	return false, -1
}

func findWinner() int {
	if isTrump(g.trick[Player1]) && !isTrump(g.trick[Player2]) {
		return Player1
	}
	if isTrump(g.trick[Player2]) && !isTrump(g.trick[Player1]) {
		return Player2
	}
	if areTheSameSuit(g.trick[Player1], g.trick[Player2]) {
		if deck.Points[g.trick[Player1][Rank]] > deck.Points[g.trick[Player2][Rank]] {
			return Player1
		}
		return Player2
	}
	return g.getPlayerNotInTurn()
}

func findPoints() int {
	return deck.Points[g.trick[Player1][Rank]] + deck.Points[g.trick[Player2][Rank]]
}

func draw() {
	if len(g.deck.Current) == 0 || g.isClosed {
		return
	} else if len(g.deck.Current) == 1 {
		g.hands[Player1][g.emptyCardSlots[Player1]], _ = g.deck.DrawCard()
		g.hands[Player2][g.emptyCardSlots[Player2]] = g.trump
	} else {
		cards, _ := g.deck.DrawNcards(2)
		g.hands[Player1][g.emptyCardSlots[Player1]] = cards[Player1]
		g.hands[Player2][g.emptyCardSlots[Player2]] = cards[Player2]
	}
}

// listenToPlayer listens what player sends.
func listenToPlayer(player int) {
	p := make([]byte, 256)
	for {
		_, err := g.players[player].Read(p)
		if err != nil {
			unexpectedExit()
			return
		}

		m := string(p)
		if m[0] >= '1' && m[0] <= '6' {
			cardIdx := int(m[0] - '1')
			if g.hands[player][cardIdx] == NoCard {
				sendTo(player, WrongInput)
				continue
			} else {
				card := g.hands[player][cardIdx]
				g.trick[player] = card
				g.hands[player][cardIdx] = NoCard
				g.emptyCardSlots[player] = cardIdx
				marriage := checkForMarriage(card, player)
				if marriage != "" {
					points, _ := strconv.Atoi(marriage)
					g.marriages[player] += points
				}
				sendTo(g.getPlayerNotInTurn(), OpponentCard+replaceTens(card)+" "+marriage+"\n")
			}

			if g.trick[getTheOtherPlayer(player)] == NoCard {
				g.playerInTurn = g.getPlayerNotInTurn()
				sendTo(g.playerInTurn, YourTurn)
			} else {
				draw()
				g.playerInTurn = findWinner()
				g.hasTrickWon[g.playerInTurn] = true
				g.dealScore[g.playerInTurn] += g.marriages[g.playerInTurn] + findPoints()
				g.marriages[g.playerInTurn] = 0

				sendTo(g.playerInTurn, Win)
				sendTo(g.getPlayerNotInTurn(), Lose)

				g.trick[Player1] = NoCard
				g.trick[Player2] = NoCard

				time.Sleep(1 * time.Second)
				g.sendTurnInfo()
			}
		} else if strings.Contains(m, Close) {
			if g.isClosed || len(g.deck.Current) < 2 || g.trick[g.getPlayerNotInTurn()] != NoCard {
				sendTo(g.playerInTurn, NotPossible)
			} else {
				g.isClosed = true
				sendTo(g.getPlayerNotInTurn(), OpponentClosed)
				sendTo(g.playerInTurn, YourTurn)
			}
		} else if strings.Contains(m, Quit) {
			unexpectedExit()
		} else if strings.Contains(m, Exchange) && g.trick[g.getPlayerNotInTurn()] == NoCard {
			if ok, idx := isPossibleExchange(player); ok {
				g.hands[player][idx], g.trump = g.trump, g.hands[player][idx]
				sendTo(g.getPlayerNotInTurn(), OpponentExchanged)
				g.sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		} else {
			sendTo(player, NotPossible)
		}
	}
}

var (
	server net.Listener
	g      = new(game)
)

// startServer starts a server and waits for two players to connect.
func startServer() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	server = l
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		buff := make([]byte, 16)
		_, e := conn.Read(buff)
		if e != nil {
			log.Fatal(err)
		}

		if string(buff[:7]) == Connect {
			if g.connectedPlayers == 0 {
				g.players[Player1] = conn
				go listenToPlayer(Player1)
			} else if g.connectedPlayers == 1 {
				g.players[Player2] = conn
				go listenToPlayer(Player2)
			} else {
				conn.Write([]byte(EnoughPlayers))
				break
			}
			g.connectedPlayers++
			playerConnected()
		} else {
			conn.Write([]byte(NotPossible))
		}
	}
}
