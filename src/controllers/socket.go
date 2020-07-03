package controllers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// CheckOrigin returns true if the request from different origin is allowed or not,
// if not then CORS error will be returned
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
  WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Channels are the pipes that connect goroutines.
var broadcast = make(chan Message)

var clients = make(map[*websocket.Conn] bool)

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	log.Println("Handling Connections - Upgrade HTTP to Websockets")
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
		log.Println("Message.., ", msg)
		if err != nil {
			log.Println("Not Found: ", err)
			delete(clients, ws)
			break
		}
		// Send message to broadcast channel.
		broadcast <- msg
	}
}

func deliverMessages() {
	log.Println("Deliverin Messages")
	for {
		msg := <- broadcast
		log.Println("yokas", msg, clients)
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("Delivery Error: ", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
