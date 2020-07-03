package controllers

import (
	"log"
	"fmt"
	"github.com/gocql/gocql"
	"time"
	"strconv"
)


type User struct {
	Id int `json:"id"`
	MobileNo string `json:"mobile_no" binding:"required"`
	Username string `json:"username" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
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

func totalUsers(Session *gocql.Session) int {
	var maxId int
	err := Session.Query("select max(id) from users").Scan(&maxId)
	if err != nil {
		log.Println("Aggregation Error: ", err)
	}
	return maxId
}