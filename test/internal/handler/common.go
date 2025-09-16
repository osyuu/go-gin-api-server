package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var (
	InvalidJSON = `{"invalid": json}`
)

// create JSON request body
func createJSONRequest(data interface{}) *bytes.Buffer {
	jsonData, _ := json.Marshal(data)
	return bytes.NewBuffer(jsonData)
}

// create HTTP request with JSON body
func createJSONHTTPRequest(method, url string, data interface{}) *http.Request {
	req, _ := http.NewRequest(method, url, createJSONRequest(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// create typed request
func createTypedJSONRequest(method, url string, data interface{}) *http.Request {
	return createJSONHTTPRequest(method, url, data)
}
