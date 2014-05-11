package main

import (
	"fmt"
	"net/http"
//	"io"
//	"bytes"
//	"log"
//	"bufio"
	"code.google.com/p/go.net/websocket"
)

// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {

	var msg string
	 websocket.Message.Receive(ws, &msg)
	msg = "hiya"
	fmt.Println(msg)
	websocket.Message.Send(ws, msg)
}

func main() {
	http.Handle("/echo", websocket.Handler(EchoServer))
	http.ListenAndServe(":9009", nil)
 
}

