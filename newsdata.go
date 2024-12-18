package newsdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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

// BreakingNewsQuery represents the query parameters for the news endpoint.
type BreakingNewsQuery struct {
	Id                []string `query:"id"`              // List of article IDs
	Query             string   `query:"q"`               // Search term
	QueryInTitle      string   `query:"qInTitle"`        // Search term in article title
	QueryInMetadata   string   `query:"qInMeta"`         // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe         string   `query:"timeframe"`       // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Categories        []string `query:"category"`        // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory"` // List of categories to exclude
	Countries         []string `query:"country"`         // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language"`        // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain"`          // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl"`       // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain"`   // List of domains to exclude
	ExcludeFields     []string `query:"excludefield"`    // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain"`  // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`        // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content"`    // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image             string   `query:"image"`           // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video             string   `query:"video"`           // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	RemoveDuplicates  bool     `query:"removeduplicate"` // If set to true, duplicate articles will be removed from the results
	Size              int      `query:"size"`            // Number of results per page
	Page              string   `query:"page"`            // Page ref
}

func (p *BreakingNewsQuery) setPage(page string) {
	p.Page = page
}

// CryptoNewsQuery represents the query parameters for the crypto news endpoint.
type CryptoNewsQuery struct {
	Id               []string  `query:"id"`              // List of article IDs
	Coins            []string  `query:"coins"`           // List of coins (e.g., ["btc","eth","usdt"])
	From             time.Time `query:"from_date"`       // From date
	To               time.Time `query:"to_date"`         // To date
	Query            string    `query:"q"`               // Search term
	QueryInTitle     string    `query:"qInTitle"`        // Search term in article title
	QueryInMetadata  string    `query:"qInMeta"`         // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe        string    `query:"timeframe"`       // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Languages        []string  `query:"language"`        // List of language codes (e.g., ["en", "es"])
	Tags             []string  `query:"tag"`             // List of tags (e.g., ["blockchain", "liquidity", "scam"])
	Sentiment        string    `query:"sentiment"`       // List of sentiment : "positive", "negative" or "neutral"]
	Domains          []string  `query:"domain"`          // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls       []string  `query:"domainurl"`       // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains   []string  `query:"excludedomain"`   // List of domains to exclude
	ExcludeFields    []string  `query:"excludefield"`    // List of fields to exclude
	PriorityDomain   string    `query:"prioritydomain"`  // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	Timezone         string    `query:"timezone"`        // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent      string    `query:"full_content"`    // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image            string    `query:"image"`           // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video            string    `query:"video"`           // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	RemoveDuplicates bool      `query:"removeduplicate"` // If set to true, duplicate articles will be removed from the results
	Size             int       `query:"size"`            // Number of results per page
	Page             string    `query:"page"`            // Page ref
}

func (p *CryptoNewsQuery) setPage(page string) {
	p.Page = page
}

// HistoricalNewsQuery represents the query parameters for the news archive endpoint.
type HistoricalNewsQuery struct {
	Id                []string `query:"id"`              // List of article IDs
	From              DateTime `query:"from_date"`       // From date
	To                DateTime `query:"to_date"`         // To date
	Query             string   `query:"q"`               // Search term
	QueryInTitle      string   `query:"qInTitle"`        // Search term in article title
	QueryInMetadata   string   `query:"qInMeta"`         // Search term in article metadata (titles, URL, meta keywords and meta description)
	Categories        []string `query:"category"`        // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory"` // List of categories to exclude
	Countries         []string `query:"country"`         // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language"`        // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain"`          // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl"`       // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain"`   // List of domains to exclude
	ExcludeFields     []string `query:"excludefield"`    // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain"`  // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`        // Search the news articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content"`    // If set to 1, only the articles with full_content response object will be returned, if set to 0, only the articles without full_content response object will be returned
	Image             string   `query:"image"`           // If set to 1, only the articles with featured image will be returned, if set to 0, only the articles without featured image will be returned
	Video             string   `query:"video"`           // If set to 1, only the articles with video will be returned, if set to 0, only the articles without video will be returned
	Size              int      `query:"size"`            // Number of results per page
	Page              string   `query:"page"`            // Page ref
}

func (p *HistoricalNewsQuery) setPage(page string) {
	p.Page = page
}

// SourcesQuery represents the query parameters for the news archive endpoint.
type SourcesQuery struct {
	Country        string `query:"country"`        // Country codes (e.g. "us", "uk")
	Category       string `query:"category"`       // Category (e.g. "technology", "sports")
	Language       string `query:"language"`       // Language code (e.g. "en", "es")
	PriorityDomain string `query:"prioritydomain"` // Search the news articles only from top news domains. Possible values : Top, Medium, Low
	DomainUrl      string `query:"domainurl"`      //Domain URLs (e.g. "nytimes.com", "bbc.com", "bbc.co.uk")
}

// newsResponse represents the API response.
type newsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []article `json:"results"`
	NextPage     string    `json:"nextPage"`
}

// sourcesResponse represents the news sources API response.
type sourcesResponse struct {
	Status       string   `json:"status"`
	TotalResults int      `json:"totalResults"`
	Sources      []source `json:"results"`
}

// errorResponse represents the error response.
type errorResponse struct {
	Status string `json:"status"`
	Error  struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"results"`
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

type DateTime struct {
	time.Time
}

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

// NewClient creates a new newsdataClient with default settings.
// nbArticlesMax is the maximum number of articles to fetch.
// If set to 0, no limit is applied.
func NewClient(apiKey string, nbArticlesMax int) *newsdataClient {
	return &newsdataClient{
		APIKey:  apiKey,
		BaseURL: "https://newsdata.io/api/1", // Base URL from the documentation
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Logger:     slog.Default(),
		maxResults: nbArticlesMax,
	}
}

// SetTimeout sets the HTTP client timeout.
func (c *newsdataClient) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

// Logger returns the client logger
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

// fetch sends an HTTP request and return the body as []byte.
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

// GetBreakingNews fetches news based on query parameters.
func (c *newsdataClient) GetBreakingNews(query BreakingNewsQuery) (*[]article, error) {
	return c.fetchArticles("/latest", &query, c.maxResults)
}

// GetCryptoNews fetches crypto news based on query parameters.
func (c *newsdataClient) GetCryptoNews(query CryptoNewsQuery) (*[]article, error) {
	return c.fetchArticles("/crypto", &query, c.maxResults)
}

// GetHistoricalNews fetches news archive based on query parameters.
func (c *newsdataClient) GetHistoricalNews(query HistoricalNewsQuery) (*[]article, error) {
	return c.fetchArticles("/archive", &query, c.maxResults)
}

// GetSources fetches news archive based on query parameters.
func (c *newsdataClient) GetSources(query SourcesQuery) (*[]source, error) {
	return c.fetchSources("/sources", &query)
}
