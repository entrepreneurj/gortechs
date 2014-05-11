package main

import (
	"fmt"
	"net/http"
	"log"
	"time"
//	"reflect"
	"encoding/json"
	"code.google.com/p/go.net/websocket"
	"github.com/garyburd/redigo/redis"

)

// Global var holding redis connection

var (
	pool *redis.Pool // Redis connection pool
)


func newPool() *redis.Pool {
	return &redis.Pool{
		MaxActive: 0,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
		       	if err!=nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
           		 _, err := c.Do("PING")
            		return err
        	},
	}
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
	redisConn := pool.Get()
	defer redisConn.Close()

	var msg T
	websocket.JSON.Receive(ws, &msg)
	fmt.Println(msg)
	if msg.Type == "connect" {
		fmt.Println("connect") // future implementation: possible authorisation for private channels
	} else if msg.Type == "subscribe"  {
		var channelExists bool = true //DoesChannelExist(redisConn, msg.Channel)
		fmt.Println("channel exist check", channelExists)
		if channelExists == true {
			chann := make(chan string,5)
			psc := redis.PubSubConn{Conn: redisConn}
			psc.Subscribe(msg.Channel)
			go SubscriberListener(psc, chann)
			SubscriberReporter(ws, psc, chann)
			return	
		} // else need to let client know to disconnect
	} else {
		msg.Type = "error"
		msg.Channel = ""
		msg.Data = "Corrupt message"
		websocket.JSON.Send(ws, msg)
	}
	
	
}

// Checks to see if event is publishing
func DoesChannelExist(conn redis.Conn, chanName string) (r bool) {
	x, err := conn.Do("SISMEMBER", "pubsub.channels", chanName)
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
					fmt.Println(conn.Unsubscribe())
					fmt.Println("Unsubscribed")
            			}
			case error:
				fmt.Printf("error: %v\n", n)
				conn.Unsubscribe()
            			return
		}
	}
}

// Collects messages from the channel and writes them to the wsocket
func SubscriberReporter(ws *websocket.Conn, conn redis.PubSubConn, chann chan string) {
	for {
	
	msgData := <- chann
	if msgData == "unsubscribe" {
		fmt.Println("Client to unsubscribe")
		fmt.Println(conn.Unsubscribe())
		break
	} else {
		msgDataJSON := &T{}
		json.Unmarshal([]byte(msgData), &msgDataJSON)
		fmt.Println(msgDataJSON)
		websocket.JSON.Send(ws, msgDataJSON)
	}
	}

} 


func main() {
	pool = newPool()
	http.Handle("/echo", websocket.Handler(EchoServer))
	http.Handle("/websockets/ws", websocket.Handler(SubServer))
	err := http.ListenAndServe(":9009", nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + "fatal" )
	}
}

