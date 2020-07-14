package controllers

import (
	"github.com/gorilla/websocket"
	"github.com/gocql/gocql"
	"log"
	"time"
	// "github.com/google/uuid"
	"encoding/json"
	"bytes"
	"strconv"
)

const (
	writewait = 100 * time.Second
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
	send  map[string](chan string)
	session *gocql.Session
	messagePayload MessagePayload
	friendId int
}

func CreateNewSocket(conn *websocket.Conn, hub *Hub, msg MessagePayload, Session *gocql.Session) {
	// id := uuid.New()
	success, friend := fetchOtherParticipant(Session, msg.ConversationId, msg.SenderId)
	if !success {
		return
	}
	client := &Client{
		hub: hub,
		wsConnection: conn,
		send: make(map[string](chan string)),
		session: Session,
		messagePayload: msg,
		friendId: friend.Id,
	}
	log.Println("client", client, hub.clients)
	go client.readMessage()
	go client.writeMessage()
	client.hub.subscribe <- client
}

func (client *Client) readMessage() {
	var socketEvent SocketEvent
	// defer closeSocketConnection(client)
	// setWebSocketConfig(client)
	log.Println("Reading Message ++++++++++++++++++++=")
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
		log.Println("read1 : ", socketEvent)
		handleSocketPayloadEvent(client, socketEvent)
	}
}

func (client *Client) writeMessage() {
	ticker := time.NewTicker(pingPeriod)
	log.Println("Writing Message ++++++++++++++++++++=")
	defer func() {
		ticker.Stop()
		// client.wsConnection.Close()
	}()
	// writeDeadline := time.Now().Add(writewait)
	channelId :=  strconv.Itoa(client.friendId) + strconv.Itoa(client.messagePayload.SenderId) 
	log.Println("Write: ", client.friendId, channelId, client.send, client.messagePayload.Message)
	for {
		select {
			case payload, success := <- client.send[channelId]:
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
				n := len(client.send[channelId])
				for i := 0; i < n; i++ {
					json.NewEncoder(requestBytes).Encode(<-client.send[channelId])
					writer.Write(requestBytes.Bytes())
				}
				log.Println("Write2: ", writer, n)
				if err := writer.Close(); err != nil {
					return
				}

			case <-ticker.C: // Receiving from channel to which tickers are delivered.
				// client.wsConnection.SetWriteDeadline(writeDeadline)
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
	log.Println("Socket Payload ++++++++++++++++++++")
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
	messageSend := false
	messagePayload := socketEvent.EventPayload.(MessagePayload)
	channelId := strconv.Itoa(messagePayload.SenderId) + strconv.Itoa(client.friendId)
	log.Println("send_to", client.friendId, channelId, client.send, client.messagePayload.Message)
	for client := range hub.clients {
		if client.messagePayload.SenderId == friendId {
			client.send[channelId] = make(chan string, 5) 
			select {
				case client.send[channelId] <- messagePayload.Message:
					messageSend = true
					log.Println("send_to found....", friendId, messagePayload.SenderId, channelId, client.send)
				default:
					if channelId == "" {
						close(client.send[channelId])
					}
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
