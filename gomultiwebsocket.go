package main

import (
	"fmt"
	"net/http"
	"log"
//	"reflect"
	"encoding/json"
	"code.google.com/p/go.net/websocket"
	"github.com/garyburd/redigo/redis"

)

// Global var holding redis connection
var redisConn redis.Conn

// Initialises program. 
// Opens connection to redis
func init() {

	c, err := redis.Dial("tcp", ":6379")
	if err!=nil {
		fmt.Println(err)
		panic(err)
	}
	redisConn=c
}

// Message structure for websockets.
// Type: kind of message. e.g. connect to channel, disconnect from channel, etc.
// Channel: which channel are we interacting with
// Data: message payload
type T struct {
	Type string
	Channel string
	Data string
}


type Post struct {
	Type string
	Date string
	Content string
	Link string
}


// Echo the data received on the WebSocket
func EchoServer(ws *websocket.Conn) {
	
	var msg string
	websocket.Message.Receive(ws, &msg)
	fmt.Println(msg)
	websocket.Message.Send(ws, msg)
}

// Handles requests from clients wishing to subscribe to channels
func SubServer(ws *websocket.Conn) {

	var msg T
	websocket.JSON.Receive(ws, &msg)
	fmt.Println(msg)
	if msg.Type == "connect" {
		fmt.Println("connect") // future implementation: possible authorisation for private channels
	} else if msg.Type == "subscribe"  {
		var channelExists bool = true //DoesChannelExist(msg.Channel)
		fmt.Println("channel exist check", channelExists)
		if channelExists == true {
			chann := make(chan string)
			psc := redis.PubSubConn{Conn: redisConn}
			psc.Subscribe(msg.Channel)
			defer psc.Unsubscribe()
			go SubscriberListener(psc, chann)
			SubscriberReporter(ws, chann)
			defer fmt.Println("end of the line bud")	
		} // else need to let client know to disconnect
	} else {
		msg.Type = "error"
		msg.Channel = ""
		msg.Data = "Corrupt message"
		websocket.JSON.Send(ws, msg)
	}
	
	
}

// Checks to see if event is publishing
func DoesChannelExist(chanName string) (r bool) {
	x, err := redisConn.Do("SISMEMBER", "pubsub.channels", chanName)
	if err!= nil {
		panic(err)
	}
	r = x != int64(0) // Turns int -> bool (i.e. is x not equal to 0?)
	return r 
}

// Listens for messages from the channel
func SubscriberListener(conn redis.PubSubConn, chann chan string)  {
	for {
		switch n := conn.Receive().(type) {
			case redis.Message:
				fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
				chann <- string(n.Data[:])
			case redis.Subscription:
				 fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
            			if n.Count == 0 {
					conn.Unsubscribe()
					fmt.Println("Unsubscribed")
                			return
            			}
			case error:
				fmt.Printf("error: %v\n", n)
				conn.Unsubscribe()
            			return
		}
	}
}

// Collects messages from the channel and writes them to the wsocket
func SubscriberReporter(ws *websocket.Conn, chann chan string) {
	for {
	
	msgData := <- chann
	if msgData == "unsubscribe" {
		fmt.Println("Client to unsubscribe")
		break
	} else {
		msgDataJSON := &T{}
		json.Unmarshal([]byte(msgData), &msgDataJSON)
		fmt.Println(msgDataJSON.Type)
		websocket.JSON.Send(ws, msgDataJSON)
	}
	}

} 


func main() {
	http.Handle("/echo", websocket.Handler(EchoServer))
	http.Handle("/websockets/ws", websocket.Handler(SubServer))
	err := http.ListenAndServe(":9009", nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + "fatal" )
	}
 	defer redisConn.Close()
}

