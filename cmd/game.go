package main

import "github.com/DanislavKirov/Sixty-six/cmd/deck"

// game contains info about the deck and the current deal.
type game struct {
	deck      *deck.Deck
	gameScore [2]int

	hands          [2][]string
	trump          string
	closedBy       int
	trick          [2]string
	hasTrickWon    [2]bool
	marriages      [2]int
	emptyCardSlots [2]int
	playerInTurn   int
	dealScore      [2]int
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
}

// deal deals the first cards as if g.playerNotInTurn() is the dealer.
func (g *game) deal() {
	g.hands[Player1] = make([]string, 6)
	g.hands[Player2] = make([]string, 6)
	hands, _ := g.deck.DrawNcards(13)
	copy(g.hands[g.playerInTurn][:3], hands[:3])
	copy(g.hands[g.playerInTurn][3:], hands[6:9])
	copy(g.hands[g.playerNotInTurn()][:3], hands[3:6])
	copy(g.hands[g.playerNotInTurn()][3:], hands[9:12])
	g.trump = hands[12]
}

// playerNotInTurn returns the player who is waiting.
func (g *game) playerNotInTurn() int {
	return 1 - g.playerInTurn
}

// isClosed returns true if the deck is closed.
func (g *game) isClosed() bool {
	return g.closedBy != Nobody
}

// isTrump gets a card and checks if it is the same suit as the trump.
func (g *game) isTrump(card string) bool {
	return card[Suit:] == g.trump[Suit:]
}

// opponentOf returns the opponent of the player given as argument.
func opponentOf(player int) int {
	return 1 - player
}

// addMarriagePoints adds marriage points to player if he has won a trick.
func (g *game) addMarriagePoints(player int) {
	if g.hasTrickWon[player] {
		g.dealScore[player] += g.marriages[player]
		g.marriages[player] = 0
	}
}

// checkForMarriage returns true and the points made from a marriage if any.
func (g *game) checkForMarriage(player int, card string) (bool, int) {
	pts := 0
	if (card[Rank] != 'Q' && card[Rank] != 'K') ||
		(g.trick[opponentOf(player)] != NoCard && !deck.AreTheSameSuit(g.trick[opponentOf(player)], card) && !g.isTrump(card)) {
		return false, pts
	}

	var rank byte
	if card[Rank] == 'Q' {
		rank = 'K'
	} else if card[Rank] == 'K' {
		rank = 'Q'
	}

	for _, c := range g.hands[player] {
		if c != NoCard && c[Rank] == rank && deck.AreTheSameSuit(c, card) {
			if g.isTrump(card) {
				pts = 40
			} else {
				pts = 20
			}
			g.marriages[player] += pts
			return true, pts
		}
	}
	return false, pts
}

// isPossibleExchange returns true if nine-trump exchange is possible.
func (g *game) isPossibleExchange(player int) (bool, int) {
	if g.trick[opponentOf(player)] != NoCard || g.trump[Rank] == '9' ||
		!g.hasTrickWon[player] || g.isClosed() || len(g.deck.Current) == 0 {
		return false, -1
	}

	for idx, card := range g.hands[player] {
		if g.isTrump(card) && card[Rank] == '9' {
			return true, idx
		}
	}
	return false, -1
}

// hasSameSuit returns true if the player has a card from the same suit as the card given as argument.
func (g *game) hasSameSuit(player int, card string) bool {
	for _, c := range g.hands[player] {
		if deck.AreTheSameSuit(c, card) {
			return true
		}
	}
	return false
}

// hasSameSuitHigher returns true if the player has a card from the same suit but higher rank than the card given.
func (g *game) hasSameSuitHigher(player int, card string) bool {
	for _, c := range g.hands[player] {
		if deck.AreTheSameSuit(c, card) && deck.Points[c[Rank]] > deck.Points[card[Rank]] {
			return true
		}
	}
	return false
}

// hasTrump returns true if player has at least one trump card.
func (g *game) hasTrump(player int) bool {
	for _, card := range g.hands[player] {
		if g.isTrump(card) {
			return true
		}
	}
	return false
}

// isGoodResponse checks if the player can respond with the given card.
func (g *game) isGoodResponse(player int, card string) bool {
	otherCard := g.trick[opponentOf(player)]
	if (g.isClosed() || len(g.deck.Current) == 0) &&
		(!deck.AreTheSameSuit(card, otherCard) &&
			(g.hasSameSuit(player, otherCard) || (!g.isTrump(otherCard) && !g.isTrump(card) && g.hasTrump(player))) ||
			(deck.AreTheSameSuit(card, otherCard) && deck.Points[card[Rank]] < deck.Points[otherCard[Rank]] && g.hasSameSuitHigher(player, otherCard))) {
		return false
	}
	return true
}

