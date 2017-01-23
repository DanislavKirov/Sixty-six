// Not finished!
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

const yourTurn = "It's your turn, pick a card index: "

func main() {
	var addr = flag.String("addr", "localhost:8081", "http service address")
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "connect"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = c.WriteMessage(websocket.TextMessage, []byte("connect"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		_, message, e := c.ReadMessage()
		if e != nil {
			fmt.Println(e.Error())
			return
		}
		m := string(message)
		fmt.Println(m)
		if m == yourTurn || m == "wrong input, try again: " {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			c.WriteMessage(websocket.TextMessage, []byte(text))
		}
	}
}
