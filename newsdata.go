package newsdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Level int

// baseClient handles the HTTP client configuration.
type baseClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
	maxResults int
}

// article represents a news article.
type article struct {
	Id             string   `json:"article_id"`
	Title          string   `json:"title"`
	Link           string   `json:"link"`
	Keywords       []string `json:"keywords"`
	Creator        []string `json:"creator"`
	VideoURL       string   `json:"video_url"`
	Description    string   `json:"description"`
	Content        string   `json:"content"`
	PubDate        DateTime `json:"pubDate"`
	PubDateTZ      string   `json:"pubDateTZ"`
	ImageURL       string   `json:"image_url"`
	SourceId       string   `json:"source_id"`
	SourcePriority int      `json:"source_priority"`
	SourceName     string   `json:"source_name"`
	SourceURL      string   `json:"source_url"`
	SourceIconURL  string   `json:"source_icon"`
	Language       string   `json:"language"`
	Countries      []string `json:"country"`
	Categories     []string `json:"category"`
	AiTag          string   `json:"ai_tag"`
	Sentiment      string   `json:"sentiment"`
	SentimentStats string   `json:"sentiment_stats"`
	AiRegion       string   `json:"ai_region"`
	AiOrganization string   `json:"ai_org"`
	Duplicate      bool     `json:"duplicate"`
}

type DateTime struct {
	time.Time
}

// newsResponse represents the API response.
type newsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []article `json:"results"`
	NextPage     string    `json:"nextPage"`
}

