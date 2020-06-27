package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"strconv"
	// "runtime"
	// "encoding/json"
	// "context"
	// "github.com/go-redis/redis/v8"
	// "github.com/d4l3k/go-pry/pry"
)

// upgrade a normal HTTP connection to websocket
var upgrader = websocket.Upgrader{
	// CheckOrigin returns true if the request from different origin is allowed or not,
	// if not then CORS error will be returned
	ReadBufferSize:  1024,
    WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message string  `json:"message"`
}

type User struct {
	Id int `json:"id"`
	MobileNo string `json:"mobile_no" binding:"required"`
	Username string `json:"username" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type Conversation struct {
	Id int `json:"id"`
	CreatorId int `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Participant struct {
	Id int `json:"id"`
	ConversationId int `json:"conversation_id"`
	UserId int `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatHistory struct {
	Conversation Conversation
	Friend User
	// Messages [] Message
}

var clients = make(map[*websocket.Conn] bool)

// Channels are the pipes that connect goroutines.
var broadcast = make(chan Message)

var Session *gocql.Session

// func init ()  {
// 	Session := startCassandraSession()
// 	log.Println("init..", Session)
// }

func main() {
	log.Println("Server Started .......")

	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/login", login)
	router.GET("/conversations/:id/chats", getChats)
	router.GET("/conversations/:id/search", searchContact)
	router.POST("/conversations/:id/add", addContact)

	router.GET("/ws", func(c *gin.Context){
		handleWebsocket(c.Writer, c.Request)
	})

	go deliverMessages()
	router.Run(":8050")
}

func login(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	Session := startCassandraSession()
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

func getChats(c *gin.Context) {
	currentUserId := c.Param("id")
	Session := startCassandraSession()
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

func searchContact(c *gin.Context) {
	mobile := c.Query("mobile_no")
	currentUserId := c.Param("id")
	Session := startCassandraSession()
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

func addContact(c *gin.Context) {
	currentUserId := c.Param("id")
	friendMobileNo := c.Query("friendMobile")
	Session := startCassandraSession()
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

func fetchUser(Session *gocql.Session, userId string)(bool, User) {
	var user User
	id, conv_err := strconv.Atoi(userId)
	if conv_err != nil {
		log.Println("String to Int Conversion Error: ", conv_err)
		return false, user
	}
	err := Session.Query("select * from users where id = ? limit 1 allow filtering", id).Scan(&user.Id, &user.CreatedAt, &user.MobileNo, &user.Username)
	if err != nil {
		log.Println("Fetch User Error: ", err, user, id)
		return false, user
	}
	return true, user
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

func createConversation(currentUserId string, friendMobileNo string, user User, Session *gocql.Session) (bool, Conversation, [] Participant) {
	var conv Conversation
	var participants = make([] Participant, 0)
	var success bool
	conv = Conversation{totalConversations(Session) + 1, user.Id, time.Now()}
	err := Session.Query("insert into conversations(id, creator_id, created_at) values(?, ? ,?)", conv.Id, conv.CreatorId, conv.CreatedAt).Exec()
	if err != nil {
		return false, conv, participants
	}
	success, participants = createParticipants(Session , conv, friendMobileNo, participants)
	return success, conv, participants
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

func startCassandraSession() *gocql.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "letsconnect"
	//cluster.Consistency = gocql.LocalQuorum
	Session, err := cluster.CreateSession()
	if err != nil {
		log.Println("Cassandra Session Error: ", err)
	}
	return Session
}

func createUser(user User, Session *gocql.Session) (bool, User) {
	fmt.Println("Creating User......")
	if mobileExists(user.MobileNo, Session) {
		iter := Session.Query("select * from users where username = ? and mobile_no = ? allow filtering", user.Username, user.MobileNo).Iter()
		if iter.NumRows() > 0 {
			iter.Scan(&user.Id, &user.CreatedAt, &user.MobileNo, &user.Username)
			return true, user
		}
		return false, user
	}
	user.Id = totalUsers(Session) + 1
	err := Session.Query("insert into users(id, mobile_no, username, created_at) VALUES(?, ?, ?, ?);", user.Id, user.MobileNo, user.Username, user.CreatedAt ).Exec()
	if err != nil {
		log.Println("\n Error Message: ", err)
		return false, user
	}
	return true, user
}

func mobileExists(mobile_no string, Session *gocql.Session) bool {
	mobile, conv_err := strconv.Atoi(mobile_no)
	if conv_err != nil {
		log.Println("Conversion Error: ", conv_err)
		return false
	}
	result := Session.Query("select * from users where mobile_no = ? allow filtering;", mobile).Iter().NumRows()
	if result == 0 {
		return false
	}
	return true
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

func fetchChatHistory(Session *gocql.Session, convId int, user User) (bool, Conversation, User) {
	var conversation Conversation
	var friend User
	var participant Participant
	err := Session.Query("select * from conversations where id = ?", convId).Scan(&conversation.Id, &conversation.CreatedAt, &conversation.CreatorId)
	if err != nil {
		log.Println("Conversation fetch error: ", err)
		return false, conversation, friend
	}
	success, friend := fetchOtherParticipant(Session, conversation, participant, user, friend)
	if !success {
		return false, conversation, friend
	}
	return true, conversation, friend
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

func totalUsers(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from users").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}

func totalConversations(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from conversations").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}

func totalParticipants(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from participants").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Handling Connections - Upgrade HTTP to Websockets")
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("Connection Handling: ", err)
	}

	// Close connection when function returns
	defer ws.Close()

	clients[ws] = true
	// queueMessage(ws)
	for {
		var msg Message
		err := ws.ReadJSON(&msg)

		if err != nil {
			// log.Println("Not Found: %v", err)
			delete(clients, ws)
			break
		}

		// userRecord, err := json.Marshal(User{Username: "Abc", Id: 1})
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// ctx := context.Background()
		// err = redisClient.Set(ctx, "Users", userRecord, 0).Err()
		// if err != nil {
		// 	fmt.Println(err)
		// }

		// Send message to broadcast channel.
		broadcast <- msg
	}
}

func deliverMessages() {
	fmt.Println("Deliverin Messages")
	for {
		msg := <- broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				// log.Println("Delivery Error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// func queueMessage(ws *Conn) {
// 	for {
// 		var msg Message
// 		err := ws.ReadJson(&msg)
// 		if err != nil {
// 			log.Println("Not Found: %v", err)
// 			delete(clients, ws)
// 			break
// 		}
// 		// Send message to broadcast channel.
// 		broadcast <- msg
// 	}
// }
