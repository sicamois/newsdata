package newsdata

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"
)

// structToMap converts a struct into a map of query parameters, handling slices.
func structToMap(params interface{}) (map[string]string, error) {

	result := make(map[string]string)
	// dereference pointer with Elem() if needed
	v := reflect.ValueOf(params).Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("params must be a struct")
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		if value.IsZero() {
			continue
		}
		if tag == "removeduplicate" {
			if value.Bool() {
				result[tag] = "1"
			}
			continue
		}
		switch value.Kind() {
		case reflect.Slice:
			// Join slices into comma-separated strings
			sliceValues := make([]string, value.Len())
			for j := 0; j < value.Len(); j++ {
				sliceValues[j] = fmt.Sprintf("%v", value.Index(j).Interface())
			}
			result[tag] = strings.Join(sliceValues, ",")
		default:
			result[tag] = fmt.Sprintf("%v", value.Interface())
		}
	}
	return result, nil
}

//
// LOGGER
//

// A LevelHandler wraps a Handler with an Enabled method
// that returns false for levels below a minimum.
type LevelHandler struct {
	level   slog.Leveler
	handler slog.Handler
	writer  io.Writer
}

// NewLevelHandler returns a LevelHandler with the given level.
// All methods except Enabled delegate to h.
func newLevelHandler(level slog.Leveler, h slog.Handler, w io.Writer) *LevelHandler {
	// Optimization: avoid chains of LevelHandlers.
	if lh, ok := h.(*LevelHandler); ok {
		h = lh.Handler()
	}
	return &LevelHandler{level, h, w}
}

// Enabled implements Handler.Enabled by reporting whether
// level is at least as large as h's level.
func (h *LevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle implements Handler.Handle.
func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newLevelHandler(h.level, h.handler.WithAttrs(attrs), h.writer)
}

// WithGroup implements Handler.WithGroup.
func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return newLevelHandler(h.level, h.handler.WithGroup(name), h.writer)
}

// Handler returns the Handler wrapped by h.
func (h *LevelHandler) Handler() slog.Handler {
	return h.handler
}

// Create a new logger that writes on the chosen io.writer with the given level.
func NewCustomLogger(w io.Writer, level slog.Level) *slog.Logger {
	th := slog.NewTextHandler(w, nil)
	logger := slog.New(newLevelHandler(level, th, w))
	return logger
}
