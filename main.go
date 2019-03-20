package main

import (
	"github.com/joisandresky/new-holotor/controllers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/ws", wsHandler)
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

	go refreshAnalytics(conn)
}

func refreshAnalytics(conn *websocket.Conn) {
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("an error occured for getting message", err)
			break
		}
		conn.WriteMessage(t, msg)
	}
	defer conn.Close()
}

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
