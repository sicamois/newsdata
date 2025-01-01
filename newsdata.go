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
	"os"
	"time"
)

// NewsdataClient is the base client to access NewsData API.
// It provides methods to fetch news data.
//
// It handles the HTTP client and the logger configurations.
type NewsdataClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// newClient creates a new  NewsdataClient with default settings.
//
// Timeout is set to 5 seconds by default.
func NewClient(apiKey string) *NewsdataClient {
	return &NewsdataClient{
		// newsdata.io API key
		apiKey: apiKey,
		// newsdata.io API base URL
		baseURL: "https://newsdata.io/api/1",
		// HTTP client is a *http.Client that can be customized
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		// logger is a *slog.logger that can be customized
		logger: newCustomLogger(os.Stdout, slog.LevelInfo),
	}
}

// SetTimeout sets the HTTP client timeout.
func (c *NewsdataClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// CustomizeLogging customizes the logger used by the client.
func (c *NewsdataClient) CustomizeLogging(w io.Writer, level slog.Level) {
	customlogger := newCustomLogger(w, level)
	c.logger = customlogger
}

// EnableDebug enables debug logging.
func (c *NewsdataClient) EnableDebug() {
	w := c.logger.Handler().(*levelHandler).writer
	c.logger = newCustomLogger(w, slog.LevelDebug)
}

// DisableDebug disables debug logging.
func (c *NewsdataClient) DisableDebug() {
	w := c.logger.Handler().(*levelHandler).writer
	c.logger = newCustomLogger(w, slog.LevelInfo)
}

// Logger() returns the logger
func (c *NewsdataClient) Logger() *slog.Logger {
	return c.logger
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

// Tags is is a wrapper around []string for coin-specific tags, AI tags & AI Regions, used to handle the case where the API returns a restriction message (typically "ONLY AVAILABLE IN PROFESSIONAL AND CORPORATE PLANS")
type Tags []string

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

// errorResponse represents the API response when an error happened.
type errorResponse struct {
	Status string `json:"status"` // Response status ("error")
	Error  struct {
		Message string `json:"message"` // Error message
		Code    string `json:"code"`    // Error code
	} `json:"results"`
}

// fetch sends an HTTP request and decodes the response.
func (c *NewsdataClient) fetch(context context.Context, endpoint string, params map[string]string) ([]byte, error) {
	// Construct the full URL with query parameters.
	reqURL, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	// Convert struct-based query parameters to URL query parameters.
	query := reqURL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Create and execute the HTTP request.
	c.logger.Debug("Request", "url", reqURL.String())
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	req = req.WithContext(context)
	if err != nil {
		return nil, fmt.Errorf("%s - url: %s", err.Error(), reqURL.String())
	}
	req.Header.Set("X-ACCESS-KEY", c.apiKey)
	resp, err := c.httpClient.Do(req)
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

func (c *NewsdataClient) fetchNews(req articleRequest) (*newsResponse, error) {
	body, err := c.fetch(req.context, req.service.Endpoint(), req.params)
	if err != nil {
		return nil, err
	}
	// Decode the JSON response.
	var data newsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// StreamArticles streams news Articles from the API.
func (c *NewsdataClient) StreamArticles(req articleRequest, maxResults int) (<-chan Article, <-chan error) {
	out := make(chan Article)
	errChan := make(chan error)
	go func() {
		defer close(out)
		defer close(errChan)
		page := ""
		index := 0
		for {
			if page != "" {
				req.params["page"] = page
			}
			res, err := c.fetchNews(req)
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

// GetArticles fetches news Articles from the API.
func (c *NewsdataClient) GetArticles(req articleRequest, maxResults int) (*[]Article, error) {
	articleChan, errChan := c.StreamArticles(req, maxResults)
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

// GetSources fetches news sources from the API.
func (c *NewsdataClient) GetSources(req sourceRequest) (*[]Source, error) {
	sources := &[]Source{}

	body, err := c.fetch(req.context, "/sources", req.params)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response.
	var res sourcesResponse
	if err := json.Unmarshal(body, &res); err != nil { // Parse []byte to go struct pointer
		return nil, err
	}

	c.logger.Debug("Response", "status", res.Status, "totalResults", res.TotalResults, "#sources", len(res.Sources))

	// Append results to the aggregate slice.
	*sources = append(*sources, res.Sources...)

	return sources, nil
}
