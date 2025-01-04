package newsdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

func (t *DateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	date, err := time.Parse(time.DateTime, strings.Trim(string(b), `"`))
	if err != nil {
		return fmt.Errorf("unmarshalDateTime - error unmarshalling date time - error: %w", err)
	}
	t.Time = date
	return nil
}

func (t *Tags) UnmarshalJSON(b []byte) error {
	value := string(b)
	if value == "null" || strings.HasPrefix(value, "ONLY AVAILABLE IN ") {
		*t = nil
		return nil
	}
	tags := strings.Split(string(b), ",")
	*t = make([]string, len(tags))
	for i, tag := range tags {
		(*t)[i] = tag
	}
	return nil
}

func (t *SentimentStats) UnmarshalJSON(b []byte) error {
	// If the API returns an error (typically "ONLY AVAILABLE IN PROFESSIONAL AND CORPORATE PLANS"), handle it nicely
	// Handle "null" case also
	var msg string
	if err := json.Unmarshal(b, &msg); err == nil {
		*t = SentimentStats{}
		return nil
	}

	stats := make(map[string]float64)
	if err := json.Unmarshal(b, &stats); err != nil {
		return fmt.Errorf("unmarshalSentimentStats - error unmarshalling sentiment stats - error: %w", err)
	}
	positive, ok := stats["positive"]
	if !ok {
		return fmt.Errorf("invalid sentiment stats: %v", stats)
	}
	neutral, ok := stats["neutral"]
	if !ok {
		return fmt.Errorf("invalid sentiment stats: %v", stats)
	}
	negative, ok := stats["negative"]
	if !ok {
		return fmt.Errorf("invalid sentiment stats: %v", stats)
	}

	t.Positive = positive
	t.Neutral = neutral
	t.Negative = negative
	return nil
}

// Logger Helpers

// A levelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type levelHandler struct {
	level   slog.Leveler
	handler slog.Handler
	writer  io.Writer
}

// NewlevelHandler returns a levelHandler with the given level.
// All methods except Enabled delegate to h.
func newlevelHandler(level slog.Leveler, h slog.Handler, w io.Writer) *levelHandler {
	// Optimization: avoid chains of levelHandlers.
	if lh, ok := h.(*levelHandler); ok {
		h = lh.Handler()
	}
	return &levelHandler{level, h, w}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *levelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *levelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newlevelHandler(h.level, h.handler.WithAttrs(attrs), h.writer)
}

// WithGroup implements Handler.WithGroup.
func (h *levelHandler) WithGroup(name string) slog.Handler {
	return newlevelHandler(h.level, h.handler.WithGroup(name), h.writer)
}

// Handler returns the Handler wrapped by h.
func (h *levelHandler) Handler() slog.Handler {
	return h.handler
}

// Create a new logger that writes on the chosen io.writer with the given level.
func newCustomLogger(w io.Writer, level slog.Level) *slog.Logger {
	th := slog.NewTextHandler(w, nil)
	logger := slog.New(newlevelHandler(level, th, w))
	return logger
}
