package controllers

import (
	"log"
	"github.com/gocql/gocql"
	"time"
	"strconv"
)

type Participant struct {
	Id int `json:"id"`
	ConversationId int `json:"conversation_id"`
	UserId int `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func createParticipants(Session *gocql.Session, conv Conversation, friendMobileNo string, participants []Participant) (bool, [] Participant) {
	total := totalParticipants(Session)
	success1, p1 := insertCreatorParticipant(Session, total, conv)
	success2, p2 := insertFriendParticipant(Session, total, conv, friendMobileNo)
	if !success1 || !success2 {
		return false, participants
	}
	participants = append(participants, p1)
	participants = append(participants, p2)
	return true, participants
}

func insertCreatorParticipant(Session *gocql.Session, total int, conv Conversation) (bool, Participant) {
	p := Participant{total + 1, conv.Id, conv.CreatorId, time.Now()}
	err := Session.Query("insert into participants(id, conversation_id, user_id, created_at) values(?,?,?,?)", p.Id, p.ConversationId, p.UserId, time.Now()).Exec()
	if err != nil {
		return false, p
	}
	return true, p
}

func insertFriendParticipant(Session *gocql.Session, total int, conv Conversation, friendMobileNo string) (bool, Participant) {
	var friendId int
	var p Participant
	friendMobile, conv_err := strconv.Atoi(friendMobileNo)
	if conv_err != nil {
		log.Println("Conversion Error: ", conv_err)
		return false, p
	}
	err := Session.Query("select id from users where mobile_no = ? limit 1 allow filtering", friendMobile).Scan(&friendId)
	p = Participant{total + 2, conv.Id, friendId, time.Now()}
	err = Session.Query("insert into participants(id, conversation_id, user_id, created_at) values(?,?,?,?)", p.Id, p.ConversationId, p.UserId, time.Now()).Exec()
	if err != nil {
		return false, p
	}
	return true, p
}

func fetchOtherParticipant(Session *gocql.Session, conv Conversation, participant Participant, currentUser User, friend User) (bool, User) {
	iter := Session.Query("select * from participants where conversation_id = ? allow filtering;", conv.Id).Iter()
	for iter.Scan(&participant.Id, &participant.ConversationId, &participant.CreatedAt, &participant.UserId) {
		if participant.UserId != currentUser.Id {
			err := Session.Query("select * from users where id = ?", participant.UserId).Scan(&friend.Id, &friend.CreatedAt, &friend.MobileNo, &friend.Username)
			if err != nil {
				log.Println("User fetch error: ", err)
				return false, friend
			}
			return true, friend
		}
	}
	return false, friend
}

func totalParticipants(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from participants").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}