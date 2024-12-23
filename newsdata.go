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
	"sync"
	"time"
)

// pageSetter is an interface for setting the page parameter.
type pageSetter interface {
	setPage(string)
}

// DateTime is a wrapper around time.Time, used to format date as defined by the API
type DateTime struct {
	time.Time
}

// Tags is is a wrapper around []string for coin-specific tags, AI tags & AI Regions, used to handle the case where the API returns a restriction message (typically "ONLY AVAILABLE IN PROFESSIONAL AND CORPORATE PLANS")
type Tags []string

// BreakingNewsQuery represents the query parameters for the breaking news endpoint.
//
// See https://newsdata.io/documentation/#latest-news
type BreakingNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of Article IDs
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in Article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in Article metadata (titles, URL, meta keywords and meta description)
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
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news Articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`                                             // Search the news Articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the Articles with full_content response object will be returned, if set to 0, only the Articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the Articles with featured image will be returned, if set to 0, only the Articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the Articles with video will be returned, if set to 0, only the Articles without video will be returned
	RemoveDuplicates  string   `query:"removeduplicate" validate:"in:binaries"`               // If set to true, duplicate Articles will be removed from the results
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

// HistoricalNewsQuery represents the query parameters for the historical news endpoint.
//
// See https://newsdata.io/documentation/#news-archive
type HistoricalNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of Article IDs
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in Article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in Article metadata (titles, URL, meta keywords and meta description)
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news Articles only from top news domains. Possible values : Top, Medium, Low
	From              DateTime `query:"from_date" validate:"time:past"`                       // From date
	To                DateTime `query:"to_date" validate:"time:past"`                         // To date
	Timezone          string   `query:"timezone"`                                             // Search the news Articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the Articles with full_content response object will be returned, if set to 0, only the Articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the Articles with featured image will be returned, if set to 0, only the Articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the Articles with video will be returned, if set to 0, only the Articles without video will be returned
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

// CryptoNewsQuery represents the query parameters for the crypto news endpoint.
//
// See https://newsdata.io/documentation/#crypto-news
type CryptoNewsQuery struct {
	Id                []string `query:"id"`                                                   // List of Article IDs
	Coins             []string `query:"coin"`                                                 // List of coins
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in Article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in Article metadata (titles, URL, meta keywords and meta description)
	Timeframe         string   `query:"timeframe"`                                            // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news Articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`                                             // Search the news Articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the Articles with full_content response object will be returned, if set to 0, only the Articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the Articles with featured image will be returned, if set to 0, only the Articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the Articles with video will be returned, if set to 0, only the Articles without video will be returned
	RemoveDuplicates  string   `query:"removeduplicate" validate:"in:binaries"`               // If set to true, duplicate Articles will be removed from the results
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

// SourcesQuery represents the query parameters for the sources endpoint.
//
// See https://newsdata.io/documentation/#news-sources
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

