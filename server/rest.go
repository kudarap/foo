package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kudarap/foo/xerror"
)

const contentType = "application/json; charset=utf-8"

func encodeJSONResp(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func encodeJSONError(w http.ResponseWriter, err error, statusCode int) {
	m := struct {
		Error  string `json:"error"`
		Code   string `json:"code,omitempty"`
		Status int    `json:"status"`
	}{}
	m.Error = err.Error()
	m.Status = statusCode

	// Custom error encoding for xerror.
	var errX xerror.XError
	if errors.As(err, &errX) {
		m.Error = errX.Err.Error()
		m.Code = errX.Code
	}

	encodeJSONResp(w, m, statusCode)
}

func decodeJSONReq(r *http.Request, in interface{}) error {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqBody, in)
	if err != nil {
		return err
	}
	return nil
}
