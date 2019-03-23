package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/now"
	"github.com/joisandresky/new-holotor/config"
	"github.com/joisandresky/new-holotor/models"
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
	body.CreatedAt = time.Now()
	if err = session.DB("holotor").C("tracking").Insert(body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Error on Saving New Tracking Point",
			"success": false,
			"error":   err,
		})
		return
	}

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

//GetTrackingByDriver - get summary tracking driver by id
func GetTrackingByDriver(c *gin.Context) {
	session, err := config.Connect()
	defer session.Close()
	paramID := c.Param("id")
	sortParam := c.DefaultQuery("sort", "asc")
	filter := c.DefaultQuery("filter", "all")
	driverID, err := strconv.Atoi(paramID)
	var sortBy string

	if sortParam == "asc" {
		sortBy = "created_at"
	} else {
		sortBy = "-created_at"
	}

	trackings := []*models.Tracking{}
	if filter == "all" {
		if err = session.DB("holotor").C("tracking").Find(bson.M{"driverId": driverID}).Sort(sortBy).All(&trackings); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Driver not Found or Some error occured!",
				"success": false,
				"error": err,
			})
			return
		}
	} else {
		var query interface{}
		switch ft := filter ; ft {
		case "day":
			startDate := now.BeginningOfDay()
			endDate := now.EndOfDay()
			query = bson.M{
				"driverId": driverID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		case "week":
			startDate := now.BeginningOfWeek()
			endDate := now.EndOfWeek()
			query = bson.M{
				"driverId": driverID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		case "month":
			startDate := now.BeginningOfMonth()
			endDate := now.EndOfMonth()
			query = bson.M{
				"driverId": driverID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		default:
			startDate := now.BeginningOfYear()
			endDate := now.EndOfYear()
			query = bson.M{
				"driverId": driverID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		}
		if err = session.DB("holotor").C("tracking").Find(query).Sort("created_at").All(&trackings); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Driver not Found or Some error occured!",
				"success": false,
				"error": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"trackings":  trackings,
	})
}

//GetAllTrackingDrivers - get all drivers tracking point history
func GetAllTrackingDrivers(c *gin.Context) {
	session, err := config.Connect()
	defer session.Close()
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
	adsId, err := strconv.Atoi(c.DefaultQuery("adsId", "-1"))
	skip := (page - 1 ) * limit

	var trackings []interface{}
	var query interface{}
	if adsId != -1 {
		query = []bson.M{
			bson.M{
				"$match": bson.M{ "adsId": adsId },
			},
			bson.M{
				"$sort": bson.M{
					"created_at": 1,
				},
			},
			bson.M{
				"$group": bson.M{
					"_id": bson.M{"driverId": "$driverId" },
					"driverId": bson.M{ "$first": "$driverId" },
					"locations": bson.M{"$push": "$$ROOT"},
					//"lat": bson.M{"$first": "$lat"},
					//"long": bson.M{"$first": "$long"},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 0,
					"adsId": "$_id.adsId",
					"driverId": 1,
					"lat": 1,
					"long": 1,
					"locations": 1,
				},
			},
			bson.M{
				"$skip": skip,
			},
			bson.M{
				"$limit": limit,
			},
		}
	} else {
		query = []bson.M{
			bson.M{
				"$sort": bson.M{
					"created_at": 1,
				},
			},
			bson.M{
				"$group": bson.M{
					"_id": bson.M{"driverId": "$driverId" },
					"driverId": bson.M{ "$first": "$driverId" },
					"locations": bson.M{"$push": "$$ROOT"},
					//"lat": bson.M{"$first": "$lat"},
					//"long": bson.M{"$first": "$long"},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 0,
					"adsId": "$_id.adsId",
					"driverId": 1,
					"lat": 1,
					"long": 1,
					"locations": 1,
				},
			},
			bson.M{
				"$skip": skip,
			},
			bson.M{
				"$limit": limit,
			},
		}
	}
	pipe := session.DB("holotor").C("tracking").Pipe(query)

	if err = pipe.All(&trackings); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Driver not Found or Some error occured!",
			"success": false,
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"trackings":  trackings,
	})
}

//GetTrackingByAd - get tracking data by Ads Id
func GetTrackingByAd(c *gin.Context) {
	session, err := config.Connect()
	defer session.Close()
	paramID := c.Param("id")
	filter := c.DefaultQuery("filter", "all")
	adsId, err := strconv.Atoi(paramID)

	trackings := []*models.Tracking{}

	if filter == "all" {
		if err = session.DB("holotor").C("tracking").Find(bson.M{"adsId": adsId}).Sort("created_at").All(&trackings); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Ads not Found or Some error occured!",
				"success": false,
				"error": err,
			})
			return
		}
	} else {
		var query interface{}
		switch ft := filter ; ft {
		case "day":
			startDate := now.BeginningOfDay()
			endDate := now.EndOfDay()
			query = bson.M{
				"adsId": adsId,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		case "week":
			startDate := now.BeginningOfWeek()
			endDate := now.EndOfWeek()
			query = bson.M{
				"adsId": adsId,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		case "month":
			startDate := now.BeginningOfMonth()
			endDate := now.EndOfMonth()
			query = bson.M{
				"adsId": adsId,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		default:
			startDate := now.BeginningOfYear()
			endDate := now.EndOfYear()
			query = bson.M{
				"adsId": adsId,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			}
		}
		if err = session.DB("holotor").C("tracking").Find(query).Sort("created_at").All(&trackings); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Ads not Found or Some error occured!",
				"success": false,
				"error": err,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"trackings":  trackings,
	})
}

//GetDriverLastLocation - get driver last Location
func GetDriverLastLocation(c *gin.Context) {
	session, err := config.Connect()
	defer session.Close()
	paramID := c.Param("id")
	driverID, err := strconv.Atoi(paramID)

	var tracking *models.Tracking
	if err = session.DB("holotor").C("tracking").Find(bson.M{"driverId": driverID}).Sort("-created_at").Limit(1).One(&tracking); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Driver not Found or Some error occured!",
			"success": false,
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"location":  tracking,
	})
}

//GetDriverLastLocationByAd - get all Driver last location by Ads Id
func GetDriverLastLocationByAd(c *gin.Context) {
	session, err := config.Connect()
	defer session.Close()
	paramID := c.Param("id")
	adsId, err := strconv.Atoi(paramID)

	var trackings []interface{}
	query := []bson.M{
		bson.M{
			"$match": bson.M{ "adsId": adsId },
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{"driverId": "$driverId", "adsId": "$adsId", },
				"created_at": bson.M{"$first": "$created_at"},
				"driverId": bson.M{"$first": "$driverId"},
				"lat": bson.M{"$first": "$lat"},
				"long": bson.M{"$first": "$long"},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id": 0,
				"adsId": "$_id.adsId",
				"driverId": 1,
				"lat": 1,
				"long": 1,
			},
		},
	}
	pipe := session.DB("holotor").C("tracking").Pipe(query)

	if err = pipe.All(&trackings); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Driver not Found or Some error occured!",
			"success": false,
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": nil,
		"success": true,
		"trackings":  trackings,
	})
}
