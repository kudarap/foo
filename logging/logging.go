package logging

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	h := slog.NewJSONHandler(os.Stderr, nil)
	return slog.New(h)
}
