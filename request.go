package newsdata

import "time"

// pagerValider is an interface for setting the page parameter and validating the query.
type pagerValider interface {
	setPage(string)
	validate() error
}

// Tags is is a wrapper around []string for coin-specific tags, AI tags & AI Regions, used to handle the case where the API returns a restriction message (typically "ONLY AVAILABLE IN PROFESSIONAL AND CORPORATE PLANS")
type Tags []string

type BaseRequest struct {
	Id                []string `query:"id"`                                                   // List of Article IDs
	Query             string   `query:"q" validate:"maxlen:512"`                              // Main search term
	QueryInTitle      string   `query:"qInTitle" validate:"maxlen:512,mutex:QueryInMetadata"` // Search term in Article title
	QueryInMetadata   string   `query:"qInMeta" validate:"maxlen:512,mutex:QueryInTitle"`     // Search term in Article metadata (titles, URL, meta keywords and meta description)
	Categories        []string `query:"category" validate:"maxlen:5,in:categories"`           // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string `query:"excludecategory" validate:"maxlen:5,in:categories"`    // List of categories to exclude
	Countries         []string `query:"country" validate:"maxlen:5,in:countries"`             // List of country codes (e.g., ["us", "uk"])
	Languages         []string `query:"language" validate:"maxlen:5,in:languages"`            // List of language codes (e.g., ["en", "es"])    // Search term in Article metadata (titles, URL, meta keywords and meta description)
	Domains           []string `query:"domain" validate:"maxlen:5"`                           // List of domains (e.g., ["nytimes", "bbc"])
	DomainUrls        []string `query:"domainurl" validate:"maxlen:5"`                        // List of domain URLs (e.g., ["nytimes.com", "bbc.com", "bbc.co.uk"])
	ExcludeDomains    []string `query:"excludedomain" validate:"maxlen:5"`                    // List of domains to exclude
	ExcludeFields     []string `query:"excludefield" validate:"custom"`                       // List of fields to exclude
	PriorityDomain    string   `query:"prioritydomain" validate:"in:priorityDomains"`         // Search the news Articles only from top news domains. Possible values : Top, Medium, Low
	Timezone          string   `query:"timezone"`                                             // Search the news Articles for a specific timezone. Example values : "America/New_york", "Asia/Kolkata" → see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	FullContent       string   `query:"full_content" validate:"in:binaries"`                  // If set to 1, only the Articles with full_content response object will be returned, if set to 0, only the Articles without full_content response object will be returned
	Image             string   `query:"image" validate:"in:binaries"`                         // If set to 1, only the Articles with featured image will be returned, if set to 0, only the Articles without featured image will be returned
	Video             string   `query:"video" validate:"in:binaries"`                         // If set to 1, only the Articles with video will be returned, if set to 0, only the Articles without video will be returned
	Size              int      `query:"size" validate:"min:1,max:50"`                         // Number of results per page
	Page              string   `query:"page"`                                                 // Page ref
}

// BreakingNewsRequest represents the query parameters for the breaking news endpoint.
//
// See https://newsdata.io/documentation/#latest-news
type BreakingNewsRequest struct {
	BaseRequest
	Timeframe        string   `query:"timeframe"`                              // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Tags             []string `query:"tag" validate:"maxlen:5,in:tags"`        // List of AI tags
	Sentiment        string   `query:"sentiment" validate:"in:sentiments"`     // Filter by sentiment ("positive", "negative", "neutral")
	Regions          []string `query:"region" validate:"maxlen:5"`             // List of regions
	RemoveDuplicates string   `query:"removeduplicate" validate:"in:binaries"` // If set to true, duplicate Articles will be removed from the results
}

// setPage sets the page parameter
func (q *BreakingNewsRequest) setPage(page string) {
	q.Page = page
}

// Validate validates the BreakingNewsRequest struct, ensuring all fields are valid.
func (query *BreakingNewsRequest) validate() error {
	return validate(query)
}

// HistoricalNewsRequest represents the query parameters for the historical news endpoint.
//
// See https://newsdata.io/documentation/#news-archive
type HistoricalNewsRequest struct {
	BaseRequest
	From time.Time `query:"from_date" validate:"time:past"` // From date
	To   time.Time `query:"to_date" validate:"time:past"`   // To date
}

// setPage sets the page parameter
func (q *HistoricalNewsRequest) setPage(page string) {
	q.Page = page
}

// Validate validates the HistoricalNewsRequest struct, ensuring all fields are valid.
func (query *HistoricalNewsRequest) validate() error {
	return validate(query)
}

// CryptoNewsRequest represents the query parameters for the crypto news endpoint.
//
// See https://newsdata.io/documentation/#crypto-news
type CryptoNewsRequest struct {
	BaseRequest
	Coins            []string  `query:"coin"`                                   // List of coins
	RemoveDuplicates string    `query:"removeduplicate" validate:"in:binaries"` // If set to true, duplicate Articles will be removed from the results
	Sentiment        string    `query:"sentiment" validate:"in:sentiments"`     // Filter by sentiment ("positive", "negative", "neutral")
	Tags             []string  `query:"tag" validate:"in:tags"`                 // Filter by crypto-specific tags
	From             time.Time `query:"from_date" validate:"time:past"`         // From date
	To               time.Time `query:"to_date" validate:"time:past"`           // To date
}

// setPage sets the page parameter
func (q *CryptoNewsRequest) setPage(page string) {
	q.Page = page
}

// Validate validates the CryptoNewsRequest struct, ensuring all fields are valid.
func (query *CryptoNewsRequest) validate() error {
	return validate(query)
}

// SourcesRequest represents the query parameters for the sources endpoint.
//
// See https://newsdata.io/documentation/#news-sources
type SourcesRequest struct {
	Country        string `query:"country" validate:"maxlen:5,in:countries"`     // Filter by country code
	Language       string `query:"language" validate:"maxlen:5,in:languages"`    // Filter by language code
	Category       string `query:"category" validate:"maxlen:5,in:categories"`   // Filter by category (e.g., "technology")
	PriorityDomain string `query:"prioritydomain" validate:"in:priorityDomains"` // Filter by priority domain (possible values: "top", "medium", "low")
	DomainUrl      string `query:"domainurl" validate:"maxlen:5"`                // Filter by domain URL (e.g., "nytimes.com")
}

// Validate validates the HistoricalNewsRequest struct, ensuring all fields are valid.
func (query *SourcesRequest) validate() error {
	return validate(query)
}
