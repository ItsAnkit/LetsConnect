package controllers

import (
	"github.com/gorilla/websocket"
	"github.com/gocql/gocql"
	"log"
	"time"
	// "github.com/google/uuid"
	"encoding/json"
	"bytes"
)

const (
	writewait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9)/10
	messageSize = 512
)

type SocketEvent struct {
	EventName string
	EventPayload interface{} 
}

type Client struct {
	hub *Hub
	wsConnection *websocket.Conn
	send chan string
	session *gocql.Session
	messagePayload MessagePayload
}

func CreateNewSocket(conn *websocket.Conn, hub *Hub, msg MessagePayload, Session *gocql.Session) {
	// id := uuid.New()
	client := &Client{
		hub: hub,
		wsConnection: conn,
		send: make(chan string),
		session: Session,
		messagePayload: msg,
	}
	log.Println("client", client, hub.clients)
	go client.writeMessage()
	go client.readMessage()
	client.hub.subscribe <- client
}

func (client *Client) readMessage() {
	var socketEvent SocketEvent
	defer closeSocketConnection(client)
	// setWebSocketConfig(client)

	for {
		x, payload, wsErr := client.wsConnection.ReadMessage()

		if wsErr != nil {
			log.Println("Socket Read Error: ", wsErr, payload, x)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(payload))
		err := decoder.Decode(&socketEvent)

		if err != nil {
			log.Println("Decoding error: ", err)
			return
		}
		log.Println("read1 : ", socketEvent, decoder)
		handleSocketPayloadEvent(client, socketEvent)
	}
}

func (client *Client) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.wsConnection.Close()
	}()
	writeDeadline := time.Now().Add(writewait)
	log.Println("Write: ", ticker)
	for {
		select {
			case payload, success := <- client.send:
				requestBytes := new(bytes.Buffer)
				json.NewEncoder(requestBytes).Encode(payload)
				finalPayload := requestBytes.Bytes()
				// client.wsConnection.SetWriteDeadline(writeDeadline)
				log.Println("Write1: ", finalPayload, success, requestBytes)
				if !success {
					client.wsConnection.WriteMessage(websocket.CloseMessage, []byte {})
					return
				}

				writer, err := client.wsConnection.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Println("Write Error: ", err)
					return
				}
				writer.Write(finalPayload)
				n := len(client.send)
				for i := 0; i < n; i++ {
					json.NewEncoder(requestBytes).Encode(<-client.send)
					writer.Write(requestBytes.Bytes())
				}
				log.Println("Write2: ", writer, n)
				if err := writer.Close(); err != nil {
					return
				}

			case <-ticker.C: // Receiving from channel to which tickers are delivered.
				client.wsConnection.SetWriteDeadline(writeDeadline)
				err := client.wsConnection.WriteMessage(websocket.PingMessage, nil)
				log.Println("Write3: ", client, err)
				if err != nil {
					return
				}
		}
	}
}

func closeSocketConnection(client *Client) {
	client.hub.unsubscribe <- client
	client.wsConnection.Close()
}

func setWebSocketConfig(client *Client) {
	deadline := time.Now().Add(pongWait)
	client.wsConnection.SetReadLimit(messageSize)
	client.wsConnection.SetReadDeadline(deadline)
	client.wsConnection.SetPongHandler(func (string) error {
		client.wsConnection.SetReadDeadline(deadline)
		return nil
	})
}

func handleSocketPayloadEvent(client *Client, socketEvent SocketEvent) {
	switch socketEvent.EventName {
		// case "join":

		case "message":
			messagePayload := socketEvent.EventPayload.(MessagePayload)
			if true {
				id := lastMessageId(client.session) + 1
				// messagePayload := MessagePayload{id, conversationId, time.Now(), message, senderId}
				messagePayload.Id = id
				success := insertMessage(client.session, messagePayload)
				log.Println("Private Chat1 ", socketEvent.EventPayload, success)
				if success {
					messagePacket := SocketEvent{"message", messagePayload}
					success, friend := fetchOtherParticipant(client.session, messagePayload.ConversationId, messagePayload.SenderId)
					if success {
						SendToSpecificClient(client, messagePacket, friend.Id)
					}
				}
			}
	}
}

func SendToSpecificClient(client *Client, socketEvent SocketEvent, friendId int) {
	hub := client.hub
	log.Println("send_to", friendId, hub.clients)
	messageSend := false
	messagePayload := socketEvent.EventPayload.(MessagePayload)
	for client := range hub.clients {
		if client.messagePayload.SenderId == friendId {
			log.Println("send_to found....", friendId, messagePayload.SenderId)
			select {
				case client.send <- messagePayload.Message:
					messageSend = true
				default:
					close(client.send)
					// delete(hub.clients, client)
			}
		}
	}
	if !messageSend {
		saveAsTransientMessage(client.session, messagePayload.Id)
	}
	return
}

// func BroadcastToAllClient(hub *Hub, socketEvent SocketEvent, userId int) {
// 	for client := range hub.clients {
// 		select {
// 			case client.send <- socketEvent:
// 			default:
// 				close(client.send)
// 				delete(hub.clients, client)
// 		}
// 	}
// }

// func handleWebsocket(w http.ResponseWriter, req *http.Request) {
// 	ws, err := upgrader.Upgrade(w, req, nil)
// 	if err != nil {
// 		log.Fatal("Connection Handling: ", err)
// 	}
// 	// Close connection when function returns
// 	defer ws.Close()

// 	clients[ws] = true
// 	for {
// 		var msg Message
// 		err := ws.ReadJSON(&msg)
// 		log.Println("Message.., ", msg)
// 		if err != nil {
// 			log.Println("Not Found: ", err)
// 			delete(clients, ws)
// 			break
// 		}
// 		// Send message to broadcast channel.
// 		broadcast <- msg
// 	}
// }

// func deliverMessages() {
// 	log.Println("Deliverin Messages")
// 	for {
// 		msg := <- broadcast
// 		log.Println("yokas", msg, clients)
// 		for client := range clients {
// 			err := client.WriteJSON(msg)
// 			if err != nil {
// 				log.Println("Delivery Error: ", err)
// 				client.Close()
// 				delete(clients, client)
// 			}
// 		}
// 	}
// }
