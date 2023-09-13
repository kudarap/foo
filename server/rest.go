package server

import (
	"encoding/json"
	"net/http"
)

const contentType = "application/json; charset=utf-8"

func encodeJSONResp(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeJSONError(w http.ResponseWriter, err error, code int) {
	m := struct {
		Error  string `json:"error"`
		Status int    `json:"status"`
	}{err.Error(), code}
	encodeJSONResp(w, m, code)
}

func decodeJSONReq(w *http.Request, in interface{}) {
	panic("implement me")
}
