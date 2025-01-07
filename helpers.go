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

// UnmarshalJSON implements the json.Unmarshaler interface for DateTime.
// It parses the date string using the time.DateTime format and handles null values.
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

// UnmarshalJSON implements the json.Unmarshaler interface for Tags.
// It handles special cases where the API returns restriction messages or null values,
// and splits comma-separated tag strings into slices.
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

// UnmarshalJSON implements the json.Unmarshaler interface for SentimentStats.
// It handles cases where the API returns error messages for plan restrictions,
// null values, and parses the sentiment statistics into their respective fields.
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

// levelHandler wraps a slog.Handler with level-based filtering capabilities.
// It implements all Handler methods and adds level-based message filtering.
type levelHandler struct {
	level   slog.Leveler
	handler slog.Handler
	writer  io.Writer
}

// newlevelHandler creates a new levelHandler with the specified minimum log level.
// It wraps the provided handler and optimizes chains of levelHandlers by unwrapping them.
func newlevelHandler(level slog.Leveler, h slog.Handler, w io.Writer) *levelHandler {
	// Optimization: avoid chains of levelHandlers.
	if lh, ok := h.(*levelHandler); ok {
		h = lh.Handler()
	}
	return &levelHandler{level, h, w}
}

// Enabled implements slog.Handler.Enabled by checking if the log level
// meets the minimum level requirement.
func (h *levelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements slog.Handler.Handle by delegating to the wrapped handler.
func (h *levelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.WithAttrs by creating a new levelHandler
// with the additional attributes.
func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newlevelHandler(h.level, h.handler.WithAttrs(attrs), h.writer)
}

// WithGroup implements slog.Handler.WithGroup by creating a new levelHandler
// with the additional group name.
func (h *levelHandler) WithGroup(name string) slog.Handler {
	return newlevelHandler(h.level, h.handler.WithGroup(name), h.writer)
}

// Handler returns the underlying slog.Handler wrapped by this levelHandler.
func (h *levelHandler) Handler() slog.Handler {
	return h.handler
}

// newCustomLogger creates a new slog.Logger with level-based filtering.
// It configures the logger with the specified writer and minimum log level.
func newCustomLogger(w io.Writer, level slog.Level) *slog.Logger {
	th := slog.NewTextHandler(w, nil)
	logger := slog.New(newlevelHandler(level, th, w))
	return logger
}
