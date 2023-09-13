package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	tests := []struct {
		name     string
		dc       databasePinger
		want     Health
		wantCode int
	}{
		{
			"ok",
			&mockDatabase{PingFn: func() (ok bool, err error) {
				return true, err
			}},
			Health{PostgresPing: "ok"},
			http.StatusOK,
		},
		{
			"failed",
			&mockDatabase{PingFn: func() (ok bool, err error) {
				return false, errors.New("connection failed")
			}},
			Health{PostgresPing: "connection failed"},
			http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost/healthcheck", nil)
			w := httptest.NewRecorder()
			Healthcheck(tt.dc).ServeHTTP(w, req)
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("reading body failed: %s", err)
			}

			var got Health
			if err = json.Unmarshal(body, &got); err != nil {
				t.Errorf("encoding payload failed: %s", err)
			}

			if resp.StatusCode != tt.wantCode {
				t.Errorf("Healthcheck() status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Healthcheck() body = %s, want %v", body, tt.want)
			}
		})
	}
}

type mockDatabase struct {
	PingFn func() (ok bool, err error)
}

func (d *mockDatabase) Ping() (ok bool, err error) {
	return d.PingFn()
}
