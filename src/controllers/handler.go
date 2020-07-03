package controllers

import (
	"log"
	"app/db"
	"github.com/gin-gonic/gin"
	"time"
)

func Login(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	Session := db.StartCassandraSession()
	defer Session.Close()
	var success bool
	user.CreatedAt = time.Now()
	success, user = createUser(user, Session)
	var chats = make([] ChatHistory, 0)
	if success {
		success, chats = fetchConversations(user, Session, chats)
	}
	c.JSON(200, gin.H{
		"success": success,
		"user": user,
		"chats": chats,
	})
}

func GetChats(c *gin.Context) {
	currentUserId := c.Param("id")
	Session := db.StartCassandraSession()
	defer Session.Close()
	success, user := fetchUser(Session, currentUserId)
	var chats = make([] ChatHistory, 0)
	if success {
		success, chats = fetchConversations(user, Session, chats)
	}
	c.JSON(200, gin.H{
		"success": success,
		"chats": chats,
	})
}

func SearchContact(c *gin.Context) {
	mobile := c.Query("mobile_no")
	currentUserId := c.Param("id")
	Session := db.StartCassandraSession()
	var success, isContact bool
	var message string = ""
	defer Session.Close()
	success = mobileExists(mobile, Session)
	if success {
		isContact = isFriend(mobile, Session, currentUserId)
		if isContact {
			message = "User already your friend."
		}
	} else {
		message = "User doesn't exist."
	}
	c.JSON(200, gin.H{
		"success": success,
		"isContact": isContact,
		"message": message,
	})
}

func AddContact(c *gin.Context) {
	currentUserId := c.Param("id")
	friendMobileNo := c.Query("friendMobile")
	Session := db.StartCassandraSession()
	defer Session.Close()
	success, user := fetchUser(Session, currentUserId)
	if success {
		success, conversation, participants := createConversation(currentUserId, friendMobileNo, user, Session)
		log.Println("Convo....", conversation, participants, success)
	}
	var chats = make([] ChatHistory, 0)
	if success {
		success, chats = fetchConversations(user, Session, chats)
	}
	c.JSON(200, gin.H{
		"success": success,
		"chats": chats,
	})
}

func HandleMessage(c *gin.Context) {
	handleWebsocket(c.Writer, c.Request)
	go deliverMessages()
}
