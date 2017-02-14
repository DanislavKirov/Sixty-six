package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/DanislavKirov/Sixty-six/cmd/deck"
)

// game contains the players' connections and game info.
type game struct {
	connectedPlayers int
	players          [2]net.Conn
	playerInTurn     int

	deck           *deck.Deck
	hands          [2][]string
	trump          string
	closedBy       int
	trick          [2]string
	hasTrickWon    [2]bool
	marriages      [2]int
	emptyCardSlots [2]int

	dealScore [2]int
	gameScore [2]int
}

// getDeckInfoMsg returns suitable for sending string containing deck info.
func (g *game) getDeckInfoMsg() string {
	deckSize := len(g.deck.Current)
	if deckSize != 0 {
		deckSize++
	}

	return "Trump: " + replaceTens(g.trump) +
		"\tDeck size: " + strconv.Itoa(deckSize) +
		"\tClosed: " + strconv.FormatBool(g.isClosed()) + "\n"
}

// getHandMsg returns suitable for sending string containing player's hand.
func (g *game) getHandMsg(player int) string {
	return "Your hand: " + replaceTens(strings.Join(g.hands[player], " ")) + "\n"
}

// getPointsMsg returns suitable for sending string containing deal and game points.
func (g *game) getPointsMsg(player int) string {
	return "Deal points: " + strconv.Itoa(g.dealScore[player]) +
		"\tGame points: " + strconv.Itoa(g.gameScore[player]) +
		":" + strconv.Itoa(g.gameScore[getTheOtherPlayer(player)]) + "\n"
}

// getPlayerNotInTurn returns the player who is waiting.
func (g *game) getPlayerNotInTurn() int {
	return 1 - g.playerInTurn
}

// sendTurnInfo sends info about the deck, hands and points to each player.
func (g *game) sendTurnInfo() {
	info := "\n" + g.getHandMsg(g.playerInTurn) +
		g.getDeckInfoMsg() + g.getPointsMsg(g.playerInTurn) + YourTurn
	sendTo(g.playerInTurn, info)

	info = "\n" + g.getHandMsg(g.getPlayerNotInTurn()) +
		g.getDeckInfoMsg() + g.getPointsMsg(g.getPlayerNotInTurn()) + OpponentTurn
	sendTo(g.getPlayerNotInTurn(), info)
}

// isClosed returns true if the deck is closed.
func (g *game) isClosed() bool {
	return g.closedBy != Nobody
}

// start creates a deck and deals the first cards.
func (g *game) start() {
	g.deck = deck.New()
	g.playerInTurn = Player2
	g.newDeal()
}

// newDeal starts new deal and resets the old deal info.
func (g *game) newDeal() {
	g.deck.Shuffle()
	g.closedBy = Nobody
	g.trick[Player1] = NoCard
	g.trick[Player2] = NoCard
	g.hasTrickWon[Player1] = false
	g.hasTrickWon[Player2] = false
	g.dealScore[Player1] = 0
	g.dealScore[Player2] = 0
	g.deal()
	g.sendTurnInfo()
}

// deal deals the first cards as if g.playerInTurn is the dealer.
func (g *game) deal() {
	g.hands[Player1] = make([]string, 6)
	g.hands[Player2] = make([]string, 6)
	hands, _ := g.deck.DrawNcards(13)
	copy(g.hands[g.playerInTurn][:3], hands[:3])
	copy(g.hands[g.playerInTurn][3:], hands[6:9])
	copy(g.hands[g.getPlayerNotInTurn()][:3], hands[3:6])
	copy(g.hands[g.getPlayerNotInTurn()][3:], hands[9:12])
	g.trump = hands[12]
}

// replaceTens gets a hand and replaces the tens to be suitable for printing.
func replaceTens(hand string) string {
	return strings.Replace(hand, "X", "10", -1)
}

// sendTo sends message to player.
func sendTo(player int, message string) {
	g.players[player].Write([]byte(message))
}