// findWinner returns the player who wins the current trick.
func (g *game) findWinner() int {
	if g.isTrump(g.trick[Player1]) && !g.isTrump(g.trick[Player2]) {
		return Player1
	}
	if g.isTrump(g.trick[Player2]) && !g.isTrump(g.trick[Player1]) {
		return Player2
	}
	if deck.AreTheSameSuit(g.trick[Player1], g.trick[Player2]) {
		if deck.Points[g.trick[Player1][Rank]] > deck.Points[g.trick[Player2][Rank]] {
			return Player1
		}
		return Player2
	}
	return g.playerNotInTurn()
}

// trickPoints returns the points in the current trick.
func (g *game) trickPoints() int {
	return deck.Points[g.trick[Player1][Rank]] + deck.Points[g.trick[Player2][Rank]]
}

// draw replenishes players' hands if deck is not empty or closed.
func (g *game) draw() {
	secondToDraw := g.playerNotInTurn()
	if len(g.deck.Current) == 0 || g.isClosed() {
		g.hands[Player1] = append(g.hands[Player1][:g.emptyCardSlots[Player1]], g.hands[Player1][g.emptyCardSlots[Player1]+1:]...)
		g.hands[Player2] = append(g.hands[Player2][:g.emptyCardSlots[Player2]], g.hands[Player2][g.emptyCardSlots[Player2]+1:]...)
	} else if len(g.deck.Current) == 1 {
		g.hands[g.playerInTurn][g.emptyCardSlots[g.playerInTurn]], _ = g.deck.DrawCard()
		g.hands[secondToDraw][g.emptyCardSlots[secondToDraw]] = g.trump
	} else {
		cards, _ := g.deck.DrawNcards(2)
		g.hands[g.playerInTurn][g.emptyCardSlots[g.playerInTurn]] = cards[0]
		g.hands[secondToDraw][g.emptyCardSlots[secondToDraw]] = cards[1]
	}
}

// findDealWinPointsAgainst returns deal win points.
func (g *game) findDealWinPointsAgainst(player int) int {
	if !g.hasTrickWon[player] {
		return 3
	}
	if g.dealScore[player] < 33 {
		return 2
	}
	return 1
}

// findDealWinnerAndPoints returns the winner of the deal and the points.
func (g *game) findDealWinnerAndPoints(player, score1, score2 int) (int, int) {
	if !g.hasTrickWon[player] {
		return opponentOf(player), 3
	}
	if score1 >= 66 && score1 > score2 {
		return player, g.findDealWinPointsAgainst(opponentOf(player))
	}
	return opponentOf(player), 2
}

// endDeal gives points to the winner and begins new deal if nobody has >= 11 points.
// It returns the winner and the points he has won.
func (g *game) endDeal(player int) (int, int) {
	score1 := g.dealScore[Player1]
	score2 := g.dealScore[Player2]

	var winner, pts int
	if player == Nobody && !g.isClosed() {
		if score1 > score2 {
			winner = Player1
			pts = g.findDealWinPointsAgainst(Player2)
		} else {
			winner = Player2
			pts = g.findDealWinPointsAgainst(Player1)
		}
	} else if player == Nobody && g.isClosed() {
		if g.closedBy == Player1 {
			winner, pts = g.findDealWinnerAndPoints(Player1, score1, score2)
		} else {
			winner, pts = g.findDealWinnerAndPoints(Player2, score2, score1)
		}
	} else if player == Player1 {
		winner, pts = g.findDealWinnerAndPoints(Player1, score1, score2)
	} else {
		winner, pts = g.findDealWinnerAndPoints(Player2, score2, score1)
	}

	g.gameScore[winner] += pts
	if g.gameScore[winner] < 11 {
		g.playerInTurn = opponentOf(winner)
		g.newDeal()
	}

	return winner, pts
}

// isCardValid returns true if player can respond with cardIdx.
func (g *game) isCardValid(player, cardIdx int) bool {
	if len(g.hands[player]) <= cardIdx ||
		(g.trick[opponentOf(player)] != NoCard && !g.isGoodResponse(player, g.hands[player][cardIdx])) {
		return false
	}
	return true
}

// playerPlayed puts the card on the table and returns it.
func (g *game) playerPlayed(player, cardIdx int) string {
	card := g.hands[player][cardIdx]
	g.trick[player] = card
	g.hands[player][cardIdx] = NoCard
	g.emptyCardSlots[player] = cardIdx
	return card
}

// close changes the deal to closed by player if possible and returns if succeeded.
func (g *game) close(player int) bool {
	if g.isClosed() || len(g.deck.Current) == 0 || g.trick[g.playerNotInTurn()] != NoCard {
		return false
	}
	g.closedBy = player
	return true
}

// exchange makes nine-trump exchange if possible and returns if it succeeded.
func (g *game) exchange(player int) bool {
	if ok, idx := g.isPossibleExchange(player); ok {
		g.hands[player][idx], g.trump = g.trump, g.hands[player][idx]
		return true
	}
	return false
}

// stop ends the current deal and finds the winner and the points if possible.
// It returns true if succeeded and winner and points.
func (g *game) stop(player int) (bool, int, int) {
	if g.trick[opponentOf(player)] == NoCard {
		winner, pts := g.endDeal(player)
		return true, winner, pts
	}
	return false, Nobody, 0
}
