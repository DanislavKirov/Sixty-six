// Package game provides api for creating and managing a game of santase.
package main

import "github.com/DanislavKirov/Sixty-six/cmd/deck"

// Game contains info about the deck and the current deal.
type Game struct {
	deck           *deck.Deck
	hands          [2][]string
	trump          string
	trick          [2]string
	closedBy       int
	hasTrickWon    [2]bool
	marriages      [2]int
	emptyCardSlots [2]int
	playerInTurn   int

	dealScore [2]int
	gameScore [2]int
}

// Start creates a deck and deals the first cards.
func (g *Game) Start() {
	g.deck = deck.New()
	g.playerInTurn = Player2
	g.newDeal()
}

// newDeal starts new deal and resets the old deal info.
func (g *Game) newDeal() {
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

// deal deals the first cards as if g.PlayerNotInTurn() is the dealer.
func (g *Game) deal() {
	g.hands[Player1] = make([]string, 6)
	g.hands[Player2] = make([]string, 6)
	hands, _ := g.deck.DrawNcards(13)
	copy(g.hands[g.playerInTurn][:3], hands[:3])
	copy(g.hands[g.playerInTurn][3:], hands[6:9])
	copy(g.hands[g.PlayerNotInTurn()][:3], hands[3:6])
	copy(g.hands[g.PlayerNotInTurn()][3:], hands[9:12])
	g.trump = hands[12]
}

// PlayerNotInTurn returns the player who is waiting.
func (g *Game) PlayerNotInTurn() int {
	return 1 - g.playerInTurn
}

// IsClosed returns true if the deck is closed.
func (g *Game) IsClosed() bool {
	return g.closedBy != Nobody
}

// IsTrump gets a card and checks if it is the same suit as the trump.
func (g *Game) IsTrump(card string) bool {
	return card[Suit:] == g.trump[Suit:]
}

// OpponentOf returns the opponent of the player given as argument.
func OpponentOf(player int) int {
	return 1 - player
}

// AddMarriagePoints adds marriage points to player if he has won a trick.
func (g *Game) AddMarriagePoints(player int) {
	if g.hasTrickWon[player] {
		g.dealScore[player] += g.marriages[player]
		g.marriages[player] = 0
	}
}

// CheckForMarriage returns true and the points made from marriage if any.
func (g *Game) CheckForMarriage(player int, card string) (bool, int) {
	pts := 0
	if (card[Rank] != 'Q' && card[Rank] != 'K') ||
		(g.trick[OpponentOf(player)] != NoCard && !deck.AreTheSameSuit(g.trick[OpponentOf(player)], card) && !g.IsTrump(card)) {
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
			if g.IsTrump(card) {
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
func (g *Game) isPossibleExchange(player int) (bool, int) {
	if g.trick[OpponentOf(player)] != NoCard || g.trump[Rank] == '9' ||
		!g.hasTrickWon[player] || g.IsClosed() || len(g.deck.Current) == 0 {
		return false, -1
	}

	for idx, card := range g.hands[player] {
		if g.IsTrump(card) && card[Rank] == '9' {
			return true, idx
		}
	}

	return false, -1
}

// hasSameSuit returns true if player has a card from the same suit as the card given as argument.
func (g *Game) hasSameSuit(player int, card string) bool {
	for _, c := range g.hands[player] {
		if deck.AreTheSameSuit(c, card) {
			return true
		}
	}
	return false
}

// hasSameSuitHigher returns true if player has a better card to play.
func (g *Game) hasSameSuitHigher(player int, card string) bool {
	for _, c := range g.hands[player] {
		if deck.AreTheSameSuit(c, card) && deck.Points[c[Rank]] > deck.Points[card[Rank]] {
			return true
		}
	}
	return false
}

// hasTrump returns true if player has at least one trump card.
func (g *Game) hasTrump(player int) bool {
	for _, card := range g.hands[player] {
		if g.IsTrump(card) {
			return true
		}
	}
	return false
}

// isGoodResponse checks if player can respond with the given card.
func (g *Game) isGoodResponse(player int, card string) bool {
	otherCard := g.trick[OpponentOf(player)]
	if (g.IsClosed() || len(g.deck.Current) == 0) &&
		(!deck.AreTheSameSuit(card, otherCard) && (g.hasSameSuit(player, otherCard) || (!g.IsTrump(otherCard) && !g.IsTrump(card) && g.hasTrump(player))) ||
			(deck.AreTheSameSuit(card, otherCard) && deck.Points[card[Rank]] < deck.Points[otherCard[Rank]] && g.hasSameSuitHigher(player, otherCard))) {
		return false
	}
	return true
}

// FindWinner returns the player who wins the current trick.
func (g *Game) FindWinner() int {
	if g.IsTrump(g.trick[Player1]) && !g.IsTrump(g.trick[Player2]) {
		return Player1
	}
	if g.IsTrump(g.trick[Player2]) && !g.IsTrump(g.trick[Player1]) {
		return Player2
	}
	if deck.AreTheSameSuit(g.trick[Player1], g.trick[Player2]) {
		if deck.Points[g.trick[Player1][Rank]] > deck.Points[g.trick[Player2][Rank]] {
			return Player1
		}
		return Player2
	}
	return g.PlayerNotInTurn()
}

// TrickPoints returns the points in the current trick.
func (g *Game) TrickPoints() int {
	return deck.Points[g.trick[Player1][Rank]] + deck.Points[g.trick[Player2][Rank]]
}

// Draw replenishes players' hands if deck is not empty or closed.
func (g *Game) Draw() {
	secondToDraw := g.PlayerNotInTurn()
	if len(g.deck.Current) == 0 || g.IsClosed() {
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
func (g *Game) findDealWinPointsAgainst(player int) int {
	if !g.hasTrickWon[player] {
		return 3
	}
	if g.dealScore[player] < 33 {
		return 2
	}
	return 1
}

// findDealWinnerAndPoints returns the winner of the deal and the points.
func (g *Game) findDealWinnerAndPoints(player, score1, score2 int) (int, int) {
	if !g.hasTrickWon[player] {
		return OpponentOf(player), 3
	}
	if score1 >= 66 && score1 > score2 {
		return player, g.findDealWinPointsAgainst(OpponentOf(player))
	}
	return OpponentOf(player), 2
}

// EndDeal gives points to the winner and begins new deal if nobody has >= 11 points.
// It returns the winner and the points he has won.
func (g *Game) EndDeal(player int) (int, int) {
	score1 := g.dealScore[Player1]
	score2 := g.dealScore[Player2]

	var winner, pts int
	if player == Nobody && !g.IsClosed() {
		if score1 > score2 {
			winner = Player1
			pts = g.findDealWinPointsAgainst(Player2)
		} else {
			winner = Player2
			pts = g.findDealWinPointsAgainst(Player1)
		}
	} else if player == Nobody && g.IsClosed() {
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
		g.playerInTurn = OpponentOf(winner)
		g.newDeal()
	}

	return winner, pts
}

// IsCardValid returns true if player can respond with cardIdx.
func (g *Game) IsCardValid(player, cardIdx int) bool {
	if len(g.hands[player]) <= cardIdx ||
		(g.trick[OpponentOf(player)] != NoCard && !g.isGoodResponse(player, g.hands[player][cardIdx])) {
		return false
	}
	return true
}

// PlayerPlayed puts the card on the table and returns it.
func (g *Game) PlayerPlayed(player, cardIdx int) string {
	card := g.hands[player][cardIdx]
	g.trick[player] = card
	g.hands[player][cardIdx] = NoCard
	g.emptyCardSlots[player] = cardIdx
	return card
}

// GetTrickCard returns the card player has put on the table.
func (g *Game) GetTrickCard(player int) string {
	return g.trick[player]
}

// IsDeckEmpty returns true if there is no more cards in the deck.
func (g *Game) IsDeckEmpty() bool {
	return len(g.deck.Current) == 0
}

// Close changes the deal to closed by player if possible and returns if succeeded.
func (g *Game) Close(player int) bool {
	if g.IsClosed() || g.IsDeckEmpty() || g.trick[g.PlayerNotInTurn()] != NoCard {
		return false
	}
	g.closedBy = player
	return true
}

// Exchange makes nine-trump exchange if possible and returns if it succeeded.
func (g *Game) Exchange(player int) bool {
	if ok, idx := g.isPossibleExchange(player); ok {
		g.hands[player][idx], g.trump = g.trump, g.hands[player][idx]
		return true
	}
	return false
}

// Stop ends the current deal and finds the winner and the points if possible.
// It returns true if succeeded and winner and points.
func (g *Game) Stop(player int) (bool, int, int) {
	if g.trick[OpponentOf(player)] == NoCard {
		winner, pts := g.EndDeal(player)
		return true, winner, pts
	}
	return false, Nobody, 0
}

// NextPlayer changes the player in turn.
func (g *Game) NextPlayer() {
	g.playerInTurn = g.PlayerNotInTurn()
}

// MakePlayerInTurn changes the player in turn to the given one.
func (g *Game) MakePlayerInTurn(player int) {
	g.playerInTurn = player
}

// PlayerInTurn returns the player in turn.
func (g *Game) PlayerInTurn() int {
	return g.playerInTurn
}

// WinTrick marks that player has won a trick.
func (g *Game) WinTrick(player int) {
	g.hasTrickWon[player] = true
}

// AddPointsTo adds the given pts to player.
func (g *Game) AddPointsTo(player, pts int) {
	g.dealScore[player] += pts
}

// ClearTable clears the trick cards.
func (g *Game) ClearTable() {
	g.trick[Player1] = NoCard
	g.trick[Player2] = NoCard
}

// IsHandEmpty returns true if deal should end.
func (g *Game) IsHandEmpty() bool {
	return len(g.hands[Player1]) == 0
}

// DeckSize returns the size of the deck (including the trump).
func (g *Game) DeckSize() int {
	size := len(g.deck.Current)
	if size == 0 {
		return size
	}
	return size + 1
}

// GetHand returns the cards of the requested player.
func (g *Game) GetHand(player int) []string {
	return g.hands[player]
}

// Trump returns the trump of the deal.
func (g *Game) Trump() string {
	return g.trump
}

// GetPoints returns players' deal score, game score and his opponents' game score.
func (g *Game) GetPoints(player int) (int, int, int) {
	return g.dealScore[player], g.gameScore[player], g.gameScore[OpponentOf(player)]
}