// playerConnected sends suitable message to the players when someone connects.
// Also starts the game when Player2 connects.
func playerConnected() {
	if g.connectedPlayers == 1 {
		sendTo(Player1, Waiting)
	} else {
		sendTo(Player1, Start)
		sendTo(Player2, Start)
		g.start()
	}
}

// exit informs players if someone quits and closes the connections.
func exit(player int) {
	if g.connectedPlayers == 2 {
		sendTo(getTheOtherPlayer(player), OpponentLeft)
		g.players[Player2].Close()
	}
	g.players[Player1].Close()
	server.Close()
}

// isTrump gets a card and checks if it is the same suit as the trump.
func isTrump(card string) bool {
	return card[Suit:] == g.trump[Suit:]
}

// areTheSameSuit returns true if card1 and card2 are from the same suit.
func areTheSameSuit(card1, card2 string) bool {
	return card1[Suit:] == card2[Suit:]
}

// getTheOtherPlayer returns the player who is not player.
func getTheOtherPlayer(player int) int {
	return 1 - player
}

// checkForMarriageWith returns the points made from marriage with card if any.
func checkForMarriageWith(rank byte, card string, player int) (bool, int) {
	for _, c := range g.hands[player] {
		if c != NoCard && c[Rank] == rank && areTheSameSuit(c, card) {
			if isTrump(card) {
				return true, 40
			}
			return true, 20
		}
	}
	return false, 0
}

// checkForMarriage returns the points made from marriage if any.
func checkForMarriage(card string, player int) (bool, int) {
	if g.trick[getTheOtherPlayer(player)] != NoCard &&
		!isTrump(card) && !areTheSameSuit(g.trick[getTheOtherPlayer(player)], card) {
		return false, 0
	}
	if card[Rank] == 'Q' {
		return checkForMarriageWith('K', card, player)
	} else if card[Rank] == 'K' {
		return checkForMarriageWith('Q', card, player)
	}
	return false, 0
}

// isPossibleExchange returns true if nine-trump exchange is possible.
func isPossibleExchange(player int) (bool, int) {
	if g.trick[getTheOtherPlayer(player)] != NoCard || g.trump[Rank] == '9' ||
		!g.hasTrickWon[player] || g.isClosed() || len(g.deck.Current) == 0 {
		return false, -1
	}

	for idx, card := range g.hands[player] {
		if isTrump(card) && card[Rank] == '9' {
			return true, idx
		}
	}

	return false, -1
}

// hasSameSuit returns true if player has a card from the same suit as cars.
func hasSameSuit(player int, card string) bool {
	for _, c := range g.hands[player] {
		if areTheSameSuit(c, card) {
			return true
		}
	}
	return false
}

// hasSameSuitHigher returns true if player has a better card to play.
func hasSameSuitHigher(player int, card string) bool {
	for _, c := range g.hands[player] {
		if areTheSameSuit(c, card) && deck.Points[c[Rank]] > deck.Points[card[Rank]] {
			return true
		}
	}
	return false
}

// hasTrump checks if playes has trumps.
func hasTrump(player int) bool {
	for _, c := range g.hands[player] {
		if isTrump(c) {
			return true
		}
	}
	return false
}

// hasHigherTrump checks if playes has a trump with higher rank than card.
func hasHigherTrump(player int, card1, card2 string) bool {
	if isTrump(card1) && deck.Points[card1[Rank]] > deck.Points[card2[Rank]] {
		return false
	}

	for _, c := range g.hands[player] {
		if isTrump(c) && deck.Points[c[Rank]] > deck.Points[card2[Rank]] {
			return true
		}
	}
	return false
}

