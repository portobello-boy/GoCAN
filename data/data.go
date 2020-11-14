package data

import (
	"encoding/json"
	"net/http"
)

// ParseData handles transforming http.Request into DataRequest with error handling
func ParseData(w http.ResponseWriter, r *http.Request) DataRequest {
	var dataReq DataRequest
	err := json.NewDecoder(r.Body).Decode(&dataReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return dataReq
}

// ParseJoin handles transforming http.Request into JoinRequest with error handling
func ParseJoin(w http.ResponseWriter, r *http.Request) JoinRequest {
	var joinReq JoinRequest
	err := json.NewDecoder(r.Body).Decode(&joinReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if joinReq.Host == "" {
		joinReq.Host = r.RemoteAddr
	}

	return joinReq
}
