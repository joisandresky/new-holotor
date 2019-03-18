package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/joisandresky/new-holotor/models"
	"github.com/joisandresky/new-holotor/config"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

//CreateTracking - creating tracking point
func CreateTracking(c *gin.Context) {
	session, err := config.Connect()

	defer session.Close()
	body := &models.Tracking{ID: bson.NewObjectId()}
	c.BindJSON(body)

	if body == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Pengisian Data Tidak Lengkap",
			"success": false,
			"body":    body,
		})
		return
	}

	if err = session.DB("holotor").C("tracking").Insert(body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Error on Saving New Tracking Point",
			"success": false,
			"error":   err,
		})
		return
	}
	defer session.Close()

	driverIdStr := strconv.Itoa(body.DriverID)

	go UpdateTrackDistance(driverIdStr, body.Distance)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Success Saving New Tracking Point",
		"success": true,
		"tracking":  body,
	})
}

//UpdateTrackingDistance - updating tracking distance
func UpdateTrackDistance(driverID string, distance float64) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	session, err := config.Connect()
	defer session.Close()

	var driver models.Driver

	if err = session.DB("holotor").C("driver").Find(bson.M{"driver_id": driverID}).One(&driver); err != nil {
		log.Println("Driver Not Found")
		return
	}

	if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"total": driver.Total + distance, "updated_at": time.Now()}}); err != nil {
		log.Println("Error Updating Driver Distance")
		return
	}

	log.Printf("Total Distance Updated!")

	go UpdateDailyDistance(&driver, driverID, distance)
	go UpdateAnnualyDistance(&driver, driverID, distance)
}