// newsResponse represents the news API response.
//
// See https://newsdata.io/documentation/#http_response
type newsResponse struct {
	Status       string    `json:"status"`       // Response status ("success" or error message)
	TotalResults int       `json:"totalResults"` // Total number of Articles matching the query
	Articles     []Article `json:"results"`      // Array of Articles
	NextPage     string    `json:"nextPage"`     // Next page token
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

// BatchInfos represents the information about the current batch of Articles being processed.
type BatchInfos struct {
	Num          int
	StartingTime time.Time
	Size         int
	TotalFetched int
	MaxResults   int
	TotalResults int
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
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Create and execute the HTTP request.
	c.Logger.Debug("Request", "url", reqURL.String())
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-ACCESS-KEY", c.apiKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

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

func (c *NewsdataClient) processArticles(endpoint string, query pageSetter, maxResults int, action func(*[]Article, BatchInfos) error) error {
	nbArticles := 0
	page := ""
	var wg sync.WaitGroup
	chanLen := 100
	if maxResults > 0 {
		chanLen = maxResults
	}
	errChan := make(chan error, chanLen)
	batchInfos := BatchInfos{Num: 0}

	// Keep fetching pages until maxResults is reached or no more results.
	for nbArticles < maxResults || maxResults == 0 {
		batchInfos.StartingTime = time.Now()
		query.setPage(page) // Set the page parameter

		res, err := c.fetchNews(endpoint, query)
		if err != nil {
			return err
		}
		c.Logger.Debug("Response", "status", res.Status, "totalResults", res.TotalResults, "#Articles", len(res.Articles), "nextPage", res.NextPage)
		if maxResults == 0 || res.TotalResults < maxResults {
			maxResults = res.TotalResults
		}
		// Append results to the aggregate slice.
		nbArticles += len(res.Articles)
		c.Logger.Debug("Articles retrieved", "#", nbArticles)

		// Update batchInfos
		batchInfos.Size = len(res.Articles)
		batchInfos.TotalFetched = nbArticles
		batchInfos.TotalResults = res.TotalResults
		if maxResults == 0 {
			batchInfos.MaxResults = res.TotalResults
		} else {
			batchInfos.MaxResults = maxResults
		}
		// Increment batch number should be here, because everuthing gets evaluated before the go routine
		batchInfos.Num++

		// Process articles
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Logger.Debug("Process articles with action", "endpoint", endpoint, "query", query, "batchInfos", batchInfos)
			if err := action(&res.Articles, batchInfos); err != nil {
				c.Logger.Error("Error processing articles", "error", err.Error())
				errChan <- err
			}
		}()

		// Check for errors as soon as possible
		if len(errChan) > 0 {
			err := <-errChan
			return fmt.Errorf("%v", err.Error())
		}

		// Update page
		if res.NextPage == "" || nbArticles >= maxResults {
			if res.NextPage == "" {
				c.Logger.Debug("All results fetched")
			}
			if nbArticles >= maxResults {
				c.Logger.Debug("Max results reached", "maxResults", maxResults, "total #Articles", nbArticles)
			}
			break
		}
		page = res.NextPage
	}
	wg.Wait()

	// Check for residual errors
	if len(errChan) > 0 {
		err := <-errChan
		return fmt.Errorf("%v", err.Error())
	}
	c.Logger.Debug(fmt.Sprintf("Batch processed %d/%d", batchInfos.TotalFetched, batchInfos.TotalResults), "batch", batchInfos.Num, "size", batchInfos.Size, "totalFetched", batchInfos.TotalFetched, "maxResults", batchInfos.MaxResults, "totalResults", batchInfos.TotalResults, "duration", time.Since(batchInfos.StartingTime))
	return nil
}

type safeArticles struct {
	mu       sync.Mutex
	articles *[]Article
}

// fetchArticles fetches news Articles from the API.
func (c *NewsdataClient) fetchArticles(endpoint string, query pageSetter, maxResults int) (*[]Article, error) {
	Articles := safeArticles{articles: &[]Article{}}

	addArticle := func(newArticles *[]Article, batchInfos BatchInfos) error {
		c.Logger.Info("Add articles", "batch", batchInfos.Num, "size", batchInfos.Size, "totalFetched", batchInfos.TotalFetched, "maxResults", batchInfos.MaxResults, "totalResults", batchInfos.TotalResults)
		Articles.mu.Lock()
		*Articles.articles = append(*Articles.articles, *newArticles...)
		Articles.mu.Unlock()
		return nil
	}

	err := c.processArticles(endpoint, query, maxResults, addArticle)
	if err != nil {
		return nil, err
	}

	// Trim results to maxResults if necessary.
	if len(*Articles.articles) > maxResults {
		*Articles.articles = (*Articles.articles)[:maxResults]
	}

	return Articles.articles, nil
}

// ProcessBreakingNews fetches breaking news Articles from the API and processes them using the provided action function as a go routine.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) ProcessBreakingNews(query BreakingNewsQuery, maxResults int, action func(*[]Article, BatchInfos) error) error {
	if err := query.Validate(); err != nil {
		return err
	}
	return c.processArticles("/latest", &query, maxResults, action)
}

// Get the latest news Articles in real-time from various sources worldwide.
// Filter by categories, countries, languages and more.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetBreakingNews(query BreakingNewsQuery, maxResults int) (*[]Article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/latest", &query, maxResults)
}

// ProcessBreakingNews fetches breaking news Articles from the API and processes them using the provided action function as a go routine.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) ProcessCreakingNews(query CryptoNewsQuery, maxResults int, action func(*[]Article, BatchInfos) error) error {
	if err := query.Validate(); err != nil {
		return err
	}
	return c.processArticles("/crypto", &query, maxResults, action)
}

// Get cryptocurrency-related news with additional filters like coin symbols, sentiment analysis, and specialized crypto tags.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetCryptoNews(query CryptoNewsQuery, maxResults int) (*[]Article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/crypto", &query, maxResults)
}

// Search through news archives with date range filters while maintaining all filtering capabilities of breaking news.
//
// maxResults is the maximum number of Articles to fetch. If set to 0, no limit is applied.
func (c *NewsdataClient) GetHistoricalNews(query HistoricalNewsQuery, maxResults int) (*[]Article, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return c.fetchArticles("/archive", &query, maxResults)
}

// ProcessHistoricalNews fetches historical news Articles from the API and processes them using the provided action function as a go routine.
//
// maxResults is the maximum number of Articles to process. If set to 0, no limit is applied.
func (c *NewsdataClient) ProcessHistoricalNews(query HistoricalNewsQuery, maxResults int, action func(*[]Article, BatchInfos) error) error {
	if err := query.Validate(); err != nil {
		return err
	}
	return c.processArticles("/archive", &query, maxResults, action)
}

// fetchSources fetches news sources from the API.
func (c *NewsdataClient) fetchSources(endpoint string, query *SourcesQuery) (*[]Source, error) {
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
func (c *NewsdataClient) GetSources(query SourcesQuery) (*[]Source, error) {
	if err := query.Validate(); err != nil {
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
