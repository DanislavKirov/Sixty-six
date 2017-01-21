// Not finished!
package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	var addr = flag.String("addr", "localhost:8081", "http service address")
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr}

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
		_, message, _ := c.ReadMessage()
		fmt.Println(string(message))
	}
}
