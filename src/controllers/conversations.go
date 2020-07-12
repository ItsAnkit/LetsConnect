package controllers

import (
	"github.com/gocql/gocql"
	"log"
	"time"
	"strconv"
)

type Conversation struct {
	Id int `json:"id"`
	CreatorId int `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

func createConversation(currentUserId string, friendMobileNo string, user User, Session *gocql.Session) (bool, Conversation, [] Participant) {
	var conv Conversation
	var participants = make([] Participant, 0)
	var success bool
	conv = Conversation{lastConversationId(Session) + 1, user.Id, time.Now()}
	err := Session.Query("insert into conversations(id, creator_id, created_at) values(?, ? ,?)", conv.Id, conv.CreatorId, conv.CreatedAt).Exec()
	if err != nil {
		return false, conv, participants
	}
	success, participants = createParticipants(Session , conv, friendMobileNo, participants)
	return success, conv, participants
}

func fetchConversations(user User, Session *gocql.Session, chats []ChatHistory) (bool, [] ChatHistory) {
	convIter := Session.Query("select conversation_id from participants where user_id = ? allow filtering;", user.Id).Iter()
	var convId int
	for convIter.Scan(&convId) {
		success, conv, friend := fetchChatHistory(Session, convId, user)
		if !success {
			return false, chats
		}
		chat := ChatHistory{conv, friend}
		chats = append(chats, chat)
	}
	return true, chats
}

func isFriend(mobile_no string, Session *gocql.Session, userId string) (bool) {
	var friendId int
	var convId int
	mobile, conv_err1 := strconv.Atoi(mobile_no)
	currentUserId, conv_err2 := strconv.Atoi(userId)
	if conv_err1 != nil || conv_err2 != nil {
		log.Println("Conversion Error: ", conv_err1, conv_err2)
		return false
	}
	err := Session.Query("select id from users where mobile_no = ? allow filtering", mobile).Scan(&friendId)
	if err != nil {
		log.Println("Fetch Friend Id: ", err)
		return false
	}
	iter := Session.Query("select conversation_id from participants where user_id = ? allow filtering", friendId).Iter()
	for iter.Scan(&convId) {
		rows := Session.Query("select * from participants where user_id = ? and conversation_id = ? allow filtering", currentUserId, convId).Iter().NumRows()
		if rows > 0 {
			return true
		}
	}
	return false
}

func lastConversationId(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from conversations").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}
