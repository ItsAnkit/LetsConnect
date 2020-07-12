package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"app/controllers"
	"app/db"
)

func main() {
	log.Println("Server Started .......")

	router := gin.Default()
	router.Use(cors.Default())
	hub := controllers.HubInit()
	Session := db.StartCassandraSession()
	go hub.Run()

	router.POST("/login", controllers.Login(Session))
	router.GET("/conversations/:id/chats", controllers.GetChats(Session))
	router.GET("/conversations/:id/search", controllers.SearchContact(Session))
	router.POST("/conversations/:id/add", controllers.AddContact(Session))

	router.GET("/ws", controllers.HandleMessage(hub, Session))
	router.Run(":8050")
}