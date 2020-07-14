package controllers

import (
	"log"
	// "strconv"
)

type Hub struct {
	clients map[*Client] bool
	subscribe chan *Client
	unsubscribe chan *Client
}

func HubInit() *Hub {
	return &Hub {
		clients: make(map[*Client] bool),
		subscribe: make(chan *Client),
		unsubscribe: make(chan *Client),
	}
}

func (hub *Hub) Run() {
	for {
		select {
			case client := <- hub.subscribe: //Receive data from subscribe channel
				subscribeUser(hub, client)
				log.Println("hub...", hub.subscribe)
			case client := <- hub.unsubscribe: //Receive data from unsubscribe channel
				unsubscribeUser(hub, client)
		}
	}
}

func subscribeUser(hub *Hub, client *Client) {
	log.Println("sub1 ...........", hub.clients)
	hub.clients[client] = true
	log.Println("sub2 ...........", hub.clients)
	socketEvent := SocketEvent{"message", client.messagePayload }
	handleSocketPayloadEvent(client, socketEvent)
}

func unsubscribeUser(hub *Hub, client *Client) {
	_, success := hub.clients[client]
	log.Println("+++++++++++    unsubscribing    +++++++++++++")
	if success {
		// channelId := strconv.Itoa(client.messagePayload.SenderId) + strconv.Itoa(client.friendId)
		// delete(hub.clients, client)
		// close(client.send[channelId])
		// socketEvent := SocketEvent{"disconnect", client.userId}
		// handleSocketPayloadEvent(client, socketEvent)
	}
}
