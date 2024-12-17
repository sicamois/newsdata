package newsdata

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//
// LATEST NEWS SERVICE
//

// latestNewsService handles news-related endpoints.
type latestNewsService struct {
	client   *baseClient
	endpoint string
}

// NewsQueryParams represents the query parameters for the news endpoint.
type NewsQueryParams struct {
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

func (p *NewsQueryParams) setPage(page string) {
	p.Page = page
}

// NewsQueryOptions represents the options for advanced search.
type NewsQueryOptions struct {
	QueryInTitle      string   // Search term in article title
	QueryInMetadata   string   // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe         string   // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Categories        []string // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string // List of categories to exclude
	Countries         []string // List of country codes (e.g., ["us", "uk"])
	Languages         []string // List of language codes (e.g., ["en", "es"])
}

// Get fetches news based on query parameters.
func (s *latestNewsService) Get(params *NewsQueryParams) (*newsResponse, error) {
	return s.client.doRequest(s.endpoint, params)
}

// AdvancedSearch fetches news based on a query and some options to filter the results.
func (s *latestNewsService) AdvancedSearch(query string, options NewsQueryOptions) (*[]article, error) {
	params := NewsQueryParams{
		Query:             query,
		QueryInTitle:      options.QueryInTitle,
		QueryInMetadata:   options.QueryInMetadata,
		Timeframe:         options.Timeframe,
		Categories:        options.Categories,
		ExcludeCategories: options.ExcludeCategories,
		Countries:         options.Countries,
		Languages:         options.Languages,
	}
	// Validate the query parameters.
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return s.client.getArticles(s.endpoint, &params, s.client.maxResults)
}

// Search fetches news based on a simple query.
func (s *latestNewsService) Search(query string) (*[]article, error) {
	return s.AdvancedSearch(query, NewsQueryOptions{})
}

// Validate validates the NewsQueryParams struct, ensuring all fields are valid.
func (p NewsQueryParams) Validate() error {
	if p.QueryInTitle != "" && p.QueryInMetadata != "" {
		return fmt.Errorf("QueryInTitle and QueryInMetadata cannot be used together")
	}
	if len(p.Categories) > 0 && len(p.ExcludeCategories) > 0 {
		return fmt.Errorf("Categories and ExcludeCategories cannot be used together")
	}
	if len(p.Query) > 512 {
		return fmt.Errorf("Query cannot be longer than 512 characters")
	}
	if len(p.QueryInTitle) > 512 {
		return fmt.Errorf("QueryInTitle cannot be longer than 512 characters")
	}
	if len(p.QueryInMetadata) > 512 {
		return fmt.Errorf("QueryInMetadata cannot be longer than 512 characters")
	}
	if p.Timeframe != "" {
		hours, err := strconv.Atoi(p.Timeframe)
		if err != nil {
			minValue, _ := strings.CutSuffix(p.Timeframe, "m")
			minutes, err := strconv.Atoi(minValue)
			if err != nil {
				return fmt.Errorf("invalid Timeframe: %s", p.Timeframe)
			}
			if minutes < 0 || minutes > 2880 {
				return fmt.Errorf("Timeframe must be between 0 and 2880 minutes")
			}
		}
		if hours < 0 || hours > 48 {
			return fmt.Errorf("Timeframe must be between 0 and 48 hours")
		}
	}
	if len(p.Countries) > 5 {
		return fmt.Errorf("Countries cannot be longer than 5 countries")
	}
	for _, countryCode := range p.Countries {
		if !isValidCountry(countryCode) {
			return fmt.Errorf("invalid country code: %s", countryCode)
		}
	}
	if len(p.Categories) > 5 {
		return fmt.Errorf("Categories cannot be longer than 5 categories")
	}
	for _, category := range p.Categories {
		if !isValidCategory(category) {
			return fmt.Errorf("invalid category in Categories: %s", category)
		}
	}
	if len(p.ExcludeCategories) > 5 {
		return fmt.Errorf("ExcludeCategories cannot be longer than 5 categories")
	}
	for _, category := range p.ExcludeCategories {
		if !isValidCategory(category) {
			return fmt.Errorf("invalid category in ExcludeCategories: %s", category)
		}
	}
	if len(p.Languages) > 5 {
		return fmt.Errorf("Languages cannot be longer than 5 languages")
	}
	for _, languageCode := range p.Languages {
		if !isValidLanguage(languageCode) {
			return fmt.Errorf("invalid language code: %s", languageCode)
		}
	}
	if len(p.Domains) > 5 {
		return fmt.Errorf("Domains cannot be longer than 5 domains")
	}
	if len(p.DomainUrls) > 5 {
		return fmt.Errorf("DomainUrls cannot be longer than 5 domain URLs")
	}
	if len(p.ExcludeDomains) > 5 {
		return fmt.Errorf("ExcludeDomains cannot be longer than 5 domains")
	}
	for _, field := range p.ExcludeFields {
		if !isValidField(field) {
			return fmt.Errorf("invalid field in ExcludeFields: %s", field)
		}
	}
	if p.PriorityDomain != "" && !isValidPriorityDomain(p.PriorityDomain) {
		return fmt.Errorf("%s is not an available priority domain. Possible options are: %v", p.PriorityDomain, strings.Join(allowedPriorityDomain, ","))
	}
	if p.Size < 0 || p.Size > 50 {
		return fmt.Errorf("Size must be between 1 and 50")
	}
	return nil
}

//
// CRYPTO NEWS SERVICE
//

// cryptoNewsService handles crypto news-related endpoints.
type cryptoNewsService struct {
	client   *baseClient
	endpoint string
}

// CryptoQueryParams represents the query parameters for the crypto news endpoint.
type CryptoQueryParams struct {
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

func (p *CryptoQueryParams) setPage(page string) {
	p.Page = page
}

// CryptoQueryOptions represents the options for advanced search.
type CryptoQueryOptions struct {
	QueryInTitle    string   // Search term in article title
	QueryInMetadata string   // Search term in article metadata (titles, URL, meta keywords and meta description)
	Timeframe       string   // Timeframe to filter by hours are represented by a integer value, minutes are represented by an integer value with a suffix of m
	Languages       []string // List of language codes (e.g., ["en", "es"])
	Tags            []string // List of tags (e.g., ["blockchain", "liquidity", "scam"])
	Sentiment       string   // List of sentiment : "positive", "negative" or "neutral"]
}

// Get fetches crypto news based on query parameters.
func (s *cryptoNewsService) Get(params CryptoQueryParams) (*newsResponse, error) {
	return s.client.doRequest(s.endpoint, &params)
}

// AdvancedSearch fetches crypto news based on a query and some options to filter the results.
func (s *cryptoNewsService) AdvancedSearch(query string, options CryptoQueryOptions) (*[]article, error) {
	params := CryptoQueryParams{
		Query:           query,
		QueryInTitle:    options.QueryInTitle,
		QueryInMetadata: options.QueryInMetadata,
		Timeframe:       options.Timeframe,
		Languages:       options.Languages,
		Tags:            options.Tags,
		Sentiment:       options.Sentiment,
	}
	// Validate the query parameters.
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return s.client.getArticles(s.endpoint, &params, s.client.maxResults)
}

// Search fetches crypto news based on a simple query.
func (s *cryptoNewsService) Search(query string) (*[]article, error) {
	return s.AdvancedSearch(query, CryptoQueryOptions{})
}

// Validate validates the CryptoQueryParams struct, ensuring all fields are valid.
func (p CryptoQueryParams) Validate() error {
	if p.QueryInTitle != "" && p.QueryInMetadata != "" {
		return fmt.Errorf("QueryInTitle and QueryInMetadata cannot be used together")
	}
	if len(p.Query) > 512 {
		return fmt.Errorf("Query cannot be longer than 512 characters")
	}
	if len(p.QueryInTitle) > 512 {
		return fmt.Errorf("QueryInTitle cannot be longer than 512 characters")
	}
	if len(p.QueryInMetadata) > 512 {
		return fmt.Errorf("QueryInMetadata cannot be longer than 512 characters")
	}
	if p.Timeframe != "" {
		hours, err := strconv.Atoi(p.Timeframe)
		if err != nil {
			minValue, _ := strings.CutSuffix(p.Timeframe, "m")
			minutes, err := strconv.Atoi(minValue)
			if err != nil {
				return fmt.Errorf("invalid Timeframe: %s", p.Timeframe)
			}
			if minutes < 0 || minutes > 2880 {
				return fmt.Errorf("Timeframe must be between 0 and 2880 minutes")
			}
		}
		if hours < 0 || hours > 48 {
			return fmt.Errorf("Timeframe must be between 0 and 48 hours")
		}
	}
	if len(p.Tags) > 5 {
		return fmt.Errorf("Countries cannot be longer than 5 countries")
	}
	for _, tag := range p.Tags {
		if !isValidTag(tag) {
			return fmt.Errorf("invalid tag: %s", tag)
		}
	}
	if len(p.Sentiment) > 0 && !isValidSentiment(p.Sentiment) {
		return fmt.Errorf("invalid sentiment: %s", p.Sentiment)
	}
	if len(p.Languages) > 5 {
		return fmt.Errorf("Languages cannot be longer than 5 languages")
	}
	for _, languageCode := range p.Languages {
		if !isValidLanguage(languageCode) {
			return fmt.Errorf("invalid language code: %s", languageCode)
		}
	}
	if len(p.Domains) > 5 {
		return fmt.Errorf("Domains cannot be longer than 5 domains")
	}
	if len(p.DomainUrls) > 5 {
		return fmt.Errorf("DomainUrls cannot be longer than 5 domain URLs")
	}
	if len(p.ExcludeDomains) > 5 {
		return fmt.Errorf("ExcludeDomains cannot be longer than 5 domains")
	}
	for _, field := range p.ExcludeFields {
		if !isValidField(field) {
			return fmt.Errorf("invalid field in ExcludeFields: %s", field)
		}
	}
	if p.PriorityDomain != "" && !isValidPriorityDomain(p.PriorityDomain) {
		return fmt.Errorf("%s is not an available priority domain. Possible options are: %v", p.PriorityDomain, strings.Join(allowedPriorityDomain, ","))
	}
	if p.Size < 0 || p.Size > 50 {
		return fmt.Errorf("Size must be between 1 and 50")
	}
	if p.From.IsZero() && p.From.After(time.Now()) {
		return fmt.Errorf("From date must be in the past")
	}
	if p.To.IsZero() && p.To.After(time.Now()) {
		return fmt.Errorf("To date must be in the past")
	}
	return nil
}

//
// NEWS ARCHIVE SERVICE
//

// newsArchiveService handles news archive-related endpoints.
type newsArchiveService struct {
	client   *baseClient
	endpoint string
}

// ArchiveQueryParams represents the query parameters for the news archive endpoint.
type ArchiveQueryParams struct {
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

func (p *ArchiveQueryParams) setPage(page string) {
	p.Page = page
}

// ArchiveQueryOptions represents the options for advanced search.
type ArchiveQueryOptions struct {
	From              DateTime // From date
	To                DateTime // To date
	QueryInTitle      string   // Search term in article title
	QueryInMetadata   string   // Search term in article metadata (titles, URL, meta keywords and meta description)
	Categories        []string // List of categories (e.g., ["technology", "sports"])
	ExcludeCategories []string // List of categories to exclude
	Countries         []string // List of country codes (e.g., ["us", "uk"])
	Languages         []string // List of language codes (e.g., ["en", "es"])
}

// Get fetches news archive based on query parameters.
func (s *newsArchiveService) Get(params *ArchiveQueryParams) (*newsResponse, error) {
	return s.client.doRequest(s.endpoint, params)
}

// AdvancedSearch fetches news archive based on a query and some options to filter the results.
func (s *newsArchiveService) AdvancedSearch(query string, from time.Time, to time.Time, options ArchiveQueryOptions) (*[]article, error) {
	params := ArchiveQueryParams{
		Query: query,
		From: DateTime{
			Time: from,
		},
		To: DateTime{
			Time: to,
		},
		QueryInTitle:      options.QueryInTitle,
		QueryInMetadata:   options.QueryInMetadata,
		Categories:        options.Categories,
		ExcludeCategories: options.ExcludeCategories,
		Countries:         options.Countries,
		Languages:         options.Languages,
	}
	// Validate the query parameters.
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return s.client.getArticles(s.endpoint, &params, s.client.maxResults)
}

// Search fetches news archive based on a simple query.
func (s *newsArchiveService) Search(query string, from time.Time, to time.Time) (*[]article, error) {
	return s.AdvancedSearch(query, from, to, ArchiveQueryOptions{})
}

// Validate validates the ArchiveQueryParams struct, ensuring all fields are valid.
func (p ArchiveQueryParams) Validate() error {
	if p.QueryInTitle != "" && p.QueryInMetadata != "" {
		return fmt.Errorf("QueryInTitle and QueryInMetadata cannot be used together")
	}
	if len(p.Categories) > 0 && len(p.ExcludeCategories) > 0 {
		return fmt.Errorf("Categories and ExcludeCategories cannot be used together")
	}
	if len(p.Query) > 512 {
		return fmt.Errorf("Query cannot be longer than 512 characters")
	}
	if len(p.QueryInTitle) > 512 {
		return fmt.Errorf("QueryInTitle cannot be longer than 512 characters")
	}
	if len(p.QueryInMetadata) > 512 {
		return fmt.Errorf("QueryInMetadata cannot be longer than 512 characters")
	}
	if len(p.Countries) > 5 {
		return fmt.Errorf("Countries cannot be longer than 5 countries")
	}
	for _, countryCode := range p.Countries {
		if !isValidCountry(countryCode) {
			return fmt.Errorf("invalid country code: %s", countryCode)
		}
	}
	if len(p.Categories) > 5 {
		return fmt.Errorf("Categories cannot be longer than 5 categories")
	}
	for _, category := range p.Categories {
		if !isValidCategory(category) {
			return fmt.Errorf("invalid category in Categories: %s", category)
		}
	}
	if len(p.ExcludeCategories) > 5 {
		return fmt.Errorf("ExcludeCategories cannot be longer than 5 categories")
	}
	for _, category := range p.ExcludeCategories {
		if !isValidCategory(category) {
			return fmt.Errorf("invalid category in ExcludeCategories: %s", category)
		}
	}
	if len(p.Languages) > 5 {
		return fmt.Errorf("Languages cannot be longer than 5 languages")
	}
	for _, languageCode := range p.Languages {
		if !isValidLanguage(languageCode) {
			return fmt.Errorf("invalid language code: %s", languageCode)
		}
	}
	if len(p.Domains) > 5 {
		return fmt.Errorf("Domains cannot be longer than 5 domains")
	}
	if len(p.DomainUrls) > 5 {
		return fmt.Errorf("DomainUrls cannot be longer than 5 domain URLs")
	}
	if len(p.ExcludeDomains) > 5 {
		return fmt.Errorf("ExcludeDomains cannot be longer than 5 domains")
	}
	for _, field := range p.ExcludeFields {
		if !isValidField(field) {
			return fmt.Errorf("invalid field in ExcludeFields: %s", field)
		}
	}
	if p.PriorityDomain != "" && !isValidPriorityDomain(p.PriorityDomain) {
		return fmt.Errorf("%s is not an available priority domain. Possible options are: %v", p.PriorityDomain, strings.Join(allowedPriorityDomain, ","))
	}
	if p.Size < 0 || p.Size > 50 {
		return fmt.Errorf("Size must be between 1 and 50")
	}
	if p.From.IsZero() && p.From.After(time.Now()) {
		return fmt.Errorf("From date must be in the past")
	}
	if p.To.IsZero() && p.To.After(time.Now()) {
		return fmt.Errorf("To date must be in the past")
	}
	return nil
}
