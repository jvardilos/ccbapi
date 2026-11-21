package ccbapi

import (
	"io"
	"log/slog"
)

var logger = slog.New(slog.NewTextHandler(io.Discard, nil))

// SetLogger overrides the package's logger, e.g. to enable output at a
// chosen level or route logs through the consuming application's handler.
func SetLogger(l *slog.Logger) {
	logger = l
}
