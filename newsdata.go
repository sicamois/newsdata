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
// It provides methods to fetch news data and manage API interactions.
// The client handles HTTP requests, authentication, and logging configurations.
type NewsdataClient struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	logger      *slog.Logger
	LatestNews  *NewsService
	NewsArchive *NewsService
	CryptoNews  *NewsService
	Sources     *SourcesService
}

type clientOptions struct {
	apiKey             string
	customLoggerWriter io.Writer
	loggerLevel        slog.Level
	timeout            time.Duration
}

// ClientOption is a functional option for configuring the NewsdataClient.
type ClientOption func(*clientOptions)

// WithAPIKey sets the API key for the client.
// If no API key is provided via options, it attempts to read from the NEWSDATA_API_KEY
// environment variable. It will panic if no API key is available.
func WithAPIKey(apiKey string) ClientOption {
	return func(o *clientOptions) {
		o.apiKey = apiKey
	}
}

// WithCustomLoggerWriter sets a custom logger writer for the client.
// If no custom logger writer is provided, the client will use the default logger.
func WithCustomLoggerWriter(w io.Writer) ClientOption {
	return func(o *clientOptions) {
		o.customLoggerWriter = w
	}
}

// WithTimeout sets the timeout for the client.
// If no timeout is provided, the client will use a default timeout of 5 seconds.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithLogLevel sets the log level for the client.
// If no log level is provided, the client will use a default log level of slog.LevelInfo.
func WithLogLevel(level slog.Level) ClientOption {
	return func(o *clientOptions) {
		o.loggerLevel = level
	}
}

// NewClient creates a new NewsData API client with the provided options.
// If no API key is provided via options, it attempts to read from the NEWSDATA_API_KEY
// environment variable. It will panic if no API key is available.
func NewClient(opts ...ClientOption) *NewsdataClient {
	options := &clientOptions{
		timeout:     5 * time.Second,
		loggerLevel: slog.LevelInfo,
	}
	for _, opt := range opts {
		opt(options)
	}
	if options.apiKey == "" {
		options.apiKey = os.Getenv("NEWSDATA_API_KEY")
		if options.apiKey == "" {
			panic("NEWSDATA_API_KEY is not set")
		}
	}

	client := &NewsdataClient{
		// newsdata.io API base URL
		baseURL: "https://newsdata.io/api/1",
		// newsdata.io API key
		apiKey: options.apiKey,
		// HTTP client is a *http.Client that can be customized
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
	}
	if options.customLoggerWriter != nil {
		client.logger = newCustomLogger(options.customLoggerWriter, options.loggerLevel)
	} else {
		slog.SetLogLoggerLevel(options.loggerLevel)
		client.logger = slog.Default()
	}
	client.LatestNews = client.newLatestNewsService()
	client.NewsArchive = client.newNewsArchiveService()
	client.CryptoNews = client.newCryptoNewsService()
	client.Sources = client.newSourcesService()
	return client
}

// Logger returns the client's configured logger instance.
func (c *NewsdataClient) Logger() *slog.Logger {
	return c.logger
}

// errorResponse represents the API response when an error happened.
type errorResponse struct {
	Status string `json:"status"` // Response status ("error")
	Error  struct {
		Message string `json:"message"` // Error message
		Code    string `json:"code"`    // Error code
	} `json:"results"`
}

// buildHttpRequest creates an HTTP request for the specified endpoint with the given parameters.
// It constructs the full URL with query parameters and returns the prepared request.
func (c *NewsdataClient) buildHttpRequest(endpoint endpoint, params requestParams) (*http.Request, error) {
	reqURL, err := url.Parse(fmt.Sprintf("%s/%s", c.baseURL, string(endpoint)))
	if err != nil {
		return nil, fmt.Errorf("buildHttpRequest: error parsing URL - baseURL: %s, endpoint: %s: %w", c.baseURL, endpoint, err)
	}

	// Convert struct-based query parameters to URL query parameters.
	query := reqURL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Create the HTTP request.
	httpReq, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("buildHttpRequest: error creating request - url: %s: %w", reqURL.String(), err)
	}
	return httpReq, nil
}

// fetch sends an HTTP request and decodes the response.
func (c *NewsdataClient) fetch(context context.Context, endpoint endpoint, params requestParams) ([]byte, error) {
	start := time.Now()

	httpReq, err := c.buildHttpRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("fetch: error building HTTP request: %w", err)
	}

	// Create and execute the HTTP request.
	defer func() {
		c.logger.Debug("newsdata: fetch - response", "url", httpReq.URL.String(), "duration", time.Since(start))
	}()
	httpReq.Header.Set("X-ACCESS-KEY", c.apiKey)
	httpReq = httpReq.WithContext(context)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fetch - error executing request - url: %s: %w", httpReq.URL.String(), err)
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("fetch - error reading response body - url: %s: %w", httpReq.URL.String(), err)
	}

	// Handle non-200 status codes.
	if resp.StatusCode != http.StatusOK {
		var errorData errorResponse
		if err := json.Unmarshal(body, &errorData); err != nil {
			return nil, fmt.Errorf("fetch - error unmarshalling error response - url: %s: %w", httpReq.URL.String(), err)
		}
		return nil, fmt.Errorf("fetch - error reading response body - url: %s: %w", httpReq.URL.String(), errors.New(errorData.Error.Message))
	}

	return body, nil
}
