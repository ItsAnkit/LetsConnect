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
	// send  map[string](chan string)
	session *gocql.Session
	messagePayload MessagePayload
	friendId int
}

func CreateNewSocket(conn *websocket.Conn, hub *Hub, Session *gocql.Session, messagePayload MessagePayload) {
	// id := uuid.New()
	success, friend := fetchOtherParticipant(Session, messagePayload.ConversationId, messagePayload.SenderId)
	if !success {
		return
	}
	client := &Client{
		hub: hub,
		wsConnection: conn,
		session: Session,
		messagePayload: messagePayload,
		friendId: friend.Id,
	}
	log.Println("client", client, hub.clients)
	if !client.hub.clients[client] {
		log.Println("subscribing!!!")
		client.hub.subscribe <- client
	}
	go client.writeMessage()
	go client.readMessage()
}

func (client *Client) readMessage() {
	// defer closeSocketConnection(client)
	// setWebSocketConfig(client)
	log.Println("Reading Message ++++++++++++++++++++", client.wsConnection, "\n likerer", client.messagePayload )
	for {
		log.Println("reading started...")
		x, payload, wsErr := client.wsConnection.ReadMessage()
		log.Println("Payload +++ ", payload)
		if wsErr != nil {
			log.Println("Socket Read Error: ", wsErr, x)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(payload))
		err := decoder.Decode(&client.messagePayload)
		log.Println("decoder ", decoder, "payload ", payload)
		if err != nil {
			log.Println("Decoding error: ", err)
			return
		}
		log.Println("read1 : ", client.messagePayload)
		handleSocketPayloadEvent(client)
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
	channelId :=  strconv.Itoa(client.friendId) + "-" + strconv.Itoa(client.messagePayload.SenderId) 
	log.Println("Write: ", client.friendId, channelId, client.hub.send, client.messagePayload.Message)
	for {
		select {
			case payload, success := <- client.hub.send[channelId]:
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
				n := len(client.hub.send[channelId])
				for i := 0; i < n; i++ {
					json.NewEncoder(requestBytes).Encode(<-client.hub.send[channelId])
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

func handleSocketPayloadEvent(client *Client) {
	log.Println("handleSocketPayloadEvent ++++++++++++++++++++")
	id := lastMessageId(client.session) + 1
	// messagePayload := MessagePayload{id, conversationId, time.Now(), message, senderId}
	// messagePayload := client.messagePayload
	client.messagePayload.Id = id
	success := insertMessage(client.session, client.messagePayload)
	log.Println("Private Chat1 ", client.messagePayload, success)
	if success {
		// messagePacket := SocketEvent{"message", messagePayload}
		success, friend := fetchOtherParticipant(client.session, client.messagePayload.ConversationId, client.messagePayload.SenderId)
		if success {
			SendToSpecificClient(client, client.messagePayload, friend.Id)
		}
	}
}

func SendToSpecificClient(client *Client, messagePayload MessagePayload, friendId int) {
	hub := client.hub
	messageSend := false
	// messagePayload := socketEvent.EventPayload.(MessagePayload)
	channelId := strconv.Itoa(messagePayload.SenderId) + "-" + strconv.Itoa(friendId)
	log.Println("send_to", client.friendId, channelId, hub.send, messagePayload.Message)
	for client := range hub.clients {
		if (client.messagePayload.SenderId == friendId && len(client.messagePayload.Message) > 0) {
			hub.send[channelId] = make(chan string, 10) 
			select {
				case hub.send[channelId] <- messagePayload.Message:
					messageSend = true
					log.Println("send_to found....", friendId, messagePayload.SenderId, channelId, hub.send)
				default:
					if channelId == "" {
						close(hub.send[channelId])
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
