package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"app/controllers"
)

func main() {
	log.Println("Server Started .......")

	router := gin.Default()
	router.Use(cors.Default())

	router.POST("/login", controllers.Login)
	router.GET("/conversations/:id/chats", controllers.GetChats)
	router.GET("/conversations/:id/search", controllers.SearchContact)
	router.POST("/conversations/:id/add", controllers.AddContact)

	router.GET("/ws", controllers.HandleMessage)
	router.Run(":8050")
}