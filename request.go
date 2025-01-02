package newsdata

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
)

// NewsService represents the type of news service to use.
type NewsService int

// NewsService represents the type of news service to use.
const (
	LatestNews NewsService = iota
	CryptoNews
	NewsArchive
)

// String returns the human-readable name of the news service
func (service NewsService) String() string {
	switch service {
	case LatestNews:
		return "Latest News"
	case CryptoNews:
		return "Crypto News"
	case NewsArchive:
		return "News Archive"
	}
	return ""
}

// endpoint returns the API endpoint path for the news service
func (service NewsService) endpoint() string {
	switch service {
	case LatestNews:
		return "/latest"
	case CryptoNews:
		return "/crypto"
	case NewsArchive:
		return "/archive"
	}
	return ""
}

// ArticleRequest represents a request for news articles.
type ArticleRequest struct {
	service NewsService
	context context.Context
	params  map[string]string
	logger  *slog.Logger
}

// NewArticleRequest creates a new article request with the specified service and query.
//
// The query is used to search for articles in the specified service. Service can be LatestNews, CryptoNews or NewsArchive.
func (c *NewsdataClient) NewArticleRequest(service NewsService, query string) ArticleRequest {
	req := ArticleRequest{
		service: service,
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.logger,
	}
	if len(query) > 512 {
		req.logger.Warn("newsdata: query length is greater than 512, truncating to 512")
		query = query[:512]
	}
	if query == "" {
		return req
	}
	req.params["q"] = query
	return req
}

// NewArticleRequestById creates a new article request to fetch articles by their IDs.
//
// Service can be LatestNews, CryptoNews or NewsArchive.
func (c *NewsdataClient) NewArticleRequestById(service NewsService, ids ...string) ArticleRequest {
	req := ArticleRequest{
		service: service,
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.logger,
	}
	if len(ids) == 0 {
		req.logger.Error("newsdata: ids cannot be empty")
		return req
	}
	if len(ids) > 50 {
		req.logger.Warn("newsdata: ids length is greater than 50, truncating to 50")
		ids = ids[:50]
	}
	req.params["id"] = strings.Join(ids, ",")
	return req
}

// WithContext sets the context for the article request.
func (req ArticleRequest) WithContext(context context.Context) ArticleRequest {
	req.context = context
	return req
}

// WithQueryInTitle adds a query to search in article titles.
func (req ArticleRequest) WithQueryInTitle(query string) ArticleRequest {
	if req.params["qInMeta"] != "" {
		req.logger.Error("newsdata: query in title and metadata cannot be used together")
		return req
	}
	if len(query) > 512 {
		req.logger.Warn("newsdata: query length is greater than 512, truncating to 512")
		query = query[:512]
	}
	req.params["qInTitle"] = query
	return req
}

// WithQueryInMetadata adds a query to search in article metadata.
func (req ArticleRequest) WithQueryInMetadata(query string) ArticleRequest {
	if req.params["qInTitle"] != "" {
		req.logger.Error("newsdata: query in title and metadata cannot be used together")
		return req
	}
	if len(query) > 512 {
		req.logger.Warn("newsdata: query length is greater than 512, truncating to 512")
		query = query[:512]
	}
	req.params["qInMeta"] = query
	return req
}

// validateCategories filters the provided categories.
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
func (req ArticleRequest) WithCategories(categories ...string) ArticleRequest {
	if len(categories) == 0 {
		return req
	}
	if req.params["excludecategory"] != "" {
		req.logger.Error("newsdata: categories and excluded categories cannot be used together")
		return req
	}
	safeCategories := validateCategories(categories, req.logger)
	if safeCategories != nil {
		req.params["category"] = strings.Join(safeCategories, ",")
	}
	return req
}

// WithCategoriesExlucded adds category exclusion filters to the article request, maximum 5 categories.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed categories.
//
// You can use either the 'category' parameter to include specific categories or the 'excludecategory' parameter to exclude them, but not both simultaneously.
func (req ArticleRequest) WithCategoriesExlucded(categories ...string) ArticleRequest {
	if len(categories) == 0 {
		return req
	}
	if req.params["category"] != "" {
		req.logger.Error("newsdata: categories and excluded categories cannot be used together")
		return req
	}
	safeCategories := validateCategories(categories, req.logger)
	req.params["excludecategory"] = strings.Join(safeCategories, ",")
	return req
}

// validateCountries filters and validates the provided country codes.
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

// WithCountries adds country filters to the article request, maximum 5 countries.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed countries.
func (req ArticleRequest) WithCountries(countries ...string) ArticleRequest {
	if len(countries) == 0 {
		return req
	}

	safeCountries := validateCountries(countries, req.logger)
	req.params["country"] = strings.Join(safeCountries, ",")
	return req
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

// WithLanguages adds language filters to the article request, maximum 5 languages.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed languages.
func (req ArticleRequest) WithLanguages(languages ...string) ArticleRequest {
	if len(languages) == 0 {
		return req
	}
	safeLanguages := validateLanguages(languages, req.logger)
	req.params["language"] = strings.Join(safeLanguages, ",")
	return req
}

// WithDomains adds domain filters to the article request, maximum 5 domains.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func (req ArticleRequest) WithDomains(domains ...string) ArticleRequest {
	if len(domains) == 0 {
		return req
	}
	if len(domains) > 5 {
		req.logger.Warn("newsdata: domains length is greater than 5, truncating to 5")
		domains = domains[:5]
	}
	req.params["domain"] = strings.Join(domains, ",")
	return req
}

// WithDomainExcluded adds domain exclusion filters to the article request, maximum 5 domains.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func (req ArticleRequest) WithDomainExcluded(domains ...string) ArticleRequest {
	if len(domains) == 0 {
		return req
	}
	if len(domains) > 5 {
		req.logger.Warn("newsdata: domains length is greater than 5, truncating to 5")
		domains = domains[:5]
	}
	req.params["excludedomain"] = strings.Join(domains, ",")
	return req
}

// validatePriorityDomain validates if the provided domain is an allowed priority domain.
func validatePriorityDomain(priorityDomain string, logger *slog.Logger) bool {
	if !slices.Contains(allowedPriorityDomains, priorityDomain) {
		logger.Warn(fmt.Sprintf("newsdata: priority domain \"%s\" is not allowed", priorityDomain))
		return false
	}
	return true
}

// WithPriorityDomain sets a priority domain for the article request.
func (req ArticleRequest) WithPriorityDomain(priorityDomain string) ArticleRequest {
	if priorityDomain == "" {
		return req
	}
	if !slices.Contains(allowedPriorityDomains, priorityDomain) {
		req.logger.Warn(fmt.Sprintf("newsdata: priority domain \"%s\" is not allowed", priorityDomain))
		return req
	}
	req.params["prioritydomain"] = priorityDomain
	return req
}

// WithDomainUrls adds domain URL filters to the article request, maximum 5 domain URLs.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func (req ArticleRequest) WithDomainUrls(domainUrls ...string) ArticleRequest {
	if len(domainUrls) == 0 {
		return req
	}
	req.params["domainurl"] = strings.Join(domainUrls, ",")
	return req
}

// WithFieldsExcluded specifies fields to exclude from the response.
func (req ArticleRequest) WithFieldsExcluded(fields ...string) ArticleRequest {
	if len(fields) == 0 {
		return req
	}
	req.params["excludefield"] = strings.Join(fields, ",")
	return req
}

// WithTimezone Search the news articles for a specific timezone.  Please refer to [timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for the list of allowed timezones.
func (req ArticleRequest) WithTimezone(timezone string) ArticleRequest {
	if timezone == "" {
		return req
	}
	req.params["timezone"] = timezone
	return req
}

// WithOnlyFullContent requests only articles with a full content.
func (req ArticleRequest) WithOnlyFullContent() ArticleRequest {
	req.params["full_content"] = "1"
	return req
}

// WithNoFullContent requests only articles without a full content.
func (req ArticleRequest) WithNoFullContent() ArticleRequest {
	req.params["full_content"] = "0"
	return req
}

// WithOnlyImage requests only articles with an image.
func (req ArticleRequest) WithOnlyImage() ArticleRequest {
	req.params["image"] = "1"
	return req
}

// WithNoImage requests only articles without image.
func (req ArticleRequest) WithNoImage() ArticleRequest {
	req.params["image"] = "0"
	return req
}

// WithOnlyVideo requests only articles with a video.
func (req ArticleRequest) WithOnlyVideo() ArticleRequest {
	req.params["video"] = "1"
	return req
}

// WithNoVideo requests only articles without video.
func (req ArticleRequest) WithNoVideo() ArticleRequest {
	req.params["video"] = "0"
	return req
}

// WithFromDate sets the start date for the article search.
func (req ArticleRequest) WithFromDate(date time.Time) ArticleRequest {
	req.params["from_date"] = date.Format("2006-01-02")
	return req
}

// WithToDate sets the end date for the article search.
func (req ArticleRequest) WithToDate(date time.Time) ArticleRequest {
	req.params["to_date"] = date.Format("2006-01-02")
	return req
}

// WithTimeframe sets a time window for the article search.
func (req ArticleRequest) WithTimeframe(hours int, minutes int) ArticleRequest {
	if hours+minutes == 0 || hours < 0 || minutes < 0 {
		return req
	}
	switch req.service {
	case LatestNews:
		if minutes == 0 {
			if hours > 48 {
				req.logger.Error("newsdata: timeframe must be between 0h and 48h")
				return req
			}
			req.params["timeframe"] = fmt.Sprintf("%d", hours)
		} else {
			min := hours*60 + minutes
			if min > 2880 {
				req.logger.Error("newsdata: timeframe must be between 0h and 48h")
				return req
			}
			req.params["timeframe"] = fmt.Sprintf("%dm", min)
		}
	case CryptoNews, NewsArchive:
		req = req.WithFromDate(time.Now().Add(-(time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))))
		req = req.WithToDate(time.Now())
	}
	return req
}

// WithSentiment adds sentiment analysis filter to the article request.
func (req ArticleRequest) WithSentiment(sentiment string) ArticleRequest {
	if req.service == NewsArchive {
		req.logger.Warn(fmt.Sprintf("newsdata: sentiment is not supported for %s", req.service.String()))
		return req
	}
	if sentiment == "" {
		return req
	}
	if !slices.Contains(allowedSentiments, sentiment) {
		req.logger.Warn(fmt.Sprintf("newsdata: sentiment \"%s\" is not allowed", sentiment))
		return req
	}
	req.params["sentiment"] = sentiment
	return req
}

// validateTags filters and validates the provided tags.
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

// WithTags adds tag filters to the article request, maximum 5 tags.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed tags.
func (req ArticleRequest) WithTags(tags ...string) ArticleRequest {
	if req.service == NewsArchive {
		req.logger.Warn(fmt.Sprintf("newsdata: tags are not supported for %s", req.service.String()))
		return req
	}
	if len(tags) == 0 {
		return req
	}
	safeTags := validateTags(tags, req.logger)
	req.params["tag"] = strings.Join(safeTags, ",")
	return req
}

// WithRemoveDuplicates removes duplicate articles from the response.
func (req ArticleRequest) WithRemoveDuplicates() ArticleRequest {
	if req.service == NewsArchive {
		req.logger.Warn(fmt.Sprintf("newsdata: remove duplicates is not supported for %s", req.service.String()))
		return req
	}
	req.params["removeduplicate"] = "1"
	return req
}

// WithCoins adds cryptocurrency coin filters to the article request
func (req ArticleRequest) WithCoins(coins ...string) ArticleRequest {
	if req.service != CryptoNews {
		req.logger.Warn(fmt.Sprintf("newsdata: coins are not supported for %s", req.service.String()))
		return req
	}
	if len(coins) == 0 {
		return req
	}
	if len(coins) > 5 {
		req.logger.Warn("newsdata: coins length is greater than 5, truncating to 5")
		coins = coins[:5]
	}
	req.params["coin"] = strings.Join(coins, ",")
	return req
}

// WithSize sets the number of articles to return per page.
func (req ArticleRequest) WithSize(size int) ArticleRequest {
	if size < 1 || size > 50 {
		req.logger.Error("newsdata: size must be between 1 and 50")
		return req
	}
	req.params["size"] = fmt.Sprintf("%d", size)
	return req
}

// WithPage sets the page for paginated results.
func (req ArticleRequest) WithPage(page string) ArticleRequest {
	if page == "" {
		return req
	}
	req.params["page"] = page
	return req
}

type SourceRequest struct {
	context context.Context
	params  map[string]string
	logger  *slog.Logger
}

// NewSourceRequest creates a new request for news sources.
func (c *NewsdataClient) NewSourceRequest() SourceRequest {
	req := SourceRequest{
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.logger,
	}
	return req
}

// WithContext sets the context for the source request.
func (req SourceRequest) WithContext(ctx context.Context) SourceRequest {
	req.context = ctx
	return req
}

// WithCountries adds country filter to the source request.
func (req SourceRequest) WithCountry(country string) SourceRequest {
	if country == "" {
		return req
	}
	for _, country := range allowedCountries {
		if country == country {
			req.params["country"] = country
			return req
		}
	}
	req.logger.Warn(fmt.Sprintf("newsdata: country \"%s\" is not allowed", country))
	return req
}

// WithCategory adds category filter to the source request
func (req SourceRequest) WithCategory(category string) SourceRequest {
	if category == "" {
		return req
	}
	for _, category := range allowedCategories {
		if category == category {
			req.params["category"] = category
			return req
		}
	}
	req.logger.Warn(fmt.Sprintf("newsdata: category \"%s\" is not allowed", category))
	return req
}

// WithLanguage adds language filter to the source request
func (req SourceRequest) WithLanguage(language string) SourceRequest {
	if language == "" {
		return req
	}
	for _, language := range allowedLanguages {
		if language == language {
			req.params["language"] = language
			return req
		}
	}
	req.logger.Warn(fmt.Sprintf("newsdata: language \"%s\" is not allowed", language))
	return req
}

// WithPriorityDomain sets a priority domain for the source request
func (req SourceRequest) WithPriorityDomain(priorityDomain string) SourceRequest {
	if priorityDomain == "" {
		return req
	}
	if !validatePriorityDomain(priorityDomain, req.logger) {
		return req
	}
	req.params["prioritydomain"] = priorityDomain
	return req
}

// WithDomainUrl sets a domain URL filter for the source request
func (req SourceRequest) WithDomainUrl(domainUrl string) SourceRequest {
	if domainUrl == "" {
		return req
	}
	req.params["domainurl"] = domainUrl
	return req
}
