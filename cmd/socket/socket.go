package main

// start websocket server and listen for incoming connections

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
)

// define websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// define client connection
type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

// define message
type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type MessageBody struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

// define pool of clients
type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

// define pool instance
var pool = Pool{
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
	Broadcast:  make(chan Message),
}

// define pool method to start listening for incoming messages
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client, _ := range pool.Clients {
				fmt.Println(client)
				//client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			//for client, _ := range pool.Clients {
			//client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
			//}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client, _ := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
			break
		}
	}
}

// define client method to listen for incoming messages
func (c *Client) Read() {
	defer func() {
		pool.Unregister <- c
		c.Conn.Close()
	}()
	for {
		//messageType, p, err := c.Conn.ReadMessage()
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		body := MessageBody{}

		c.Conn.ReadJSON(&body)

		//pool.Broadcast <- message

		keyboardEvent(body.Key)
		fmt.Printf("Message Received: %+v\n", body)
	}
}

// define client method to write messages to client
func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message := <-c.Pool.Broadcast:
			fmt.Printf("Sending message to client: %+v\n", message)
			if err := c.Conn.WriteJSON(message); err != nil {
				fmt.Println(err)
				return
			}
			break
		}
	}
}

// define websocket endpoint
func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	// generate unique id for client
	client := &Client{
		ID:   "1",
		Conn: conn,
		Pool: pool,
	}

	// register client with pool
	pool.Register <- client
	client.Read()
}

func fadeLeds(leds [][3]byte) {
	length := len(leds)
	for i := 0; i < length; i++ {
		for j := 0; j < 3; j++ {
			if leds[i][j] > 10 {
				leds[i][j] -= 10
			} else {
				leds[i][j] = 0
			}
		}
	}
}

func setLed(leds [][3]byte, color [3]byte) {
	for i := 0; i < len(leds); i++ {
		leds[i] = color
	}
}

func setLedRange(leds [][3]byte, color [3]byte, start int, end int) {
	for i := start; i < end; i++ {
		leds[i] = color
	}
}

func keyboardEvent(key string) {
	if key == "ArrowLeft" {
		setLedRange(leds, [3]byte{255, 255, 255}, 0, 32)
	}
	if key == "ArrowRight" {
		setLedRange(leds, [3]byte{255, 255, 255}, 64, 96)
	}
	if key == "ArrowUp" {
		setLedRange(leds, [3]byte{255, 255, 255}, 32, 64)
	}
	if key == "ArrowDown" {
		setLed(leds, [3]byte{255, 255, 255})
	}
	if key == "k" {
		setLedRange(leds, [3]byte{255, 255, 0}, 0, 32)
	}
	if key == "l" {
		setLedRange(leds, [3]byte{255, 255, 0}, 64, 96)
	}

}

var leds [][3]byte

// define main function
func main() {
	leds = make([][3]byte, 97)

	err := config.ParseEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}
	flags := config.GetFlags()
	connectionc := connection.StartConnection(*flags.Ip, *flags.Port)
	fps := 60
	//leds := make([][3]byte, 97)
	go func() {
		for {
			start := time.Now()

			fadeLeds(leds)
			fmt.Println(leds[0])

			connection.SendUdpPacket(connectionc, leds)
			for time.Since(start) < time.Duration(1000/fps)*time.Millisecond {
			}
		}
	}()

	// start listening for incoming messages
	go pool.Start()

	// define websocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(&pool, w, r)
	})

	// start http server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("http server started on :8080")
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
