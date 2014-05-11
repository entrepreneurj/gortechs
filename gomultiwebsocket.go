package main

import (
	"fmt"
	"net/http"
//	"io"
//	"bytes"
//	"log"
	"bufio"
	"code.google.com/p/go.net/websocket"
)

// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {
//	buf := new(bytes.Buffer)
//	buf.ReadFrom(ws)
	bufReader := bufio.NewReader(ws)
	msg, err := bufReader.ReadString('\n')
	fmt.Println(msg)
	if err != nil {
		panic(err)
	}
//	io.WriteString(ws, buf.String())
//	io.Copy(ws, ws)
//	fmt.Println(buf.String())
//	if _, err := ws.Write([]byte(buf.String())); err != nil {
//    		log.Fatal(err)
//		fmt.Println("Error")
//	}
//	fmt.Println(buf.String())
	ws.Write([]byte("hi"))
}

func main() {
	http.Handle("/echo", websocket.Handler(EchoServer))
	http.ListenAndServe(":9009", nil)
 
}

