package main

const (
	Player1 = 0
	Player2 = 1
	Nobody  = 2

	NoCard = ""
	Rank   = 0
	Suit   = 1

	LastTrickBonus = 10

	// client -> server

	Connect  = "connect"
	Exchange = "exchange"
	Close    = "close"
	Stop     = "stop"
	Help     = "help"
	Quit     = "quit"

	// server -> client

	Waiting           = "Waiting for the other player to connect.\n"
	Start             = "The game starts now.\n\n"
	YourTurn          = "It's your turn, pick a card number (1-6) or write a command: "
	OpponentTurn      = "It's your opponent's turn, please wait.\n"
	OpponentCard      = "Opponent's card: "
	OpponentLeft      = "Opponent left.\n"
	OpponentClosed    = "Opponent closed.\n"
	OpponentExchanged = "Opponent exchanged the trump.\n"
	TryAgain          = "Something went wrong! Please try again."
	WrongInput        = "Wrong input, try again: "
	WonTrick          = "You won this trick.\n"
	WonDeal           = "You won this deal. Points: "
	WonGame           = "YOU WON THE GAME!\n"
	LostTrick         = "You lost this trick.\n"
	LostDeal          = "You lost this deal. Opponents gets: "
	LostGame          = "You lost the game.\n"
	NotPossible       = "Operation not possible. Try something else: "
	Commands          = "Commands:\n* exchange\n* close\n* stop\n* quit\n"
)
