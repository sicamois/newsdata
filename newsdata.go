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
	"strings"
	"time"
)

// NewsDataClient is the base client to access NewsData API.
// It provides methods to fetch news data and manage API interactions.
//
// The client handles HTTP requests, authentication, and logging configurations.
type NewsDataClient struct {
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

// NewsDataClientOption is a functional option for configuring the NewsDataClient.
type NewsDataClientOption func(*clientOptions)

// WithAPIKey sets the API key for the client.
//
// If no API key is provided via options, it attempts to read from the NEWSDATA_API_KEY
// environment variable. It will panic if no API key is available.
func WithAPIKey(apiKey string) NewsDataClientOption {
	return func(o *clientOptions) {
		o.apiKey = apiKey
	}
}

// WithTimeout sets the global timeout for the http client.
//
// If no timeout is provided, the client will use a default timeout of 5 seconds.
func WithTimeout(timeout time.Duration) NewsDataClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// NewClient creates a new NewsData API client with the provided options.
//
// If no API key is provided via options, it attempts to read from the NEWSDATA_API_KEY
// environment variable. It will panic if no API key is available.
func NewClient(opts ...NewsDataClientOption) *NewsDataClient {
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

	client := &NewsDataClient{
		// newsdata.io API base URL
		baseURL: "https://newsdata.io/api/1",
		// newsdata.io API key
		apiKey: options.apiKey,
		// HTTP client is a *http.Client that can be customized
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
	}
	defaultLogger := *slog.Default()
	defaultCopy := &defaultLogger
	client.logger = defaultCopy.With(slog.String("package", "newsdata"))
	client.LatestNews = client.newLatestNewsService()
	client.NewsArchive = client.newNewsArchiveService()
	client.CryptoNews = client.newCryptoNewsService()
	client.Sources = client.newSourcesService()
	return client
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
func (c *NewsDataClient) buildHttpRequest(endpoint endpoint, params requestParams) (*http.Request, error) {
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
func (c *NewsDataClient) fetch(context context.Context, endpoint endpoint, params requestParams) ([]byte, error) {
	start := time.Now()

	httpReq, err := c.buildHttpRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("fetch: error building HTTP request: %w", err)
	}

	var resp *http.Response
	// Create and execute the HTTP request.
	defer func() {
		attrs := make([]any, 0, 3)
		attrs = append(attrs, slog.Group("request", "method", httpReq.Method, "url", httpReq.URL.String()))
		if resp != nil {
			headers := make([]string, 0, len(resp.Header))
			for key, value := range resp.Header {
				if key == "X-ACCESS-KEY" || key == "Date" || key == "Server" || key == "Vary" {
					continue
				}
				headers = append(headers, fmt.Sprintf("%s=%s", key, strings.Join(value, ", ")))
			}
			headersString := fmt.Sprintf("{%s}", strings.Join(headers, ", "))
			attrs = append(attrs, slog.Group("response", "status_code", resp.StatusCode, "status", resp.Status, "headers", headersString))
		} else {
			attrs = append(attrs, slog.String("response", "null"))
		}
		attrs = append(attrs, slog.Duration("duration", time.Since(start)))
		c.logger.Debug("request completed", attrs...)
	}()
	httpReq.Header.Set("X-ACCESS-KEY", c.apiKey)
	httpReq = httpReq.WithContext(context)

	resp, err = c.httpClient.Do(httpReq)
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
