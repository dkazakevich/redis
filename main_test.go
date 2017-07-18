package main_test

import (
	"os"
	"testing"

	"github.com/dkazakevich/redis"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize()
	code := m.Run()
	os.Exit(code)
}

func TestGetNonExistentItem(t *testing.T) {

	req, _ := http.NewRequest("GET", "/cache/string", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestPutItem(t *testing.T) {

	payload := []byte(`"value"`)

	req, _ := http.NewRequest("PUT", "/cache/string", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var value string
	json.Unmarshal(response.Body.Bytes(), &value)

	if value != "value" {
		t.Errorf("Expected: 'value'. Actual: '%v'", value)
	}
}

func TestGetExistentItem(t *testing.T) {

	req, _ := http.NewRequest("GET", "/cache/string", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var value string
	json.Unmarshal(response.Body.Bytes(), &value)

	if value != "value" {
		t.Errorf("Expected: 'value'. Actual: '%v'", value)
	}
}

func TestDeleteItem(t *testing.T) {

	req, _ := http.NewRequest("DELETE", "/cache/string", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/cache/string", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	a.Router.ServeHTTP(resp, req)

	return resp
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Actual %d\n", expected, actual)
	}
}