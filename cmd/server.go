package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

// deckInfoMsg returns suitable for sending string containing deck info.
func deckInfoMsg() string {
	deckSize := len(g.deck.Current)
	if deckSize != 0 {
		deckSize++ // counting the trump
	}

	return "Trump: " + replaceTens(g.trump) +
		"\tDeck size: " + strconv.Itoa(deckSize) +
		"\tClosed: " + strconv.FormatBool(g.isClosed()) + "\n"
}

// handMsg returns suitable for sending string containing player's hand.
func handMsg(player int) string {
	return "Your hand: " + replaceTens(strings.Join(g.hands[player], " ")) + "\n"
}

// pointsMsg returns suitable for sending string containing deal and g points.
func pointsMsg(player int) string {
	return "Deal points: " + strconv.Itoa(g.dealScore[player]) +
		"\tGame points: " + strconv.Itoa(g.gameScore[player]) +
		":" + strconv.Itoa(g.gameScore[opponentOf(player)]) + "\n"
}

// sendTurnInfo sends info about the deck, hands and points to each player.
func sendTurnInfo() {
	info := "\n" + handMsg(g.playerInTurn) +
		deckInfoMsg() + pointsMsg(g.playerInTurn) + YourTurn
	sendTo(g.playerInTurn, info)

	info = "\n" + handMsg(g.playerNotInTurn()) +
		deckInfoMsg() + pointsMsg(g.playerNotInTurn()) + OpponentTurn
	sendTo(g.playerNotInTurn(), info)
}

// replaceTens gets a hand and replaces the tens to be suitable for printing.
func replaceTens(hand string) string {
	return strings.Replace(hand, "X", "10", -1)
}

// sendTo sends message to player.
func sendTo(player int, message string) {
	players[player].Write([]byte(message))
}

// exit informs players if someone quits and closes the connections.
func exit(player int) {
	if connected == 2 {
		if player != Nobody {
			sendTo(opponentOf(player), OpponentLeft)
		}
		players[Player2].Close()
	}
	players[Player1].Close()
	server.Close()
}

// listenTo listens what player sends and responds accordingly.
func listenTo(player int) {
	buff := make([]byte, 256)
	for {
		size, err := players[player].Read(buff)
		if err != nil {
			exit(player)
			return
		}

		m := string(buff)[:size]
		if size == 2 && m[0] >= '1' && m[0] <= '6' {
			cardIdx := int(m[0] - '1')
			if !g.isCardValid(player, cardIdx) {
				sendTo(player, WrongInput)
				continue
			}

			card := g.playerPlayed(player, cardIdx)
			msg := OpponentCard + replaceTens(card)

			hasMarriage, pts := g.checkForMarriage(player, card)
			if hasMarriage {
				g.addMarriagePoints(player)
				marriage := "Marriage: " + strconv.Itoa(pts) + "\n"
				sendTo(player, marriage)
				msg += " " + marriage
			} else {
				msg += "\n"
			}
			sendTo(g.playerNotInTurn(), msg)

			if g.trick[opponentOf(player)] == NoCard {
				sendTo(player, OpponentTurn)
				g.playerInTurn = g.playerNotInTurn()
				sendTo(g.playerInTurn, YourTurn)
			} else {
				winner := g.findWinner()
				g.playerInTurn = winner
				g.hasTrickWon[winner] = true

				g.addMarriagePoints(winner)
				g.dealScore[winner] += g.trickPoints()

				sendTo(g.playerInTurn, WonTrick)
				sendTo(g.playerNotInTurn(), LostTrick)
				g.trick[Player1] = NoCard
				g.trick[Player2] = NoCard

				g.draw()
				if len(g.hands[player]) == 0 {
					if !g.isClosed() {
						g.dealScore[g.playerInTurn] += LastTrickBonus
					}
					winner, pts := g.endDeal(Nobody)
					ptsStr := strconv.Itoa(pts)
					sendTo(winner, WonDeal+ptsStr+"\n")
					sendTo(opponentOf(winner), LostDeal+ptsStr+"\n")
					sendTurnInfo()
				} else {
					sendTurnInfo()
				}
			}
			continue
		}

		var success bool
		switch m {
		case Close:
			success = g.close(player)
			if success {
				sendTo(g.playerNotInTurn(), OpponentClosed)
				sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		case Exchange:
			success = g.exchange(player)
			if success {
				sendTo(g.playerNotInTurn(), OpponentExchanged)
				sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		case Stop:
			success, winner, pts := g.stop(player)
			if success {
				ptsStr := strconv.Itoa(pts) + "\n"
				sendTo(winner, WonDeal+ptsStr)
				sendTo(opponentOf(winner), LostDeal+ptsStr)
				sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		case Help:
			sendTo(player, Commands+YourTurn)
		case Quit:
			exit(player)
		default:
			sendTo(player, NotPossible)
		}
	}
}

var (
	server    net.Listener
	err       error
	wg        sync.WaitGroup
	players   [2]net.Conn
	g         = new(game)
	connected = 0
)

// startServer starts a server and waits for two players to connect.
func startServer() {
	server, err = net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}
	wg.Done()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		buff := make([]byte, 16)
		size, e := connection.Read(buff)
		if e != nil {
			fmt.Println(err)
			return
		}

		if string(buff[:size]) == Connect {
			connected++
			if connected == 1 {
				players[Player1] = connection
				sendTo(Player1, Waiting)
				go listenTo(Player1)
			} else {
				players[Player2] = connection
				go listenTo(Player2)
				sendTo(Player1, Start)
				sendTo(Player2, Start)
				g.start()
				sendTurnInfo()
				break
			}
		}
	}
}
