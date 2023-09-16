package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func GetArchersByID(s service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		c, err := s.ArcherByID(r.Context(), v["id"])
		if err != nil {
			encodeJSONError(w, err, http.StatusBadRequest)
			return
		}

		encodeJSONResp(w, c, http.StatusOK)
	}
}

func ListArchers(s service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodeJSONResp(w, struct {
			Msg string `json:"message"`
		}{"no fighters yet implemented"}, http.StatusOK)
	}
}
