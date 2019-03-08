package main

import (
	"net/http"

	"github.com/joisandresky/new-holotor/controllers"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	r := gin.Default()
	r.GET("/ws", wsHandler)

	driver := r.Group("/api/drivers")
	{
		driver.GET("/analytics/:id", controllers.GetAnalytics)
		driver.PUT("/analytics/:id", controllers.UpdateDistance)
	}

	r.Run(":8787")
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
			break
		}
		conn.WriteMessage(t, msg)
	}
}
