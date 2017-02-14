package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// menu connects the client depending on his choise.
func menu() {
	choice := 0
	reader := bufio.NewReader(os.Stdin)
	for choice < 1 || choice > 3 {
		fmt.Print("\nPick one:\n1. Create game\n2. Join game\n3. Single player\nYour choise: ")
		input, err := reader.ReadString('\n')
		if err != nil || len(input) > 2 {
			continue
		}
		choice = int(input[0] - '0')
	}

	switch choice {
	case 1:
		client1()
	case 2:
		client2()
	case 3:
		client3()
	}
}

// externalIP finds the IP of the client creating the game.
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// client1 starts the server and connects the first player.
func client1() {
	wg.Add(1)
	go startServer()
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	wg.Wait()
	port := server.Addr().String()
	idx := strings.LastIndex(port, ":")
	fmt.Println("IP:port = " + ip + port[idx:])
	singlePlayer := false
	connect("localhost"+port[idx:], singlePlayer)
}

// client2 connects the second player to the server entering IP:port.
func client2() {
	fmt.Print("Enter ip:port: ")
	reader := bufio.NewReader(os.Stdin)
	ip, err := reader.ReadString('\n')
	for err != nil {
		fmt.Println(TryAgain)
		ip, err = reader.ReadString('\n')
	}
	singlePlayer := false
	connect(ip[:len(ip)-1], singlePlayer)
}

// client3 starts the server, connects the player and creates a bot.
func client3() {
	wg.Add(1)
	go startServer()
	wg.Wait()
	port := server.Addr().String()
	idx := strings.LastIndex(port, ":")
	ip := "localhost" + port[idx:]
	wg.Add(1)
	singlePlayer := true
	go connect(ip, singlePlayer)
	wg.Wait()
	startBot(ip)
}

type bot struct {
	cardToPlay int
	points     int
}

// startBot creates and manages the bot.
func startBot(ip string) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println(err.Error()) // kill the app? log.fatal?
		return
	}

	conn.Write([]byte(Connect))
	p := make([]byte, 256)
	b := new(bot)

	for {
		size, err := conn.Read(p)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err.Error())
			}
			return
		}
		m := string(p)[:size]

		if strings.Contains(m, YourTurn) {
			b.cardToPlay = 0
			conn.Write([]byte(strconv.Itoa(b.cardToPlay)))
		} else if m == NotPossible || m == WrongInput {
			b.cardToPlay++
			conn.Write([]byte(strconv.Itoa(b.cardToPlay)))
		}
	}
}

// connect creates a client-server connection and communicates through it.
func connect(ip string, singlePlayer bool) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn.Write([]byte(Connect))

	p := make([]byte, 256)
	reader := bufio.NewReader(os.Stdin)
	var input string

	if singlePlayer {
		wg.Done()
	}

	for {
		size, err := conn.Read(p)
		if err != nil {
			if err == io.EOF {
				fmt.Print(OpponentLeft)
			} else {
				fmt.Println(err.Error())
			}
			return
		}
		m := string(p)[:size]
		fmt.Print(m)

		for strings.Contains(m, YourTurn) || m == WrongInput || m == NotPossible {
			input, err = reader.ReadString('\n')
			for err != nil {
				fmt.Println(TryAgain)
				input, err = reader.ReadString('\n')
			}

			if len(input) == 2 && input[0] >= '1' && input[0] <= '6' {
				conn.Write([]byte(input))
				break
			}

			input = strings.ToLower(input[:strings.Index(input, "\n")])
			switch input {
			case Close:
				conn.Write([]byte(Close))
			case Exchange:
				conn.Write([]byte(Exchange))
			case Stop:
				conn.Write([]byte(Stop))
			case Help:
				conn.Write([]byte(Help))
			case Quit:
				conn.Write([]byte(Quit))
				return
			default:
				fmt.Print(WrongInput)
				continue
			}

			break
		}

		if m == OpponentLeft {
			return
		}
	}
}

// main starts the game.
func main() {
	menu()
}
