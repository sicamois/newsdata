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

// newsdataClient handles the HTTP client configuration.
type newsdataClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
	maxResults int
}

// pageSetter is an interface for setting the page parameter.
type pageSetter interface {
	setPage(string)
}

// DateTime is used to format date as defined by the API
type DateTime struct {
	time.Time
}

// Tags is a structure for AI tags, AI Regions to handle the case where the API returns an error (typically "ONLY AVAILABLE IN PROFESSIONAL AND CORPORATE PLANS")
type Tags []Tag
type Tag struct {
	string
}

// BreakingNewsQuery represents the query parameters for the breaking news endpoint. See https://newsdata.io/documentation/#latest-news
type BreakingNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of article IDs
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe         string   `query:"timeframe"`                                            // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])
	Tags              []string `query:"tag" validate:"maxlen:5,in:tags"`                      // List of AI tags
	Sentiment         string   `query:"sentiment" validate:"in:sentiments"`                   // Filter by sentiment ("positive", "negative", "neutral")
	Regions           []string `query:"region" validate:"maxlen:5"`                           // List of regions
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`                                             // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	RemoveDuplicates  string   `query:"removeduplicate" validate:"in:binaries"`               // If set to true, duplicate articles will be removed from the results
	Size              int      `query:"size" validate:"min:1,max:50"`                         // Number of results per page
	Page              string   `query:"page"`                                                 // Page ref
}

// setPage sets the page parameter
func (q *BreakingNewsQuery) setPage(page string) {
	q.Page = page
}

// Validate validates the BreakingNewsQuery struct, ensuring all fields are valid.
func (query *BreakingNewsQuery) Validate() error {
	return validate(query)
}

// HistoricalNewsQuery represents the query parameters for the historical news endpoint. See https://newsdata.io/documentation/#news-archive
type HistoricalNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of article IDs
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in article metadata (titles, URL, meta keywords and meta description)
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	From              DateTime `query:"from_date" validate:"time:past"`                       // From date
	To                DateTime `query:"to_date" validate:"time:past"`                         // To date
	Timezone          string   `query:"timezone"`                                             // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	Size              int      `query:"size" validate:"min:1,max:50"`                         // Number of results per page
	Page              string   `query:"page"`                                                 // Page ref
}

// setPage sets the page parameter
func (q *HistoricalNewsQuery) setPage(page string) {
	q.Page = page
}

// Validate validates the HistoricalNewsQuery struct, ensuring all fields are valid.
func (query *HistoricalNewsQuery) Validate() error {
	return validate(query)
}

// CryptoNewsQuery represents the query parameters for the crypto news endpoint. See https://newsdata.io/documentation/#crypto-news
type CryptoNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of article IDs
	Coins             []string `query:"coin"`                                                 // List of coins
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe         string   `query:"timeframe"`                                            // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`                                             // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	RemoveDuplicates  string   `query:"removeduplicate" validate:"in:binaries"`               // If set to true, duplicate articles will be removed from the results
	Sentiment         string   `query:"sentiment" validate:"in:sentiments"`                   // Filter by sentiment ("positive", "negative", "neutral")
	Tags              []string `query:"tag" validate:"in:tags"`                               // Filter by crypto-specific tags
	From              DateTime `query:"from_date" validate:"time:past"`                       // From date
	To                DateTime `query:"to_date" validate:"time:past"`                         // To date
	Size              int      `query:"size" validate:"min:1,max:50"`                         // Number of results per page
	Page              string   `query:"page"`                                                 // Page ref
}

// setPage sets the page parameter
func (q *CryptoNewsQuery) setPage(page string) {
	q.Page = page
}

// Validate validates the CryptoNewsQuery struct, ensuring all fields are valid.
func (query *CryptoNewsQuery) Validate() error {
	return validate(query)
}

// SourcesQuery represents the query parameters for the sources endpoint. See https://newsdata.io/documentation/#news-sources
type SourcesQuery struct {
	Country        string `query:"country" validate:"maxlen:5,in:countries"`     // Filter by country code
	Language       string `query:"language" validate:"maxlen:5,in:languages"`    // Filter by language code
	Category       string `query:"category" validate:"maxlen:5,in:categories"`   // Filter by category (e.g., "technology")
	PriorityDomain string `query:"prioritydomain" validate:"in:priorityDomains"` // Filter by priority domain (possible values: "top", "medium", "low")
	domainurl      string `query:"domainurl" validate:"maxlen:5"`                // Filter by domain URL (e.g., "nytimes.com")
}

// Validate validates the HistoricalNewsQuery struct, ensuring all fields are valid.
func (query *SourcesQuery) Validate() error {
	return validate(query)
}

// newsResponse represents the news API response. See https://newsdata.io/documentation/#http_response
type newsResponse struct {
	Status       string    `json:"status"`       // Response status ("success" or error message)
	TotalResults int       `json:"totalResults"` // Total number of articles matching the query
	Articles     []article `json:"results"`      // Array of articles
	NextPage     string    `json:"nextPage"`     // Next page token
}

