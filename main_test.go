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

func TestGetNonExistentItem(t *testing.T) {

	executeRequest(t, "GET", "/api/v1/values/string", nil, http.StatusNotFound)
}

func TestPut(t *testing.T) {

	buff, _ := json.Marshal("value")
	response := executeRequest(t, "PUT", "/api/v1/values/string", buff, http.StatusOK)

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, "value", result["value"])
}

func TestGet(t *testing.T) {

	response := executeRequest(t, "GET", "/api/v1/values/string", nil, http.StatusOK)

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, "value", result["value"])
}

func TestDelete(t *testing.T) {

	executeRequest(t, "DELETE", "/api/v1/values/string", nil, http.StatusOK)
	executeRequest(t, "GET", "/api/v1/values/string", nil, http.StatusNotFound)
}

func TestExpire(t *testing.T) {

	buff, _ := json.Marshal("value")
	executeRequest(t, "PUT", "/api/v1/values/string", buff, http.StatusOK)

	buff, _ = json.Marshal(10)
	response := executeRequest(t, "PUT", "/api/v1/expire/string", buff, http.StatusOK)

	var result map[string]bool
	json.Unmarshal(response.Body.Bytes(), &result)
	assertEqual(t, true, result["value"])
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