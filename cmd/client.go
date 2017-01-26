package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func menu() {
	choice := 0
	reader := bufio.NewReader(os.Stdin)
	for choice != 1 && choice != 2 {
		fmt.Print("Pick one:\n1. Create game\n2. Join game\nYour choise: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(TryAgain)
		} else {
			choice = int(input[0] - '0')
		}
	}
	if choice == 1 {
		client1()
	} else {
		client2()
	}
}

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

func client1() {
	go startServer()
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	port := server.Addr().String()
	idx := strings.LastIndex(port, ":")
	fmt.Println(ip + port[idx:])
	connect("localhost" + port[idx:])
}

func client2() {
	fmt.Print("Enter ip:port: ")
	reader := bufio.NewReader(os.Stdin)
	ip, err := reader.ReadString('\n')
	for err != nil {
		fmt.Println(TryAgain)
		ip, err = reader.ReadString('\n')
	}
	connect(ip[:len(ip)-1])
}

func connect(ip string) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn.Write([]byte(Connect))

	p := make([]byte, 128)
	reader := bufio.NewReader(os.Stdin)
	var input string
	for {
		size, err := conn.Read(p)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		m := string(p)[:size]
		fmt.Print(m)

		for strings.Contains(m, YourTurn) || m == WrongInput {
			input, err = reader.ReadString('\n')
			for err != nil {
				fmt.Println(TryAgain)
				input, err = reader.ReadString('\n')
			}

			if input[0] >= '1' && input[0] <= '6' {
				conn.Write([]byte(input))
			} else if strings.Contains(strings.ToLower(input), Close) {
				conn.Write([]byte(Close))
			} else if strings.Contains(strings.ToLower(input), Quit) {
				conn.Write([]byte(Quit))
			} else if strings.Contains(strings.ToLower(input), Exchange) {
				conn.Write([]byte(strings.ToLower(input)))
			} else {
				fmt.Print(WrongInput)
				continue
			}
			break
		}

		if m == EnoughPlayers || m == OpponentLeft {
			return
		}
	}
}

func main() {
	menu()
}
