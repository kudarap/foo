package server

import (
	"net/http"
	"time"
)

// Version represents service version details.
type Version struct {
	Tag    string    `json:"tag"`
	Commit string    `json:"commit"`
	Built  time.Time `json:"built"`
}

func GetVersion(v Version) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodeJSONResp(w, v, http.StatusOK)
	}
}