type errorResponse struct {
	Status string `json:"status"`
	Error  struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"results"`
}

// source represents a news source.
type source struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Url         string   `json:"url"`
	IconUrl     string   `json:"icon"`
	Priority    int      `json:"priority"`
	Description string   `json:"description"`
	Categories  []string `json:"category"`
	Languages   []string `json:"language"`
	Countries   []string `json:"country"`
	LastFetch   DateTime `json:"last_fetch"`
}

// sourcesResponse represents the news sources API response.
type sourcesResponse struct {
	Status       string   `json:"status"`
	TotalResults int      `json:"totalResults"`
	Sources      []source `json:"results"`
}

func (t *DateTime) UnmarshalJSON(b []byte) error {
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

// newBaseClient creates a new baseClient with default settings.
func newBaseClient(apiKey string) *baseClient {
	return &baseClient{
		APIKey:  apiKey,
		BaseURL: "https://newsdata.io/api/1", // Base URL from the documentation
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Logger: slog.Default(),
	}
}

// fetch sends an HTTP request and decodes the response.
func (c *baseClient) fetch(endpoint string, params interface{}) ([]byte, error) {
	// Construct the full URL with query parameters.
	reqURL, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}

	// Convert struct-based query parameters to URL query parameters.
	query := reqURL.Query()
	paramMap, err := structToQuery(params)
	if err != nil {
		return nil, err
	}
	for key, value := range paramMap {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Create and execute the HTTP request.
	c.Logger.Debug("Request", "url", reqURL.String())
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-ACCESS-KEY", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	// Handle non-200 status codes.
	if resp.StatusCode != http.StatusOK {
		var errorData errorResponse
		if err := json.Unmarshal(body, &errorData); err != nil {
			return nil, err
		}
		return nil, errors.New(errorData.Error.Message)
	}

	return body, nil
}

func (c *baseClient) fetchNews(endpoint string, params interface{}) (*newsResponse, error) {
	body, err := c.fetch(endpoint, params)
	if err != nil {
		return nil, err
	}
	// Decode the JSON response.
	var data newsResponse
	if err := json.Unmarshal(body, &data); err != nil { // Parse []byte to go struct pointer
		return nil, err
	}
	return &data, nil
}

// pageSetter is an interface for setting the page parameter.
type pageSetter interface {
	setPage(string)
}

// fetchArticles fetches news articles from the API.
func (c *baseClient) fetchArticles(endpoint string, params pageSetter, maxResults int) (*[]article, error) {
	articles := &[]article{}

	page := ""

	// Keep fetching pages until maxResults is reached or no more results.
	for len(*articles) < maxResults || maxResults == 0 {
		params.setPage(page)

		res, err := c.fetchNews(endpoint, params)
		if err != nil {
			return nil, err
		}
		c.Logger.Debug("Response", "status", res.Status, "totalResults", res.TotalResults, "#articles", len(res.Articles), "nextPage", res.NextPage)

		if maxResults == 0 || res.TotalResults < maxResults {
			maxResults = res.TotalResults
		}

		// Append results to the aggregate slice.
		*articles = append(*articles, res.Articles...)

		// Update page
		if res.NextPage == "" || len(*articles) >= maxResults {
			if res.NextPage == "" {
				c.Logger.Debug("All results fetched")
			}
			if len(*articles) >= maxResults {
				c.Logger.Debug("Max results reached", "maxResults", maxResults, "total #articles", len(*articles))
			}
			break
		}
		page = res.NextPage
	}

	// Trim results to maxResults if necessary.
	if len(*articles) > maxResults {
		*articles = (*articles)[:maxResults]
	}

	return articles, nil
}

// fetchSources fetches news sources from the API.
func (c *baseClient) fetchSources(endpoint string, params SourcesQueryParams) (*[]source, error) {
	sources := &[]source{}

	body, err := c.fetch(endpoint, params)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response.
	var res sourcesResponse
	if err := json.Unmarshal(body, &res); err != nil { // Parse []byte to go struct pointer
		return nil, err
	}

	c.Logger.Debug("Response", "status", res.Status, "totalResults", res.TotalResults, "#sources", len(res.Sources))

	// Append results to the aggregate slice.
	*sources = append(*sources, res.Sources...)

	return sources, nil
}

// NewsdataClient is the main client that composes all services.
type newsdataClient struct {
	baseClient  *baseClient
	logger      **slog.Logger
	LatestNews  *latestNewsService
	CryptoNews  *cryptoNewsService
	NewsArchive *newsArchiveService
	Sources     *sourcesService
}

// NewNewsdataClient creates a new instance of NewsdataClient.
func NewClient(apiKey string) *newsdataClient {
	baseClient := newBaseClient(apiKey)
	return &newsdataClient{
		baseClient: baseClient,
		logger:     &baseClient.Logger,
		LatestNews: &latestNewsService{
			client:   baseClient,
			endpoint: "/latest",
		},
		CryptoNews: &cryptoNewsService{
			client:   baseClient,
			endpoint: "/crypto",
		},
		NewsArchive: &newsArchiveService{
			client:   baseClient,
			endpoint: "/archive",
		},
		Sources: &sourcesService{
			client:   baseClient,
			endpoint: "/sources",
		},
	}
}

// SetTimeout sets the HTTP client timeout.
func (c *newsdataClient) SetTimeout(timeout time.Duration) {
	c.baseClient.HTTPClient.Timeout = timeout
}

// Logger returns the client logger
func (c *newsdataClient) Logger() *slog.Logger {
	return c.baseClient.Logger
}

// CustomizeLogging customizes the logger used by the client.
func (c *newsdataClient) CustomizeLogging(w io.Writer, level slog.Level) {
	customLogger := NewCustomLogger(w, level)
	c.baseClient.Logger = customLogger
}

// EnableDebug enables debug logging.
func (c *newsdataClient) EnableDebug() {
	w := c.baseClient.Logger.Handler().(*LevelHandler).writer
	c.baseClient.Logger = NewCustomLogger(w, slog.LevelDebug)
}

// DisableDebug disables debug logging.
func (c *newsdataClient) DisableDebug() {
	w := c.baseClient.Logger.Handler().(*LevelHandler).writer
	c.baseClient.Logger = NewCustomLogger(w, slog.LevelInfo)
}

// LimitResultsToFirst limits the number of results returned by the client.
func (c *newsdataClient) LimitResultsToFirst(n int) error {
	if n < 0 {
		return fmt.Errorf("max returned results must be positive")
	}
	c.baseClient.maxResults = n
	return nil
}

//
// PARAMETERS HELPERS
//

// isValidCategory checks if a category is in the allowed list.
func isValidCategory(category string) bool {
	for _, allowed := range allowedCategories {
		if category == allowed {
			return true
		}
	}
	return false
}

// isValidCountry checks if a country code is in the allowed list.
func isValidCountry(countryCode string) bool {
	for _, allowed := range allowedCountries {
		if countryCode == allowed {
			return true
		}
	}
	return false
}

// isValidLanguage checks if a language code is in the allowed list.
func isValidLanguage(languageCode string) bool {
	for _, allowed := range allowedLanguages {
		if languageCode == allowed {
			return true
		}
	}
	return false
}

// isValidField checks if a field exists in the article struct.
func isValidField(field string) bool {
	articleFields := make([]string, 0)
	t := reflect.TypeOf(article{})
	for i := 0; i < t.NumField(); i++ {
		articleFields = append(articleFields, t.Field(i).Name)
	}
	for _, allowed := range articleFields {
		if field == allowed {
			return true
		}
	}
	return false
}

// isValidPriorityDomain checks if a priority domain is in the allowed list.
func isValidPriorityDomain(priorityDomain string) bool {
	for _, allowed := range allowedPriorityDomain {
		if priorityDomain == allowed {
			return true
		}
	}
	return false
}

// isValidSentiment checks if a sentiment is in the allowed list.
func isValidSentiment(sentiment string) bool {
	for _, allowed := range allowedSentiment {
		if sentiment == allowed {
			return true
		}
	}
	return false
}

// isValidTag checks if a tag is in the allowed list.
func isValidTag(tag string) bool {
	for _, allowed := range allowedTags {
		if tag == allowed {
			return true
		}
	}
	return false
}

// structToQuery converts a struct into a map of query parameters, handling slices.
func structToQuery(params interface{}) (map[string]string, error) {

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