// isGoodResponse checks if player can respond with card.
func isGoodResponse(player int, card string) bool {
	otherCard := g.trick[getTheOtherPlayer(player)]
	if (g.isClosed() || len(g.deck.Current) == 0) &&
		(!areTheSameSuit(card, otherCard) && (hasSameSuit(player, otherCard) || (!isTrump(otherCard) && !isTrump(card) && hasTrump(player))) ||
			(areTheSameSuit(card, otherCard) && deck.Points[card[Rank]] < deck.Points[otherCard[Rank]] && hasSameSuitHigher(player, otherCard))) {
		return false
	}
	return true
}

// findWinner returns the player who wins the current trick.
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

// findPoints returns points in current trick.
func findPoints() int {
	return deck.Points[g.trick[Player1][Rank]] + deck.Points[g.trick[Player2][Rank]]
}

// draw replenishes players' hands if deck not empty or closed.
func draw() {
	if len(g.deck.Current) == 0 || g.isClosed() {
		g.hands[Player1] = append(g.hands[Player1][:g.emptyCardSlots[Player1]], g.hands[Player1][g.emptyCardSlots[Player1]+1:]...)
		g.hands[Player2] = append(g.hands[Player2][:g.emptyCardSlots[Player2]], g.hands[Player2][g.emptyCardSlots[Player2]+1:]...)
	} else if len(g.deck.Current) == 1 {
		g.hands[Player1][g.emptyCardSlots[Player1]], _ = g.deck.DrawCard()
		g.hands[Player2][g.emptyCardSlots[Player2]] = g.trump
	} else {
		cards, _ := g.deck.DrawNcards(2)
		g.hands[Player1][g.emptyCardSlots[Player1]] = cards[Player1]
		g.hands[Player2][g.emptyCardSlots[Player2]] = cards[Player2]
	}
}

// findDealWinPointsAgainst returns deal win points.
func findDealWinPointsAgainst(player int) int {
	if !g.hasTrickWon[player] {
		return 3
	}
	if g.dealScore[player] < 33 {
		return 2
	}
	return 1
}

// findDealWinnerAndPoints returns the winner of the deal and the points.
func findDealWinnerAndPoints(player, score1, score2 int) (int, int) {
	if !g.hasTrickWon[player] {
		return getTheOtherPlayer(player), 3
	}
	if score1 >= 66 && score1 > score2 {
		return player, findDealWinPointsAgainst(getTheOtherPlayer(player))
	}
	return getTheOtherPlayer(player), 2
}

// endDeal gives points to the winner and begins new deal if nobody has >= 11 points.
func endDeal(player int) {
	score1 := g.dealScore[Player1]
	score2 := g.dealScore[Player2]

	var winner, pts int
	if player == Nobody && !g.isClosed() {
		if score1 > score2 {
			winner = Player1
			pts = findDealWinPointsAgainst(Player2)
		} else {
			winner = Player2
			pts = findDealWinPointsAgainst(Player1)
		}
	} else if player == Nobody && g.isClosed() {
		if g.closedBy == Player1 {
			winner, pts = findDealWinnerAndPoints(Player1, score1, score2)
		} else {
			winner, pts = findDealWinnerAndPoints(Player2, score2, score1)
		}
	} else if player == Player1 {
		winner, pts = findDealWinnerAndPoints(Player1, score1, score2)
	} else {
		winner, pts = findDealWinnerAndPoints(Player2, score2, score1)
	}

	g.gameScore[winner] += pts
	ptsString := strconv.Itoa(pts) + "\n"
	sendTo(winner, WonDeal+ptsString)
	sendTo(getTheOtherPlayer(winner), LostDeal+ptsString)

	if g.gameScore[winner] >= 11 {
		sendTo(winner, WonGame)
		sendTo(getTheOtherPlayer(winner), LostGame)
		endGame()
		return
	}

	g.playerInTurn = getTheOtherPlayer(winner)
	g.newDeal()
}

// endGame closes the connections and the server.
func endGame() {
	g.players[Player1].Close()
	g.players[Player2].Close()
	server.Close()
}

