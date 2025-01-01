package newsdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

// NewsdataClient is the base client to access NewsData API.
// It provides methods to fetch news data.
//
// It handles the HTTP client and the Logger configurations.
type NewsdataClient struct {
	apiKey     string
	baseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

// newClient creates a new  NewsdataClient with default settings.
//
// Timeout is set to 5 seconds by default.
func NewClient(apiKey string) *NewsdataClient {
	logger := newCustomLogger(os.Stdout, slog.LevelInfo)
	return &NewsdataClient{
		// newsdata.io API key
		apiKey: apiKey,
		// newsdata.io API base URL
		baseURL: "https://newsdata.io/api/1",
		// HTTP client is a *http.Client that can be customized
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		// Logger is a *slog.Logger that can be customized
		Logger: logger,
	}
}

// DateTime is a wrapper around time.Time, used to format date as defined by the API
type DateTime struct {
	time.Time
}

// SentimentStats represents the sentiment stats for a news Article.
type SentimentStats struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}

// Article represents a news Article.
//
// See https://newsdata.io/documentation/#http_response
type Article struct {
	Id             string         `json:"Article_id"`
	Title          string         `json:"title"`
	Link           string         `json:"link"`
	Keywords       []string       `json:"keywords"`
	Creator        []string       `json:"creator"`
	VideoURL       string         `json:"video_url"`
	Description    string         `json:"description"`
	Content        string         `json:"content"`
	PubDate        DateTime       `json:"pubDate"`
	PubDateTZ      string         `json:"pubDateTZ"`
	ImageURL       string         `json:"image_url"`
	SourceId       string         `json:"source_id"`
	SourcePriority int            `json:"source_priority"`
	SourceName     string         `json:"source_name"`
	SourceURL      string         `json:"source_url"`
	SourceIconURL  string         `json:"source_icon"`
	Language       string         `json:"language"`
	Countries      []string       `json:"country"`
	Categories     []string       `json:"category"`
	AiTags         Tags           `json:"ai_tag"`
	Sentiment      string         `json:"sentiment"`
	SentimentStats SentimentStats `json:"sentiment_stats"`
	AiRegions      Tags           `json:"ai_region"`
	Coin           []string       `json:"coin"`
	Duplicate      bool           `json:"duplicate"`
}

// newsResponse represents the news API response.
//
// See https://newsdata.io/documentation/#http_response
type newsResponse struct {
	Status       string    `json:"status"`       // Response status ("success" or error message)
	TotalResults int       `json:"totalResults"` // Total number of Articles matching the query
	Articles     []Article `json:"results"`      // Array of Articles
	NextPage     string    `json:"nextPage"`     // Next page token
}

// Source represents a news source.
//
// See https://newsdata.io/documentation/#news-sources
type Source struct {
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
//
// See https://newsdata.io/documentation/#news-sources
type sourcesResponse struct {
	Status       string   `json:"status"`       // Response status ("success" or error message)
	TotalResults int      `json:"totalResults"` // Total number of news sources matching the query
	Sources      []Source `json:"results"`      // Array of news sources
}

// errorResponse represents the API response when an error happened.
type errorResponse struct {
	Status string `json:"status"` // Response status ("error")
	Error  struct {
		Message string `json:"message"` // Error message
		Code    string `json:"code"`    // Error code
	} `json:"results"`
}

// fetch sends an HTTP request and decodes the response.
func (c *NewsdataClient) fetch(endpoint string, q interface{}) ([]byte, error) {
	// Construct the full URL with query parameters.
	reqURL, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	// Convert struct-based query parameters to URL query parameters.
	query := reqURL.Query()
	paramMap, err := structToMap(q)
	if err != nil {
		return nil, err
	}
	for key, value := range paramMap {
		// Dirty fix for date format - value in map is "2024-12-01 00:00:00 +0000 UTC" and it should be "2024-12-01"
		if key == "from_date" || key == "to_date" {
			query.Add(key, value[:10])
		} else {
			query.Add(key, value)
		}
	}
	reqURL.RawQuery = query.Encode()

	// Create and execute the HTTP request.
	c.Logger.Debug("Request", "url", reqURL.String())
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s - url: %s", err.Error(), reqURL.String())
	}
	req.Header.Set("X-ACCESS-KEY", c.apiKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s - url: %s", err.Error(), reqURL.String())
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("%s - url: %s", err.Error(), reqURL.String())
	}

	// Handle non-200 status codes.
	if resp.StatusCode != http.StatusOK {
		var errorData errorResponse
		if err := json.Unmarshal(body, &errorData); err != nil {
			return nil, fmt.Errorf("%s - url: %s", err.Error(), reqURL.String())
		}
		slog.Error("Error reading response body", "error", errors.New(errorData.Error.Message), "url", reqURL.String())
		return nil, fmt.Errorf("%s - url: %s", errorData.Error.Message, reqURL.String())
	}

	return body, nil
}

