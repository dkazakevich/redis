package restserver

import (
	"os"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dkazakevich/redis/internal/testutil"
)

const (
	stringKey 		= "sixthMonth"
	stringValue 	= "June"
	tempStringKey 	= "tempString"
	tempStringValue = "temp string value"
	dictKey 		= "planets"
	listKey 		= "cars"
)

var (
	rs RestServer

	dictValue = map[string]interface{}{"planet1": "Mercury", "planet2": "Venus", "planet3": "Earth"}
	listValue = []interface{}{"Toyota", "Opel", "Ford"}
)

func TestMain(m *testing.M) {
	rs.Initialize()
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
	testutil.AssertEquals(t, stringValue, result[valueParam])
}

func TestPutAndGetDict(t *testing.T) {
	buff, _ := json.Marshal(dictValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v", baseUrl, dictKey), buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v?dictKey=planet1", baseUrl, dictKey),
		nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	testutil.AssertEquals(t, "Mercury", result[valueParam])
}

func TestPutAndGetList(t *testing.T) {
	buff, _ := json.Marshal(listValue)
	executeRequest(t, http.MethodPut, fmt.Sprintf("%vvalues/%v", baseUrl, listKey), buff, http.StatusOK)

	response := executeRequest(t, http.MethodGet, fmt.Sprintf("%vvalues/%v?listIndex=1", baseUrl, listKey),
		nil, http.StatusOK)
	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)
	testutil.AssertEquals(t, listValue[1], result[valueParam])
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
	response := executeRequest(t, http.MethodGet, baseUrl + "ttl/" + dictKey, buff, http.StatusOK)
	var result map[string]int
	json.Unmarshal(response.Body.Bytes(), &result)
	testutil.AssertEquals(t, -1, result[valueParam])
}

func TestNonExistentTtl(t *testing.T) {
	buff, _ := json.Marshal(10)
	executeRequest(t, http.MethodGet, baseUrl + "ttl/nonExistent", buff, http.StatusNotFound)
}

func executeRequest(t *testing.T, method string, url string, buffer []byte, expectedCode int) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(buffer))
	resp := httptest.NewRecorder()
	rs.router.ServeHTTP(resp, req)
	testutil.AssertEquals(t, expectedCode, resp.Code)
	return resp
}