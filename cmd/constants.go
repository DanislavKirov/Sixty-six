// This file contains the messages needed for
// the communication between the server and the clients.
package main

const (
	Player1 = 0
	Player2 = 1

	NoCard = ""
	Rank   = 0
	Suit   = 1

	// client -> server
	Connect  = "connect"
	Exchange = "exchange"
	Close    = "close"
	Quit     = "quit"

	// server -> client
	Waiting           = "Waiting for the other player to connect.\n"
	EnoughPlayers     = "Already enough players.\n"
	Start             = "The game starts now.\n\n"
	YourTurn          = "It's your turn, pick a card number or write command: "
	NotYourTurn       = "It's your opponent's turn, please wait.\n"
	OpponentCard      = "Opponent's card: "
	OpponentLeft      = "Opponent left.\n"
	OpponentClosed    = "Opponent closed.\n"
	OpponentExchanged = "Opponent exchanged the trump.\n"
	YourHand          = "Your hand: "
	Trump             = "Trump: "
	TryAgain          = "Something went wrong! Please try again."
	WrongInput        = "Wrong input, try again: "
	Win               = "You win.\n"
	Lose              = "You lose.\n"
	NotPossible       = "Operation not possible. Try something else: "
)
