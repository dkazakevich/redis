package main

import (
	"os"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
	"fmt"
)

var app App

var stringKey = "sixthMonth"
var stringValue = "June"
var tempStringKey = "tempString"
var tempStringValue = "temp string value"
var dictKey = "planets"
var dictValue = map[string]interface{}{"planet1": "Mercury", "planet2": "Venus", "planet3": "Earth"}
var listKey = "cars"
var listValue = []interface{}{"Toyota", "Opel", "Ford"}

func TestMain(m *testing.M) {
	app = App{}
	app.Initialize()
	code := m.Run()
	os.Exit(code)
}

func TestPutAndGetString(t *testing.T) {
	buff, _ := json.Marshal(stringValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v?expire=20", baseUrl, stringKey), buff,
		http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v", baseUrl, stringKey), nil,
		http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, stringValue, result[valueParam])
}

func TestPutAndGetDict(t *testing.T) {
	buff, _ := json.Marshal(dictValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v", baseUrl, dictKey), buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v?dictKey=planet1", baseUrl, dictKey),
		nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, "Mercury", result[valueParam])
}

func TestPutAndGetList(t *testing.T) {
	buff, _ := json.Marshal(listValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v", baseUrl, listKey), buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v?listIndex=1", baseUrl, listKey),
		nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, listValue[1], result[valueParam])
}

func TestGetNonExistentItem(t *testing.T) {
	executeRequest(t, http.MethodGet, baseUrl + "values/nonExistent", nil, http.StatusNotFound)
}

func TestKeys(t *testing.T) {
	executeRequest(t, http.MethodGet, baseUrl + "keys", nil, http.StatusOK)
}

func TestDelete(t *testing.T) {
	executeRequest(t, http.MethodDelete, fmt.Sprintf("%vvalues/%v", baseUrl, listKey), nil, http.StatusOK)
	executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v", baseUrl, listKey), nil, http.StatusNotFound)
}

func TestExpireAndCheckTtl(t *testing.T) {
	buff, _ := json.Marshal(tempStringValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v", baseUrl, tempStringKey), buff, http.StatusOK)

	buff, _ = json.Marshal(10)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vexpire/%v", baseUrl, tempStringKey), buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vttl/%v", baseUrl, tempStringKey), nil, http.StatusOK)
	var result map[string]int
	json.Unmarshal(response.Body.Bytes(), &result)
	if result[valueParam] < 0 {
		t.Fatalf("Expected: '%v' not positive.", result[valueParam])
	}
}

func TestNonExistentExpire(t *testing.T) {
	buff, _ := json.Marshal(10)
	executeRequest(t, http.MethodPut, baseUrl + "expire/nonExistent", buff, http.StatusNotFound)
}

func TestPersistItemTtl(t *testing.T) {
	buff, _ := json.Marshal(10)
	executeRequest(t, http.MethodGet, baseUrl + "ttl/planets", buff, http.StatusNotFound)
}

func TestNonExistentTtl(t *testing.T) {
	buff, _ := json.Marshal(10)
	executeRequest(t, http.MethodGet, baseUrl + "ttl/nonExistent", buff, http.StatusNotFound)
}

func executeRequest(t *testing.T, method string, url string, buffer []byte, expectedCode int) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(buffer))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)
	assertEquals(t, expectedCode, resp.Code)
	return resp
}

func assertEquals(t *testing.T, expected interface{}, actual interface{}) {
	if expected != actual {
		t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
	}
}