package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEmptyTable(t *testing.T) {
	dbDial, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	clearDB(*dbDial)

	appHandler := AppHandler{dbDial}
	req, _ := http.NewRequest("GET", "/key1", nil)

	e := echo.New()
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)

	if assert.NoError(t, appHandler.GetAnswer(c)) {
		assert.Equal(t, http.StatusNotFound, rr.Code)
	}
}

func clearDB(db gorm.DB) {
	_ = db.Exec("DELETE  FROM answers")
	_ = db.Exec("DELETE  FROM events")
}

//create → delete → create → update
func TestCreateDeleteCreateUpdateEventSuccess(t *testing.T) {

	dbDial, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	dbDial.AutoMigrate(&Answer{}, &Event{})
	clearDB(*dbDial)
	defer func() {
		clearDB(*dbDial)
	}()
	appHandler := AppHandler{dbDial}

	testRequest := []EventRequest{
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Delete,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Update,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value2",
			},
		},
	}
	testExpectedResponse := []int{
		http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK,
	}
	for i, body := range testRequest {
		payload := new(bytes.Buffer)
		err = json.NewEncoder(payload).Encode(body)

		if err != nil {
			panic("failed to connect to database")
		}
		req, _ := http.NewRequest("POST", "/", payload)
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		e := echo.New()
		rr := httptest.NewRecorder()
		c := e.NewContext(req, rr)
		if assert.NoError(t, appHandler.PostEvent(c)) {
			fmt.Println(rr.Result().Body)
			assert.Equal(t, testExpectedResponse[i], rr.Code)
		}
	}
}

//create → update → delete → create → update
func TestCreateUpdateDeleteCreateUpdateEventSuccess(t *testing.T) {

	dbDial, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	dbDial.AutoMigrate(&Answer{}, &Event{})
	clearDB(*dbDial)
	defer func() {
		clearDB(*dbDial)
	}()
	appHandler := AppHandler{dbDial}

	testRequest := []EventRequest{
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Update,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value2",
			},
		},
		{
			Type: Delete,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Update,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value2",
			},
		},
	}
	testExpectedResponse := []int{
		http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK,
	}
	for i, body := range testRequest {
		payload := new(bytes.Buffer)
		err = json.NewEncoder(payload).Encode(body)

		if err != nil {
			panic("failed to connect to database")
		}
		req, _ := http.NewRequest("POST", "/", payload)
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		e := echo.New()
		rr := httptest.NewRecorder()
		c := e.NewContext(req, rr)
		if assert.NoError(t, appHandler.PostEvent(c)) {
			fmt.Println(rr.Result().Body)
			assert.Equal(t, testExpectedResponse[i], rr.Code)
		}
	}
}

//create → delete → update
func TestCreateDeleteUpdateEventFail(t *testing.T) {

	dbDial, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	dbDial.AutoMigrate(&Answer{}, &Event{})
	clearDB(*dbDial)
	defer func() {
		clearDB(*dbDial)
	}()
	appHandler := AppHandler{dbDial}

	testRequest := []EventRequest{
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Delete,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Update,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value2",
			},
		},
	}
	testExpectedResponse := []int{
		http.StatusOK, http.StatusOK, http.StatusBadRequest,
	}
	for i, body := range testRequest {
		payload := new(bytes.Buffer)
		err = json.NewEncoder(payload).Encode(body)

		if err != nil {
			panic("failed to connect to database")
		}
		req, _ := http.NewRequest("POST", "/", payload)
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		e := echo.New()
		rr := httptest.NewRecorder()
		c := e.NewContext(req, rr)
		if assert.NoError(t, appHandler.PostEvent(c)) {
			fmt.Println(rr.Result().Body)
			assert.Equal(t, testExpectedResponse[i], rr.Code)
		}
	}
}

//create → create
func TestCreateCreateEventFail(t *testing.T) {

	dbDial, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	dbDial.AutoMigrate(&Answer{}, &Event{})
	clearDB(*dbDial)
	defer func() {
		clearDB(*dbDial)
	}()
	appHandler := AppHandler{dbDial}

	testRequest := []EventRequest{
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
		{
			Type: Create,
			Data: map[string]interface{}{
				"key":   "key1",
				"value": "value1",
			},
		},
	}
	testExpectedResponse := []int{
		http.StatusOK, http.StatusBadRequest,
	}
	for i, body := range testRequest {
		payload := new(bytes.Buffer)
		err = json.NewEncoder(payload).Encode(body)

		if err != nil {
			panic("failed to connect to database")
		}
		req, _ := http.NewRequest("POST", "/", payload)
		req.Header.Add("Content-Type", "application/json;charset=UTF-8")

		e := echo.New()
		rr := httptest.NewRecorder()
		c := e.NewContext(req, rr)
		if assert.NoError(t, appHandler.PostEvent(c)) {
			fmt.Println(rr.Result().Body)
			assert.Equal(t, testExpectedResponse[i], rr.Code)
		}
	}
}
