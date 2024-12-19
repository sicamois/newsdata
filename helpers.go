package newsdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (t *DateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	date, err := time.Parse(time.DateTime, strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	t.Time = date
	return nil
}

func (t *DateTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(time.DateTime))), nil
}

func (t *DateTime) IsZero() bool {
	return t.Time.IsZero()
}

func (t *DateTime) After(other time.Time) bool {
	return t.Time.After(other)
}

func (t *Tags) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*t = nil
		return nil
	}
	tags := strings.Split(string(b), ",")
	*t = make([]Tag, len(tags))
	for i, tag := range tags {
		(*t)[i] = Tag{tag}
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
	err := json.Unmarshal(b, &stats)
	if err != nil {
		return err
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

// isValidField checks if a field exists in the Article struct.
func isValidField(field string) bool {
	ArticleFields := make([]string, 0)
	t := reflect.TypeOf(Article{})
	for i := 0; i < t.NumField(); i++ {
		ArticleFields = append(ArticleFields, t.Field(i).Name)
	}
	for _, allowed := range ArticleFields {
		if field == allowed {
			return true
		}
	}
	return false
}

// isValidTimeframe checks if a timeframe is well-formed.
func isValidTimeframe(timeframe string) bool {
	hours, err := strconv.Atoi(timeframe)
	if err != nil {
		minValue, _ := strings.CutSuffix(timeframe, "m")
		minutes, err := strconv.Atoi(minValue)
		if err != nil {
			return false
		}
		if minutes < 0 || minutes > 2880 {
			return false
		}
	}
	if hours < 0 || hours > 48 {
		return false
	}
	return true
}

// validate is a generic validator for query structs.
func validate[T *BreakingNewsQuery | *CryptoNewsQuery | *HistoricalNewsQuery | *SourcesQuery](query T) error {
	// dereference pointer with Elem() if needed
	v := reflect.ValueOf(query).Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("query must be a struct")
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		tag := field.Tag.Get("validate")
		if value.IsZero() || tag == "" {
			continue
		}

		// split comma-separated rules
		// each rule is of the form "rulename:rulevalue"
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			if rule == "custom" {
				// Use custom validators
				if field.Name == "Timeframe" {
					if !isValidTimeframe(value.String()) {
						return fmt.Errorf("invalid timeframe: %s", value.String())
					}
					continue
				}
				if field.Name == "ExcludeFields" {
					for i := 0; i < value.Len(); i++ {
						field := value.Index(i).String()
						if !isValidField(field) {
							return fmt.Errorf("invalid field \"%v\" in ExcludeFields", field)
						}
						continue
					}
				}
				continue
			}
			parts := strings.Split(rule, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid validation rule: %s", rule)
			}
			ruleName := parts[0]
			ruleValue := parts[1]
			switch ruleName {
			case "min":
				val, err := strconv.ParseInt(ruleValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid %v value: %v", ruleName, ruleValue)
				}
				if value.Int() < val {
					return fmt.Errorf("field %s must be greater than or equal to %v", field.Name, ruleValue)
				}
			case "max":
				val, err := strconv.ParseInt(ruleValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid %v value: %v", ruleName, ruleValue)
				}
				if value.Int() > val {
					return fmt.Errorf("field %v must be less than or equal to %v", field.Name, ruleValue)
				}
			case "in":
				var fieldValues []string
				switch value.Kind() {
				case reflect.Slice:
					fieldValues = make([]string, value.Len())
					for i := 0; i < value.Len(); i++ {
						fieldValues[i] = fmt.Sprintf("%v", value.Index(i).Interface())
					}
				case reflect.String:
					fieldValues = []string{value.String()}
				default:
					return fmt.Errorf("invalid field  valuse %v in %v: %v", value.Interface(), field.Name, ruleValue)
				}
				for _, val := range fieldValues {
					if !slices.Contains(allowedValues[ruleValue], val) {
						return fmt.Errorf("invalid value field in %v: %v", field.Name, val)
					}
				}
			case "maxlen":
				val, err := strconv.Atoi(ruleValue)
				if err != nil {
					return fmt.Errorf("invalid %v value: %v", ruleName, ruleValue)
				}
				if value.Len() > val {
					return fmt.Errorf("field %v cannot be longer than %v", field.Name, ruleValue)
				}
			case "time":
				t := value.Interface().(time.Time)
				switch ruleValue {
				case "past":
					if t.After(time.Now()) {
						return fmt.Errorf("%v must be in the past", field.Name)
					}
				case "future":
					if t.Before(time.Now()) {
						return fmt.Errorf("%v must be in the future", field.Name)
					}
				default:
					return fmt.Errorf("invalid %v validation rule: %v", ruleName, ruleValue)
				}
			case "mutex":
				if !v.FieldByName(ruleValue).IsZero() {
					return fmt.Errorf("%v and %v cannot be used together", field.Name, ruleValue)
				}
			}
		}
	}
	return nil
}

// structToMap converts a struct into a map of query parameters, handling slices.
func structToMap(s interface{}) (map[string]string, error) {

	result := make(map[string]string)
	// dereference pointer with Elem() if needed
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
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
