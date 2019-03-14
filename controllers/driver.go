package controllers

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joisandresky/new-holotor/config"
	"github.com/joisandresky/new-holotor/models"
	"gopkg.in/mgo.v2/bson"
)

// GetAnalytics - get all analytics
func GetAnalytics(c *gin.Context) {
	session, err := config.Connect()
	driverID := c.Param("id")
	var driver models.Driver
	if driverID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Param ID must be set!",
			"success": false,
		})
		return
	}

	if err = session.DB("holotor").C("driver").Find(bson.M{"driver_id": driverID}).One(&driver); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "Driver Not found or something error!",
			"success": false,
			"error":   err,
		})
		return
	}
	defer session.Close()

	currentTime := time.Now()
	if dateEqual(currentTime, driver.UpdatedAt) != true {
		driver.Daily = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"driver":  driver,
	})
}

// UpdateDistance - Update Driver Distance
func UpdateDistance(c *gin.Context) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	session, err := config.Connect()
	driverID := c.Param("id")
	distance := c.PostForm("distance")
	var driver models.Driver
	defer session.Close()

	var newDistance float64
	if s, errParse := strconv.ParseFloat(distance, 64); errParse == nil {
		newDistance = s
	}

	if err = session.DB("holotor").C("driver").Find(bson.M{"driver_id": driverID}).One(&driver); err != nil {
		driver = models.Driver{ID: bson.NewObjectId(), DriverID: driverID, Daily: newDistance, Annualy: newDistance, Total: newDistance, UpdatedAt: time.Now()}
		errCreate := session.DB("holotor").C("driver").Insert(driver)

		if errCreate != nil {
			log.Println(errCreate)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": "Driver Not found or something error!",
				"success": false,
				"error":   errCreate,
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"message": nil,
			"success": true,
			"error":   err,
			"driver":  driver,
		})
		return
	}

	if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"total": driver.Total + newDistance, "updated_at": time.Now()}}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Error Updating Driver Distance",
			"success": false,
			"error":   err,
		})
		return
	}

	go UpdateDailyDistance(&driver, driverID, newDistance)
	go UpdateAnnualyDistance(&driver, driverID, newDistance)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Distance Driver Updated!",
		"driver_id": driverID,
		"success":   true,
	})
}

// CreateNewDriver - Create new driver
func CreateNewDriver(c *gin.Context) {
	session, err := config.Connect()

	defer session.Close()
	body := &models.Driver{ID: bson.NewObjectId()}
	c.BindJSON(body)

	if body == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Pengisian Data Tidak Lengkap",
			"success": false,
			"body":    body,
		})
	}

	if err = session.DB("holotor").C("driver").Insert(body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Error on Saving New Driver",
			"success": false,
			"error":   err,
		})
	}
	defer session.Close()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Success Saving New Driver",
		"success": true,
		"driver":  body,
	})
}

// UpdateDailyDistance - update distance daily
func UpdateDailyDistance(driver *models.Driver, driverID string, distance float64) {
	currentTime := time.Now()
	session, err := config.Connect()
	defer session.Close()
	log.Print(dateEqual(currentTime, driver.UpdatedAt))
	if dateEqual(currentTime, driver.UpdatedAt) != true {
		if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"daily": distance, "updated_at": time.Now()}}); err != nil {
			log.Println("Error Updating Driver Distance")
			return
		}

		log.Printf("Daily Distance Updated!")
	} else {
		if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"daily": driver.Daily + distance, "updated_at": time.Now()}}); err != nil {
			log.Println("Error Updating Driver Distance")
			return
		}

		log.Printf("Daily Distance Updated!")
	}
}

// UpdateAnnualyDistance - update distance Annualy
func UpdateAnnualyDistance(driver *models.Driver, driverID string, distance float64) {
	currentTime := time.Now()
	session, err := config.Connect()
	defer session.Close()

	if currentTime.Year() != driver.UpdatedAt.Year() {
		if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"annualy": distance, "updated_at": time.Now()}}); err != nil {
			log.Println("Error Updating Driver Distance")
			return
		}

		log.Printf("Annual Distance Updated!")
	} else {
		if err = session.DB("holotor").C("driver").Update(bson.M{"driver_id": driverID}, bson.M{"$set": bson.M{"annualy": driver.Annualy + distance, "updated_at": time.Now()}}); err != nil {
			log.Println("Error Updating Driver Distance")
			return
		}

		log.Printf("Annual Distance Updated!")
	}
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
