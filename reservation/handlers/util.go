package handlers

import (
	"encoding/json"
	"net/http"
)

func WriteResp(resp any, statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	if resp == nil {
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(respBytes)
}
