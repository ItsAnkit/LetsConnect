package controllers

import (
	"log"
	"app/auth"
	"github.com/gin-gonic/gin"
	"time"
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/gocql/gocql"
	// "strconv"
)

// CheckOrigin returns true if the request from different origin is allowed or not,
// if not then CORS error will be returned
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
  WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Login(Session *gocql.Session) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var user User
		c.BindJSON(&user)
		// Session := db.StartCassandraSession()
		// defer Session.Close()
		var success bool
		var token string
		user.CreatedAt = time.Now()
		success, user = createUser(user, Session)
		var chats = make([] ChatHistory, 0)
		if success {
			success, token = auth.CreateToken(user.Id)
			if success {
				success, chats = fetchConversations(user, Session, chats)
			}
		}
		c.JSON(200, gin.H{
			"success": success,
			"user": user,
			"chats": chats,
			"auth-token": token,
		})
	}
	return gin.HandlerFunc(fn)
}

func GetChats(Session *gocql.Session) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		currentUserId := c.Param("id")
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
	return gin.HandlerFunc(fn)
}

func SearchContact(Session *gocql.Session) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		mobile := c.Query("mobile_no")
		currentUserId := c.Param("id")
		var success, isContact bool
		var message string = ""
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
	return gin.HandlerFunc(fn)
}

func AddContact(Session *gocql.Session) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		currentUserId := c.Param("id")
		friendMobileNo := c.Query("friendMobile")
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
	return gin.HandlerFunc(fn)
}

func HandleMessage(hub *Hub, Session *gocql.Session) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		// senderId, u_err := strconv.Atoi(c.Param("id"))
		// conversationId, c_err := strconv.Atoi(c.Param("conv_id"))
		// if u_err != nil || c_err != nil {
		// 	log.Println("conversion error: ", u_err, c_err)
		// }
		// messagePayload := MessagePayload{ ConversationId: conversationId, SenderId: senderId }
		var messagePayload MessagePayload
		handleWebsocket(c.Writer, c.Request, hub, Session, messagePayload)
	}
	return gin.HandlerFunc(fn)
}

func handleWebsocket(w http.ResponseWriter, req *http.Request, hub *Hub, Session *gocql.Session, messagePayload MessagePayload) {
	ws, err := upgrader.Upgrade(w, req, nil)
	log.Println("fun hub", hub, "hub")
	if err != nil {
		log.Fatal("Problem updating to sockets: ", err)
	}
	var msg MessagePayload
	err = ws.ReadJSON(&msg)
	if err != nil {
		log.Println("Error fetching params: ", err)
	}
	log.Println("mp", messagePayload)
	CreateNewSocket(ws, hub, Session, msg)
}

// Channels are the pipes that connect goroutines.
// var broadcast = make(chan Message)

// var clients = make(map[*websocket.Conn] bool)


// func TokenAuthMiddleware() gin.HandlerFunc {
// 	return func (c *gin.Context) {
// 		err := TokenValid(c.Request)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, err.Error())
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }