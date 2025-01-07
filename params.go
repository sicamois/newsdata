package newsdata

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
)

// endpoint represents an API endpoint in the NewsData API.
type endpoint string

// Predefined API endpoints
const (
	endpointLatestNews  endpoint = "latest"  // Endpoint for fetching latest news
	endpointNewsArchive endpoint = "archive" // Endpoint for accessing news archives
	endpointCoinNews    endpoint = "crypto"  // Endpoint for cryptocurrency news
	endpointSources     endpoint = "sources" // Endpoint for news sources information
)

// String returns a human-readable description of the endpoint.
func (e endpoint) String() string {
	switch e {
	case endpointLatestNews:
		return "Latest News"
	case endpointNewsArchive:
		return "News Archive"
	case endpointCoinNews:
		return "Crypto News"
	case endpointSources:
		return "Sources"
	}
	return "Unknown"
}

// requestParams represents a map of query parameters for API requests.
type requestParams map[string]string

// newRequestParams creates a new set of request parameters with the given query and options.
// It validates and processes the parameters based on the endpoint type.
func newRequestParams[T NewsRequestParams | SourceRequestParams](query string, logger *slog.Logger, endpoint endpoint, params ...T) requestParams {
	p := requestParams{}
	if query != "" {
		if endpoint != endpointSources {
			p["q"] = query
		} else {
			logger.Warn("newsdata: query is not supported for sources")
		}
	}
	for _, param := range params {
		param(p, endpoint, logger)
	}
	return p
}

type NewsRequestParams func(p requestParams, endpoint endpoint, logger *slog.Logger)

// WithQueryInTitle adds a query to search in article titles.
//
// QueryInTitle can't be used with Query or QueryInMeta parameter in the same query.
func WithQueryInTitle(query string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if p["qInMeta"] != "" || p["q"] != "" {
			logger.Error("newsdata: QueryInTitle can't be used with Query or QueryInMeta. Only QueryInTitle will be used.")
			delete(p, "qInMeta")
			delete(p, "q")
		}
		if len(query) > 512 {
			logger.Warn("newsdata: query length is greater than 512, truncating to 512")
			query = query[:512]
		}
		p["qInTitle"] = query
	}
}

// WithQueryInMetadata adds a query to search in article metadata.
//
// QueryInMetadata can't be used with Query or QueryInTitle parameter in the same query.
func WithQueryInMetadata(query string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if p["qInTitle"] != "" || p["q"] != "" {
			logger.Error("newsdata: QueryInMetadata can't be used with Query or QueryInTitle. Only QueryInMetadata will be used.")
			delete(p, "qInTitle")
			delete(p, "q")
		}
		if len(query) > 512 {
			logger.Warn("newsdata: query length is greater than 512, truncating to 512")
			query = query[:512]
		}
		p["qInMeta"] = query
	}
}

// validateCategories validates and filters the provided category list.
// It ensures only allowed categories are included and limits the total to 5.
func validateCategories(categories []string, logger *slog.Logger) []string {
	safeCategories := make([]string, 0, len(categories))
	for _, category := range categories {
		if slices.Contains(allowedCategories, category) {
			safeCategories = append(safeCategories, category)
		} else {
			logger.Warn(fmt.Sprintf("newsdata: category \"%s\" is not allowed", category))
		}
	}
	if len(safeCategories) > 5 {
		logger.Warn("newsdata: categories length is greater than 5, truncating to 5")
		categories = categories[:5]
	}
	return safeCategories
}

// WithCategories adds category filters to the article request, maximum 5 categories.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed categories.
//
// You can use either the 'categories' parameter to include specific categories or the 'excludecategories' parameter to exclude them, but not both simultaneously.
func WithCategories(categories ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(categories) == 0 {
			return
		}
		if p["excludecategory"] != "" {
			logger.Error("newsdata: categories and excluded categories cannot be used together")
			return
		}
		safeCategories := validateCategories(categories, logger)
		if safeCategories != nil {
			p["category"] = strings.Join(safeCategories, ",")
		}
	}
}

// WithCategoriesExlucded adds category exclusion filters to the article request, maximum 5 categories.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed categories.
//
// You can use either the 'category' parameter to include specific categories or the 'excludecategory' parameter to exclude them, but not both simultaneously.
func WithCategoriesExlucded(categories ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(categories) == 0 {
			return
		}
		if p["category"] != "" {
			logger.Error("newsdata: categories and excluded categories cannot be used together")
			return
		}
		safeCategories := validateCategories(categories, logger)
		if safeCategories != nil {
			p["excludecategory"] = strings.Join(safeCategories, ",")
		}
	}
}

// validateCountries validates and filters the provided country codes.
// It ensures only allowed country codes are included and limits the total to 5.
func validateCountries(countries []string, logger *slog.Logger) []string {
	safeCountries := make([]string, 0, len(countries))
	for _, country := range countries {
		if slices.Contains(allowedCountries, country) {
			safeCountries = append(safeCountries, country)
		} else {
			logger.Warn(fmt.Sprintf("newsdata: country \"%s\" is not allowed", country))
		}
	}
	if len(safeCountries) > 5 {
		logger.Warn("newsdata: countries length is greater than 5, truncating to 5")
		countries = countries[:5]
	}
	return safeCountries
}

// WithCountries adds country filters to the article request.
//
// It accepts up to 5 country codes and validates them against allowed values.
func WithCountries(countries ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(countries) == 0 {
			return
		}

		safeCountries := validateCountries(countries, logger)
		if safeCountries != nil {
			p["country"] = strings.Join(safeCountries, ",")
		}
	}
}

// validateLanguages filters and validates the provided language codes.
func validateLanguages(languages []string, logger *slog.Logger) []string {
	safeLanguages := make([]string, 0, len(languages))
	for _, language := range languages {
		if slices.Contains(allowedLanguages, language) {
			safeLanguages = append(safeLanguages, language)
		} else {
			logger.Warn(fmt.Sprintf("newsdata: language \"%s\" is not allowed", language))
		}
	}
	if len(safeLanguages) > 5 {
		logger.Warn("newsdata: languages length is greater than 5, truncating to 5")
		languages = languages[:5]
	}
	return safeLanguages
}

// WithLanguages adds language filters to the article request, maximum 5 languages.
//
// Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed languages.
func WithLanguages(languages ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(languages) == 0 {
			return
		}
		safeLanguages := validateLanguages(languages, logger)
		if safeLanguages != nil {
			p["language"] = strings.Join(safeLanguages, ",")
		}
	}
}

// WithDomains adds domain filters to the article request, maximum 5 domains.
//
// Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func WithDomains(domains ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(domains) == 0 {
			return
		}
		if len(domains) > 5 {
			logger.Warn("newsdata: domains length is greater than 5, truncating to 5")
			domains = domains[:5]
		}
		p["domain"] = strings.Join(domains, ",")
	}
}

// WithDomainExcluded adds domain exclusion filters to the article request, maximum 5 domains.
//
// Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func WithDomainExcluded(domains ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(domains) == 0 {
			return
		}
		if len(domains) > 5 {
			logger.Warn("newsdata: domains length is greater than 5, truncating to 5")
			domains = domains[:5]
		}
		p["excludedomain"] = strings.Join(domains, ",")
	}
}

// validatePriorityDomain validates if the provided domain is an allowed priority domain.
// Returns true if the domain is valid, false otherwise.
func validatePriorityDomain(priorityDomain string, logger *slog.Logger) bool {
	if !slices.Contains(allowedPriorityDomains, priorityDomain) {
		logger.Warn(fmt.Sprintf("newsdata: priority domain \"%s\" is not allowed", priorityDomain))
		return false
	}
	return true
}

// WithDomainUrls adds domain URL filters to the article request.
//
// It accepts up to 5 domain URLs for filtering news sources.
func WithDomainUrls(domainUrls ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(domainUrls) == 0 {
			return
		}
		if len(domainUrls) > 5 {
			logger.Warn("newsdata: domain URLs length is greater than 5, truncating to 5")
			domainUrls = domainUrls[:5]
		}
		p["domainurl"] = strings.Join(domainUrls, ",")
	}
}

// WithSourcePriorityDomain sets a priority domain for the article request
func WithSourcePriorityDomain(priorityDomain string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if priorityDomain == "" {
			return
		}
		if !validatePriorityDomain(priorityDomain, logger) {
			return
		}
		p["prioritydomain"] = priorityDomain
	}
}

// WithFieldsExcluded specifies fields to exclude from the response.
func WithFieldsExcluded(fields ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(fields) == 0 {
			return
		}
		p["excludefield"] = strings.Join(fields, ",")
	}
}

// WithTimezone Search the news articles for a specific timezone.
//
// Please refer to [timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for the list of allowed timezones.
func WithTimezone(timezone string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if timezone == "" {
			return
		}
		p["timezone"] = timezone
	}
}

// WithOnlyFullContent requests only articles with a full content.
func WithOnlyFullContent() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["full_content"] = "1"
	}
}

// WithNoFullContent requests only articles without a full content.
func WithNoFullContent() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["full_content"] = "0"
	}
}

// WithOnlyImage requests only articles with an image.
func WithOnlyImage() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["image"] = "1"
	}
}

// WithNoImage requests only articles without image.
func WithNoImage() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["image"] = "0"
	}
}

// WithOnlyVideo requests only articles with a video.
func WithOnlyVideo() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["video"] = "1"
	}
}

// WithNoVideo requests only articles without video.
func WithNoVideo() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["video"] = "0"
	}
}

// WithFromDate sets the start date for the article search.
//
// The date is formatted as YYYY-MM-DD in the request.
func WithFromDate(date time.Time) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["from_date"] = date.Format("2006-01-02")
	}
}

// WithToDate sets the end date for the article search.
//
// The date is formatted as YYYY-MM-DD in the request.
func WithToDate(date time.Time) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		p["to_date"] = date.Format("2006-01-02")
	}
}

