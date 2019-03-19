package main

import (
	"log"
	"net/http"

	"github.com/joisandresky/new-holotor/controllers"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{}

func main() {
	r := gin.Default()
	r.GET("/ws", wsHandler)
	r.POST("/headless", headlessHandler)
	driver := r.Group("/api/drivers")
	{
		driver.GET("/analytics/:id", controllers.GetAnalytics)
		driver.POST("/analytics", controllers.CreateNewDriver)
		driver.PUT("/analytics/:id", controllers.UpdateDistance)
	}

	tracking := r.Group("/api/trackings")
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

	defer conn.Close()
	go refreshAnalytics(conn)
}

func refreshAnalytics(conn *websocket.Conn) {
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(t, msg)
	}
}

func headlessHandler (c *gin.Context) {
	var data interface{}
	c.Bind(&data)

	log.Println(data)
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}
