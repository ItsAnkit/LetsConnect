package controllers

import (
	"time"
	"log"
	"github.com/gocql/gocql"
)

type MessagePayload struct {
	Id int 
	ConversationId int `json:"conversation_id"`
	CreatedAt time.Time
	Message string `json:"message"`
	SenderId int	`json:"sender_id"`
	// ReceiverId int `json:"receiver_id`
}

type TransientMessage struct {
	Id int
	MessageId int
}

func insertMessage(Session *gocql.Session, mp MessagePayload) bool {
	err := Session.Query("insert into messages(id, conversation_id, created_at, message, sender_id) values(?,?,?,?,?)", 
												mp.Id, mp.ConversationId, time.Now(), mp.Message, mp.SenderId).Exec()
	if err != nil {
		log.Println("Insertion Error: ", err)
		return false
	}
	return true
}

func lastMessageId(Session *gocql.Session) (int) {
	var maxId int
	err := Session.Query("select max(id) from messages").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
		return -1
	}
	return maxId
}

func lastTransientMessageId(Session *gocql.Session) (int) {
	var maxId int
	err := Session.Query("select max(id) from transient_messages").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
		return -1
	}
	return maxId
}

func saveAsTransientMessage(Session *gocql.Session, messageId int) bool {
	id := lastMessageId(Session) + 1
	if id <= 0 {
		return false
	}
	err := Session.Query("insert into transient_messages(id, created_at, message_id) values(?, ?, ?)", id, time.Now(), messageId).Exec()
	if err != nil {
		log.Println("Insertion Error: ", err)
		return false
	} 
	return true
}