func (c *NewsdataClient) fetchNews(endpoint string, query interface{}) (*newsResponse, error) {
	body, err := c.fetch(endpoint, query)
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

func (c *NewsdataClient) streamArticles(endpoint string, query pagerValider, maxResults int) (<-chan Article, <-chan error) {
	out := make(chan Article)
	errChan := make(chan error)
	go func() {
		defer close(out)
		defer close(errChan)
		if err := query.validate(); err != nil {
			errChan <- err
			return
		}
		page := ""
		index := 0
		for {
			query.setPage(page)
			res, err := c.fetchNews(endpoint, query)
			if err != nil {
				errChan <- err
				return
			}
			if maxResults == 0 {
				maxResults = res.TotalResults
			}
			for _, article := range res.Articles {
				if index < maxResults {
					out <- article
					index++
				} else {
					return
				}
			}
			page = res.NextPage
		}
	}()
	return out, errChan
}

// fetchArticles fetches news Articles from the API.
func (c *NewsdataClient) fetchArticles(endpoint string, query pagerValider, maxResults int) (*[]Article, error) {
	articleChan, errChan := c.streamArticles(endpoint, query, maxResults)
	articles := []Article{}
	for {
		select {
		case article, ok := <-articleChan:
			if !ok {
				// Channel is closed, all articles have been processed
				return &articles, nil
			}
			// Process each article
			articles = append(articles, article)
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		}
	}
}

// Get the latest news Articles in real-time from various sources worldwide.
// Filter by categories, countries, languages and more.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetBreakingNews(query BreakingNewsRequest, maxResults int) (*[]Article, error) {
	return c.fetchArticles("/latest", &query, maxResults)
}

// StreamBreakingNews creates a "pipeline" (an Article channel) of breaking news Articles from the API and passes them to an Article channel for async processing.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) StreamBreakingNews(query BreakingNewsRequest, maxResults int) (<-chan Article, <-chan error) {
	return c.streamArticles("/latest", &query, maxResults)
}

// Get cryptocurrency-related news with additional filters like coin symbols, sentiment analysis, and specialized crypto tags.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetCryptoNews(query CryptoNewsRequest, maxResults int) (*[]Article, error) {
	return c.fetchArticles("/crypto", &query, maxResults)
}

// StreamCryptoNews creates a "pipeline" (an Article channel) cryptocurrency-related news Articles from the API and passes them to an Article channel for async processing.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) StreamCryptoNews(query CryptoNewsRequest, maxResults int) (<-chan Article, <-chan error) {
	return c.streamArticles("/crypto", &query, maxResults)
}

// Search through news archives with date range filters while maintaining all filtering capabilities of breaking news.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetHistoricalNews(query HistoricalNewsRequest, maxResults int) (*[]Article, error) {
	if err := query.validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/archive", &query, maxResults)
}

// StreamHistoricalNews creates a "pipeline" (an Article channel) of historical news Articles from the API and passes them to an Article channel for async processing.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) StreamHistoricalNews(query HistoricalNewsRequest, maxResults int) (<-chan Article, <-chan error) {
	return c.streamArticles("/archive", &query, maxResults)
}

// fetchSources fetches news sources from the API.
func (c *NewsdataClient) fetchSources(endpoint string, query *SourcesRequest) (*[]Source, error) {
	sources := &[]Source{}

	body, err := c.fetch(endpoint, query)
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

// Get information about available news sources with filters for country, category, language and priority level.
func (c *NewsdataClient) GetSources(query SourcesRequest) (*[]Source, error) {
	if err := query.validate(); err != nil {
		return nil, err
	}
	return c.fetchSources("/sources", &query)
}

// SetTimeout sets the HTTP client timeout.
func (c *NewsdataClient) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

// CustomizeLogging customizes the logger used by the client.
func (c *NewsdataClient) CustomizeLogging(w io.Writer, level slog.Level) {
	customLogger := newCustomLogger(w, level)
	c.Logger = customLogger
}

// EnableDebug enables debug logging.
func (c *NewsdataClient) EnableDebug() {
	w := c.Logger.Handler().(*levelHandler).writer
	c.Logger = newCustomLogger(w, slog.LevelDebug)
}

// DisableDebug disables debug logging.
func (c *NewsdataClient) DisableDebug() {
	w := c.Logger.Handler().(*levelHandler).writer
	c.Logger = newCustomLogger(w, slog.LevelInfo)
}