// sourcesResponse represents the news sources API response. See https://newsdata.io/documentation/#news-sources
type sourcesResponse struct {
	Status       string   `json:"status"`       // Response status ("success" or error message)
	TotalResults int      `json:"totalResults"` // Total number of news sources matching the query
	Sources      []source `json:"results"`      // Array of news sources
}

// errorResponse represents the API response when an error happened.
type errorResponse struct {
	Status string `json:"status"` // Response status ("error")
	Error  struct {
		Message string `json:"message"` // Error message
		Code    string `json:"code"`    // Error code
	} `json:"results"`
}

// SentimentStats represents the sentiment stats for a news article.
type SentimentStats struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}

// article represents a news article. See https://newsdata.io/documentation/#http_response
type article struct {
	Id             string         `json:"article_id"`
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

// source represents a news source. See https://newsdata.io/documentation/#news-sources
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

// newClient creates a new newsdataClient with default settings.
// nbArticlesMax is the maximum number of articles to fetch.
// If set to 0, no limit is applied.
// Timeout is set to 5 seconds by default.
func NewClient(apiKey string, nbArticlesMax int) *newsdataClient {
	logger := NewCustomLogger(os.Stdout, slog.LevelInfo)
	return &newsdataClient{
		APIKey:  apiKey,
		BaseURL: "https://newsdata.io/api/1", // Base URL from the documentation
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Logger:     logger,
		maxResults: nbArticlesMax,
	}
}

// fetch sends an HTTP request and decodes the response.
func (c *newsdataClient) fetch(endpoint string, q interface{}) ([]byte, error) {
	// Construct the full URL with query parameters.
	reqURL, err := url.Parse(c.BaseURL + endpoint)
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

func (c *newsdataClient) fetchNews(endpoint string, query interface{}) (*newsResponse, error) {
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

// fetchArticles fetches news articles from the API.
func (c *newsdataClient) fetchArticles(endpoint string, query pageSetter, maxResults int) (*[]article, error) {
	articles := &[]article{}

	page := ""

	// Keep fetching pages until maxResults is reached or no more results.
	for len(*articles) < maxResults || maxResults == 0 {
		query.setPage(page) // Set the page parameter

		body, err := c.fetch(endpoint, query)
		if err != nil {
			return nil, err
		}

		var res newsResponse
		if err := json.Unmarshal(body, &res); err != nil {
			return nil, err
		}

		c.Logger.Debug("Response", "status", res.Status, "totalResults", res.TotalResults, "#articles", len(res.Articles), "nextPage", res.NextPage)

		if maxResults == 0 || res.TotalResults < maxResults {
			maxResults = res.TotalResults
		}

		// Append results to the aggregate slice.
		*articles = append(*articles, res.Articles...)
		c.Logger.Debug("Articles retrieved", "#", len(*articles))

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

// GetBreakingNews fetches breaking news based on query parameters.
func (c *newsdataClient) GetBreakingNews(query BreakingNewsQuery) (*[]article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/latest", &query, c.maxResults)
}

// GetCryptoNews fetches crypto news based on query parameters.
func (c *newsdataClient) GetCryptoNews(query CryptoNewsQuery) (*[]article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/crypto", &query, c.maxResults)
}

// GetHistoricalNews fetches historical news based on query parameters.
func (c *newsdataClient) GetHistoricalNews(query HistoricalNewsQuery) (*[]article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/archive", &query, c.maxResults)
}

// fetchSources fetches news sources from the API.
func (c *newsdataClient) fetchSources(endpoint string, query *SourcesQuery) (*[]source, error) {
	sources := &[]source{}

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

// GetSources fetches news sources based on query parameters.
func (c *newsdataClient) GetSources(query SourcesQuery) (*[]source, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchSources("/sources", &query)
}

// SetTimeout sets the HTTP client timeout.
func (c *newsdataClient) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

// GetLogger returns the client logger
func (c *newsdataClient) GetLogger() *slog.Logger {
	return c.Logger
}

// CustomizeLogging customizes the logger used by the client.
func (c *newsdataClient) CustomizeLogging(w io.Writer, level slog.Level) {
	customLogger := NewCustomLogger(w, level)
	c.Logger = customLogger
}

// EnableDebug enables debug logging.
func (c *newsdataClient) EnableDebug() {
	w := c.Logger.Handler().(*LevelHandler).writer
	c.Logger = NewCustomLogger(w, slog.LevelDebug)
}

// DisableDebug disables debug logging.
func (c *newsdataClient) DisableDebug() {
	w := c.Logger.Handler().(*LevelHandler).writer
	c.Logger = NewCustomLogger(w, slog.LevelInfo)
}

// setNbArticlesMax limits the number of results returned by the client.
func (c *newsdataClient) setNbArticlesMax(n int) error {
	if n < 0 {
		return fmt.Errorf("Nb articles max must be positive")
	}
	c.maxResults = n
	return nil
}
