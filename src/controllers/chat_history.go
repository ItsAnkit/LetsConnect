package controllers

import (
	"log"
	"github.com/gocql/gocql"
	// "github.com/gorilla/websocket"
	// "time"
)

type ChatHistory struct {
	Conversation Conversation `json:"conversation"`
	Friend User `json:"friend"`
	// Messages [] Message
}

func fetchChatHistory(Session *gocql.Session, convId int, user User) (bool, Conversation, User) {
	var conversation Conversation
	// var friend User
	// var participant Participant
	err := Session.Query("select * from conversations where id = ?", convId).Scan(&conversation.Id, &conversation.CreatedAt, &conversation.CreatorId)
	if err != nil {
		log.Println("Conversation fetch error: ", err)
		return false, conversation, User{}
	}
	success, friend := fetchOtherParticipant(Session, conversation.Id, user.Id)
	if !success {
		return false, conversation, friend
	}
	return true, conversation, friend
}