// WithTimeframe sets a time window for the article search.
//
// The timeframe can be specified in hours and minutes, up to 48 hours.
func WithTimeframe(hours int, minutes int) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if hours+minutes == 0 || hours < 0 || minutes < 0 {
			logger.Error("newsdata: timeframe arguments must be greater than 0")
			return
		}
		switch endpoint {
		case endpointLatestNews:
			if minutes == 0 {
				if hours > 48 {
					logger.Error("newsdata: timeframe must be between 0h and 48h")
					return
				}
				p["timeframe"] = fmt.Sprintf("%d", hours)
			} else {
				totalMinutes := hours*60 + minutes
				if totalMinutes > 2880 {
					logger.Error("newsdata: timeframe must be between 0h and 48h")
					return
				}
				p["timeframe"] = fmt.Sprintf("%dm", totalMinutes)
			}
		}
	}
}

// WithSentiment adds sentiment analysis filter to the article request.
//
// It validates the sentiment value against allowed options.
func WithSentiment(sentiment string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if endpoint == endpointNewsArchive {
			logger.Warn(fmt.Sprintf("newsdata: sentiment is not supported for %s", endpoint.String()))
			return
		}
		if sentiment == "" {
			return
		}
		if !slices.Contains(allowedSentiments, sentiment) {
			logger.Warn(fmt.Sprintf("newsdata: sentiment \"%s\" is not allowed", sentiment))
			return
		}
		p["sentiment"] = sentiment
	}
}

// validateTags validates and filters the provided tags.
//
// It ensures only allowed tags are included.
func validateTags(tags []string, logger *slog.Logger) []string {
	safeTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		if slices.Contains(allowedTags, tag) {
			safeTags = append(safeTags, tag)
		} else {
			logger.Warn(fmt.Sprintf("newsdata: tag \"%s\" is not allowed", tag))
		}
	}
	return safeTags
}

// WithTags adds tag filters to the article request.
//
// It accepts multiple tags and validates them against allowed values.
func WithTags(tags ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(tags) == 0 {
			return
		}
		if endpoint == endpointNewsArchive {
			logger.Warn(fmt.Sprintf("newsdata: tags are not supported for %s", endpoint.String()))
			return
		}
		safeTags := validateTags(tags, logger)
		if safeTags != nil {
			p["tag"] = strings.Join(safeTags, ",")
		}
	}
}

// WithRemoveDuplicates enables duplicate article filtering in the response.
// This option is not supported for news archive requests.
func WithRemoveDuplicates() NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if endpoint == endpointNewsArchive {
			logger.Warn(fmt.Sprintf("newsdata: remove duplicates is not supported for %s", endpoint.String()))
			return
		}
		p["removeduplicate"] = "1"
	}
}

// WithCoins adds cryptocurrency coin filters to the article request.
//
// It accepts up to 5 coin symbols (like btc, eth, usdt, bnb, etc.) for filtering.
func WithCoins(coins ...string) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if len(coins) == 0 {
			return
		}
		if len(coins) > 5 {
			logger.Warn("newsdata: coins length is greater than 5, truncating to 5")
			coins = coins[:5]
		}
		p["coin"] = strings.Join(coins, ",")
	}
}

// WithSize sets the number of articles to return per page.
//
// The value must be between 1 and 50.
func WithSize(size int) NewsRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if size < 1 || size > 50 {
			logger.Error("newsdata: size must be between 1 and 50")
			return
		}
		p["size"] = fmt.Sprintf("%d", size)
	}
}

// SourceRequestParams is a function type for configuring source request parameters.
type SourceRequestParams func(p requestParams, endpoint endpoint, logger *slog.Logger)

// WithCountry adds a country filter to the source request.
// It validates the country code against allowed values.
func WithCountry(country string) SourceRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if country == "" {
			return
		}
		for _, cnt := range allowedCountries {
			if cnt == country {
				p["country"] = country
				return
			}
		}
		logger.Warn(fmt.Sprintf("newsdata: country \"%s\" is not allowed", country))
	}
}

// WithCategory adds category filter to the source request
func WithCategory(category string) SourceRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if category == "" {
			return
		}
		for _, ctg := range allowedCategories {
			if ctg == category {
				p["category"] = category
				return
			}
		}
		logger.Warn(fmt.Sprintf("newsdata: category \"%s\" is not allowed", category))
	}
}

// WithLanguage adds language filter to the source request
func WithLanguage(language string) SourceRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if language == "" {
			return
		}
		for _, lang := range allowedLanguages {
			if lang == language {
				p["language"] = language
				return
			}
		}
		logger.Warn(fmt.Sprintf("newsdata: language \"%s\" is not allowed", language))
	}
}

// WithPriorityDomain sets a priority domain for the source request
func WithPriorityDomain(priorityDomain string) SourceRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if priorityDomain == "" {
			return
		}
		if !validatePriorityDomain(priorityDomain, logger) {
			return
		}
		p["prioritydomain"] = priorityDomain
	}
}

// WithDomainUrl sets a domain URL filter for the source request
func WithDomainUrl(domainUrl string) SourceRequestParams {
	return func(p requestParams, endpoint endpoint, logger *slog.Logger) {
		if domainUrl == "" {
			return
		}
		p["domainurl"] = domainUrl
	}
}
