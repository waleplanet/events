package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createEchoInstance() *echo.Echo {
	e := echo.New()
	// We don't want trailing slashes
	e.Pre(middleware.RemoveTrailingSlash())

	// Logging + recovery middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	return e
}

type AppHandler struct {
	*gorm.DB
}

func (a *AppHandler) GetAnswer(c echo.Context) error {
	var answer Answer
	key := c.Param("key")
	query := a.DB.Where("active =?", true).Where("key= ?", key)

	err := query.Last(&answer).Error
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"exception": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"anwer":   answer,
	})
}
func (a *AppHandler) PostEvent(c echo.Context) error {
	var requestBody EventRequest
	err := c.Bind(&requestBody)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"exception": err.Error(),
		})

	}
	var postErrCh = make(chan error)

	//increase No of goroutines cand use buffered channel to handle more request
	go func(dbConnection *gorm.DB, requestBody EventRequest, postErrCh chan error) {
		ProcessEvents(dbConnection, requestBody, postErrCh)

	}(a.DB, requestBody, postErrCh)

	err = <-postErrCh
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"exception": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
	})
}

func (a *AppHandler) GetHistory(c echo.Context) error {
	var events []Event
	key := c.Param("key")
	err := a.DB.Where("key =?", key).Find(&events).Error
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"exception": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success",
		"events":  events,
	})

}

//Support Multiple Users
// declare a user type
// link to model as a foreign key

type EventRequest struct {
	gorm.Model
	Type EventType `gorm:"column:type;type:varchar(255);not_null" json:"type,omitempty"`
	// Support anwer with other types using  interface
	Data map[string]interface{} `gorm:"column:data;type:text" json:"data,omitempty"`
}
type Event struct {
	gorm.Model
	Type EventType `gorm:"column:type;type:varchar(255);not_null" json:"type,omitempty"`
	Key  string    `gorm:"column:key;type:varchar(255);not_null" json:"key,omitempty"`
}

type Answer struct {
	gorm.Model
	Key    string `gorm:"column:key;type:varchar(255);not_null" json:"key,omitempty"`
	Value  string `gorm:"column:value;type:varchar(255);not null" json:"value,omitempty"`
	Active bool   `gorm:"column:active;type:tinyint(1);default:1;precision:3;scale:0;not null" json:"active"`
}

type EventType string

const (
	Create EventType = "create"
	Delete EventType = "delete"
	Update EventType = "update"
)

//sendChannel := make(chan Events)
//recieveChannel := make(chan Events)

func main() {
	dbConnection, err := gorm.Open(sqlite.Open("events.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	dbConnection.AutoMigrate(&Answer{}, &Event{})

	var e = createEchoInstance()
	//endpoints

	app := &AppHandler{
		dbConnection,
	}
	e.POST("/", app.PostEvent)

	e.GET("/:key", app.GetAnswer)
	e.GET("/history/:key", app.GetHistory)
	// for testing
	// e.GET("/", func(c echo.Context) error {
	// 	var answer []Answer
	// 	err := dbConnection.Where("active =?", true).Find(&answer).Error
	// 	if err != nil {
	// 		return c.JSON(http.StatusNotFound, map[string]interface{}{
	// 			"exception": err.Error(),
	// 		})
	// 	}
	// 	return c.JSON(http.StatusOK, map[string]interface{}{
	// 		"message": "success",
	// 		"anwer":   answer,
	// 	})
	// })

	e.Logger.Fatal(e.Start(":1323"))
}

func ProcessEvents(db *gorm.DB, eventRequest EventRequest, errChan chan error) {

	data := eventRequest.Data
	jsonbody, err := json.Marshal(data)
	if err != nil {
		errChan <- err
		return
	}

	answer := Answer{}
	if err := json.Unmarshal(jsonbody, &answer); err != nil {
		errChan <- err
		return
	}

	switch eventRequest.Type {

	case Create:
		result := db.Where(Answer{Key: answer.Key, Value: answer.Value, Active: true}).FirstOrCreate(&answer)

		if result.RowsAffected > 0 {
			errChan <- db.Create(&Event{Key: answer.Key, Type: eventRequest.Type}).Error
			return
		} else {
			errChan <- errors.New("answer with key  already exists ")
		}
	case Update:
		err := db.Where(Answer{Key: answer.Key, Active: true}).First(&answer).Error

		if err != nil {
			errChan <- err
			return
		}
		err = db.Create(&Answer{Key: answer.Key, Value: data["value"].(string)}).Error
		if err != nil {
			errChan <- err
			return
		}
		errChan <- db.Create(&Event{Key: answer.Key, Type: eventRequest.Type}).Error

	case Delete:
		err = db.Where("key =? AND value=?", answer.Key, answer.Value).Find(&answer).Error
		if err != nil {
			errChan <- err
			return
		}
		if !answer.Active {
			errChan <- errors.New("cannot delete answer")
			return
		}
		err := db.Model(&answer).Update("active", false).Error
		if err != nil {
			errChan <- err
			return
		}
		errChan <- db.Create(&Event{Key: answer.Key, Type: eventRequest.Type}).Error

	default:
		errChan <- errors.New("event type not implemented")
	}
}
