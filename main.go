package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joisandresky/new-holotor/controllers"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var wsupgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	MsgType	string
	Message string
	Sender float64
}

func main() {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.GET("/ws", wsHandler)
	go handleMessage()
	r.POST("/headless", headlessHandler)
	driver := r.Group("/api/drivers")
	{
		driver.GET("/analytics/:id", controllers.GetAnalytics)
		driver.POST("/analytics", controllers.CreateNewDriver)
		driver.PUT("/analytics/:id", controllers.UpdateDistance)
	}

	tracking := r.Group("/api/trackings", gin.BasicAuth(gin.Accounts{
		"holotor-go": "golangkencengcoyy",
	}))
	{
		tracking.GET("/summary", controllers.GetAllTrackingDrivers)
		tracking.GET("/ads/:id/driver-location", controllers.GetDriverLastLocationByAd)
		tracking.GET("/ads/:id", controllers.GetTrackingByAd)
		tracking.GET("/driver/:id/location", controllers.GetDriverLastLocation)
		tracking.GET("/driver/:id", controllers.GetTrackingByDriver)
		tracking.POST("/", controllers.CreateTracking)
	}

	r.Run(":8989")
}

func wsHandler(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.Error(c.Writer, "Could not open websocket connection", http.StatusBadRequest)
	}

	defer conn.Close()
	clients[conn] = true

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error: %v", err)
			delete(clients, conn)
		}
		broadcast <- msg
	}
	//for {
	//	t, msg, err := conn.ReadMessage()
	//	if err != nil {
	//		log.Println("an error occured for getting message", err)
	//		break
	//	}
	//	conn.WriteMessage(t, msg)
	//}
	//go refreshAnalytics(conn)
}

func handleMessage() {
	for {
		msg := <- broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

//func refreshAnalytics(conn *websocket.Conn) {
//	for {
//		t, msg, err := conn.ReadMessage()
//		if err != nil {
//			log.Println("an error occured for getting message", err)
//			break
//		}
//		conn.WriteMessage(t, msg)
//	}
//}

func headlessHandler (c *gin.Context) {
	var data interface{}
	c.Bind(&data)

	log.Println(data)
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
