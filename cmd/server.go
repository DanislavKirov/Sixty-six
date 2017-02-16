package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

// deckInfoMsg returns suitable for sending string containing deck info.
func deckInfoMsg() string {
	return "Trump: " + replaceTens(game.Trump()) +
		"\tDeck size: " + strconv.Itoa(game.DeckSize()) +
		"\tClosed: " + strconv.FormatBool(game.IsClosed()) + "\n"
}

// handMsg returns suitable for sending string containing player's hand.
func handMsg(player int) string {
	return "Your hand: " + replaceTens(strings.Join(game.GetHand(player), " ")) + "\n"
}

// pointsMsg returns suitable for sending string containing deal and game points.
func pointsMsg(player int) string {
	dealPts, gamePts1, gamePts2 := game.GetPoints(player)

	return "Deal points: " + strconv.Itoa(dealPts) +
		"\tGame points: " + strconv.Itoa(gamePts1) +
		":" + strconv.Itoa(gamePts2) + "\n"
}

// sendTurnInfo sends info about the deck, hands and points to each player.
func sendTurnInfo() {
	info := "\n" + handMsg(game.PlayerInTurn()) +
		deckInfoMsg() + pointsMsg(game.PlayerInTurn()) + YourTurn
	sendTo(game.PlayerInTurn(), info)

	info = "\n" + handMsg(game.PlayerNotInTurn()) +
		deckInfoMsg() + pointsMsg(game.PlayerNotInTurn()) + OpponentTurn
	sendTo(game.PlayerNotInTurn(), info)
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
			sendTo(OpponentOf(player), OpponentLeft)
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
			if !game.IsCardValid(player, cardIdx) {
				sendTo(player, WrongInput)
				continue
			}

			card := game.PlayerPlayed(player, cardIdx)
			msg := OpponentCard + replaceTens(card)

			hasMarriage, pts := game.CheckForMarriage(player, card)
			if hasMarriage {
				game.AddMarriagePoints(player)
				marriage := " Marriage: " + strconv.Itoa(pts) + "\n"
				sendTo(player, marriage)
				msg += marriage
			} else {
				msg += "\n"
			}
			sendTo(game.PlayerNotInTurn(), msg)

			if game.GetTrickCard(OpponentOf(player)) == NoCard {
				sendTo(player, OpponentTurn)
				game.NextPlayer()
				sendTo(game.PlayerInTurn(), YourTurn)
			} else {
				winner := game.FindWinner()
				game.MakePlayerInTurn(winner)
				game.WinTrick(winner)

				game.AddMarriagePoints(winner)
				game.AddPointsTo(winner, game.TrickPoints())

				sendTo(game.PlayerInTurn(), WonTrick)
				sendTo(game.PlayerNotInTurn(), LostTrick)
				game.ClearTable()

				game.Draw()
				if game.IsHandEmpty() {
					if !game.IsClosed() {
						game.AddPointsTo(game.PlayerInTurn(), LastTrickBonus)
					}
					winner, pts := game.EndDeal(Nobody)
					ptsStr := strconv.Itoa(pts)
					sendTo(winner, WonDeal+ptsStr+"\n")
					sendTo(OpponentOf(winner), LostDeal+ptsStr+"\n")
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
			success = game.Close(player)
			if success {
				sendTo(game.PlayerNotInTurn(), OpponentClosed)
				sendTurnInfo()
			} else {
				sendTo(game.PlayerInTurn(), NotPossible)
			}
		case Exchange:
			success = game.Exchange(player)
			if success {
				sendTo(game.PlayerNotInTurn(), OpponentExchanged)
				sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		case Stop:
			success, winner, pts := game.Stop(player)
			if success {
				ptsStr := strconv.Itoa(pts) + "\n"
				sendTo(winner, WonDeal+ptsStr)
				sendTo(OpponentOf(winner), LostDeal+ptsStr)
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
	game      = new(Game)
	connected = 0
)

// startServer starts a server and waits for two players to connect.
func startServer() {
	server, err = net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	wg.Done()

	for {
		connection, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}

		buff := make([]byte, 16)
		size, e := connection.Read(buff)
		if e != nil {
			log.Fatal(err)
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
				game.Start()
				sendTurnInfo()
				break
			}
		}
	}
}