// listenToPlayer listens what player sends and responds accordingly.
func listenToPlayer(player int) {
	p := make([]byte, 256)
	for {
		size, err := g.players[player].Read(p)
		if err != nil {
			exit(player)
			return
		}

		m := string(p)[:size]
		if m[0] >= '1' && m[0] <= '6' {
			cardIdx := int(m[0] - '1')
			if len(g.hands[player]) <= cardIdx ||
				(g.trick[getTheOtherPlayer(player)] != NoCard && !isGoodResponse(player, g.hands[player][cardIdx])) {
				sendTo(player, WrongInput)
				continue
			} else {
				card := g.hands[player][cardIdx]
				g.trick[player] = card
				g.hands[player][cardIdx] = NoCard
				g.emptyCardSlots[player] = cardIdx
				hasMarriage, points := checkForMarriage(card, player)
				if hasMarriage {
					g.marriages[player] += points
					marriage := "Marriage: " + strconv.Itoa(points) + "\n"
					sendTo(g.getPlayerNotInTurn(), OpponentCard+replaceTens(card)+" "+marriage)
					sendTo(player, marriage)
				} else {
					sendTo(g.getPlayerNotInTurn(), OpponentCard+replaceTens(card)+"\n")
				}
			}

			if g.hasTrickWon[player] {
				g.dealScore[player] += g.marriages[player]
				g.marriages[player] = 0
			}

			if g.trick[getTheOtherPlayer(player)] == NoCard {
				sendTo(player, OpponentTurn)
				g.playerInTurn = g.getPlayerNotInTurn()
				sendTo(g.playerInTurn, YourTurn)
			} else {
				draw()
				g.playerInTurn = findWinner()
				g.hasTrickWon[g.playerInTurn] = true
				g.dealScore[g.playerInTurn] += g.marriages[g.playerInTurn] + findPoints()
				g.marriages[g.playerInTurn] = 0

				sendTo(g.playerInTurn, WonTrick)
				sendTo(g.getPlayerNotInTurn(), LostTrick)

				g.trick[Player1] = NoCard
				g.trick[Player2] = NoCard

				if len(g.hands[player]) == 0 {
					g.dealScore[g.playerInTurn] += LastTrickBonus
					endDeal(Nobody)
				} else {
					g.sendTurnInfo()
				}
			}
			continue
		}

		switch m {
		case Close:
			if g.isClosed() || len(g.deck.Current) < 2 || g.trick[g.getPlayerNotInTurn()] != NoCard {
				sendTo(g.playerInTurn, NotPossible)
			} else {
				g.closedBy = player
				sendTo(g.getPlayerNotInTurn(), OpponentClosed)
				g.sendTurnInfo()
			}
		case Exchange:
			if ok, idx := isPossibleExchange(player); ok {
				g.hands[player][idx], g.trump = g.trump, g.hands[player][idx]
				sendTo(g.getPlayerNotInTurn(), OpponentExchanged)
				g.sendTurnInfo()
			} else {
				sendTo(player, NotPossible)
			}
		case Stop:
			if g.trick[getTheOtherPlayer(player)] != NoCard {
				sendTo(player, NotPossible)
			} else {
				endDeal(player)
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
	server net.Listener
	wg     sync.WaitGroup
	g      = new(game)
)

// startServer starts a server and waits for two players to connect.
func startServer() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	server = l
	wg.Done()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		buff := make([]byte, 16)
		size, e := conn.Read(buff)
		if e != nil {
			log.Fatal(err)
		}

		if string(buff[:size]) == Connect {
			if g.connectedPlayers == 0 {
				g.players[Player1] = conn
				go listenToPlayer(Player1)
			} else {
				g.players[Player2] = conn
				go listenToPlayer(Player2)
			}
			g.connectedPlayers++
			go playerConnected()
			if g.connectedPlayers == 2 {
				return
			}
		} else {
			conn.Write([]byte(NotPossible))
		}
	}
}
