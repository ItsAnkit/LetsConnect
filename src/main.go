package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/websocket"
	// "time"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	// "encoding/json"
	// "context"
	// "github.com/go-redis/redis/v8"
	// "github.com/d4l3k/go-pry/pry"
)

var clients = make(map[*websocket.Conn] bool)

// Channels are the pipes that connect goroutines.
var broadcast = make(chan Message)

// upgrade a normal HTTP connection to websocket
var upgrader = websocket.Upgrader{
	// CheckOrigin returns true if the request from different origin is allowed or not,
	// if not then CORS error will be returned
	ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


// var redisClient = redis.NewClient(&redis.Options{
// 	Addr: "localhost:6379",
// 	Password: "",
// 	DB: 0, // use default DB
// })
type Message struct {
	Username string `json:"username"`
	Message string  `json:"message"`
}

// type User struct {
// 	Username string
// 	Created time.Time
// 	Id int64 `json:"ref"`
// }

// type Conversation struct {
// 	Id int64
// 	Sender string
// 	Created time.Time
// 	Message string
// }

func main() {
	fmt.Println("Started .......")

	router := gin.Default()
	log.Println(router)

	router.Use(cors.Default())

	router.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	router.GET("/ws", func(c *gin.Context){
		handleWebsocket(c.Writer, c.Request)
	})
	go deliverMessages()

	fmt.Println("gin !!!")
	router.Run(":8050")
}	

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling Connections - Upgrade HTTP to Websockets")
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("Connection Handling: ", err)
	}

	// Close connection when function returns
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)

		if err != nil {
			// log.Println("Not Found: %v", err)
			delete(clients, ws)
			break
		}

		broadcast <- msg
	}
}

func deliverMessages() {
	fmt.Println("Deliverin Messages")
	for {
		msg := <- broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				// log.Println("Delivery Error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// func queueMessage(ws *Conn) {
// 	for {
// 		var msg Message
// 		err := ws.ReadJson(&msg)
// 		if err != nil {
// 			log.Println("Not Found: %v", err)
// 			delete(clients, ws)
// 			break
// 		}
// 		// Send message to broadcast channel.
// 		broadcast <- msg
// 	}
// }
