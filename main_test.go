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
	executeRequest(t, http.MethodPut, baseUrl + "values/sixthMonth?expire=20", buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, baseUrl + "values/sixthMonth", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, value, result[valueParam])
}

func TestPutAndGetDict(t *testing.T) {

	value := map[string]string{"planet1": "Mercury", "planet2": "Venus", "planet3": "Earth"}
	buff, _ := json.Marshal(value)
	executeRequest(t, http.MethodPut, baseUrl + "values/planets", buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, baseUrl + "values/planets?dictKey=planet1", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, "Mercury", result[valueParam])
}

func TestPutAndGetList(t *testing.T) {

	value := [3]string{"Toyota", "Opel", "Ford"}
	buff, _ := json.Marshal(value)
	executeRequest(t, http.MethodPut, baseUrl + "values/cars", buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, baseUrl + "values/cars?listIndex=1", nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEquals(t, "Opel", result[valueParam])
}

func TestGetNonExistentItem(t *testing.T) {

	executeRequest(t, http.MethodGet, baseUrl + "values/nonExistent", nil, http.StatusNotFound)
}

func TestKeys(t *testing.T) {

	executeRequest(t, http.MethodGet, baseUrl + "keys", nil, http.StatusOK)
}

func TestDelete(t *testing.T) {

	executeRequest(t, http.MethodDelete, baseUrl + "values/cars", nil, http.StatusOK)
	executeRequest(t, http.MethodGet, baseUrl + "values/cars", nil, http.StatusNotFound)
}

func TestExpireAndCheckTtl(t *testing.T) {

	buff, _ := json.Marshal("temp string")
	executeRequest(t, http.MethodPut, baseUrl + "values/tempString", buff, http.StatusOK)

	buff, _ = json.Marshal(10)
	executeRequest(t, http.MethodPut, baseUrl + "expire/tempString", buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, baseUrl + "ttl/tempString", nil, http.StatusOK)
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
	a.Router.ServeHTTP(resp, req)
	assertEquals(t, expectedCode, resp.Code)

	return resp
}

func assertEquals(t *testing.T, expected interface{}, actual interface{}) {

	if expected == actual {
		return
	}

	t.Fatalf("Expected: '%v'. Actual: '%v'", expected, actual)
}