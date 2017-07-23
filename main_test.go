package main

import (
	"os"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
)

var a App

func TestMain(m *testing.M) {

	a = App{}
	a.Initialize()
	code := m.Run()
	os.Exit(code)
}

func TestPutAndGetString(t *testing.T) {

	value := "June"
	buff, _ := json.Marshal(value)
	executeRequest(t, "PUT", "/api/v1/values/sixthMonth?expire=20", buff, http.StatusOK)

	response := executeRequest(t, "GET", "/api/v1/values/sixthMonth", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, value, result["value"])
}

func TestPutAndGetDict(t *testing.T) {

	value := map[string]string{"planet1": "Mercury", "planet2": "Venus", "planet3": "Earth"}
	buff, _ := json.Marshal(value)
	executeRequest(t, "PUT", "/api/v1/values/planets", buff, http.StatusOK)

	response := executeRequest(t, "GET", "/api/v1/values/planets?dictKey=planet1", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, "Mercury", result["value"])
}

func TestPutAndGetList(t *testing.T) {

	value := [3]string{"Toyota", "Opel", "Ford"}
	buff, _ := json.Marshal(value)
	executeRequest(t, "PUT", "/api/v1/values/cars", buff, http.StatusOK)

	response := executeRequest(t, "GET", "/api/v1/values/cars?listIndex=1", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, "Opel", result["value"])
}

func TestGetNonExistentItem(t *testing.T) {

	executeRequest(t, "GET", "/api/v1/values/nonExistent", nil, http.StatusNotFound)
}

func TestKeys(t *testing.T) {

	executeRequest(t, "GET", "/api/v1/keys", nil, http.StatusOK)
}

func TestDelete(t *testing.T) {

	executeRequest(t, "DELETE", "/api/v1/values/cars", nil, http.StatusOK)
	executeRequest(t, "GET", "/api/v1/values/cars", nil, http.StatusNotFound)
}

func TestExpireAndCheckTtl(t *testing.T) {

	buff, _ := json.Marshal("temp string")
	executeRequest(t, "PUT", "/api/v1/values/tempString", buff, http.StatusOK)

	buff, _ = json.Marshal(10)
	executeRequest(t, "PUT", "/api/v1/expire/tempString", buff, http.StatusOK)

	response := executeRequest(t, "GET", "/api/v1/ttl/tempString", nil, http.StatusOK)
	var result map[string]int
	json.Unmarshal(response.Body.Bytes(), &result)
	if result["value"] < 0 {
		t.Fatalf("Expected: '%v' not positive.", result["value"])
	}
}

func TestNonExistentExpire(t *testing.T) {

	buff, _ := json.Marshal(10)
	executeRequest(t, "PUT", "/api/v1/expire/nonExistent", buff, http.StatusNotFound)
}

func TestPersistItemTtl(t *testing.T) {

	buff, _ := json.Marshal(10)
	executeRequest(t, "GET", "/api/v1/ttl/planets", buff, http.StatusNotFound)
}

func TestNonExistentTtl(t *testing.T) {

	buff, _ := json.Marshal(10)
	executeRequest(t, "GET", "/api/v1/ttl/nonExistent", buff, http.StatusNotFound)
}

func executeRequest(t *testing.T, method string, url string, buffer []byte, expectedCode int) *httptest.ResponseRecorder {

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(buffer))
	resp := httptest.NewRecorder()
	a.Router.ServeHTTP(resp, req)
	assertEqual(t, expectedCode, resp.Code)

	return resp
}

func assertEqual(t *testing.T, expected interface{}, actual interface{}) {

	if expected == actual {
		return
	}

	t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
}