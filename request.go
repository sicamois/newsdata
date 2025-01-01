package newsdata

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type newsService int

const (
	LatestNews newsService = iota
	CryptoNews
	NewsArchive
)

// String returns the human-readable name of the news service
func (service newsService) String() string {
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

// Endpoint returns the API endpoint path for the news service
func (service newsService) Endpoint() string {
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

type articleRequest struct {
	service newsService
	context context.Context
	params  map[string]string
	logger  *slog.Logger
}

// NewArticleRequest creates a new article request with the specified service and query.
func (c *NewsdataClient) NewArticleRequest(service newsService, query string) articleRequest {
	req := articleRequest{
		service: service,
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.logger,
	}
	if len(query) > 512 {
		req.logger.Warn("newsdata: query length is greater than 512, truncating to 512")
		query = query[:512]
	}
	req.params["q"] = query
	return req
}

// NewArticleRequestById creates a new article request to fetch articles by their IDs.
func (c *NewsdataClient) NewArticleRequestById(service newsService, ids ...string) articleRequest {
	req := articleRequest{
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
func (req articleRequest) WithContext(context context.Context) articleRequest {
	req.context = context
	return req
}

// WithQueryInTitle adds a query to search in article titles.
func (req articleRequest) WithQueryInTitle(query string) articleRequest {
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
func (req articleRequest) WithQueryInMetadata(query string) articleRequest {
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
func (req articleRequest) WithCategories(categories ...string) articleRequest {
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
func (req articleRequest) WithCategoriesExlucded(categories ...string) articleRequest {
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
func (req articleRequest) WithCountries(countries ...string) articleRequest {
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
func (req articleRequest) WithLanguages(languages ...string) articleRequest {
	if len(languages) == 0 {
		return req
	}
	safeLanguages := validateLanguages(languages, req.logger)
	req.params["language"] = strings.Join(safeLanguages, ",")
	return req
}

// WithDomains adds domain filters to the article request, maximum 5 domains.  Please refer to [newsdata.io docs](https://newsdata.io/documentation/#latest-news) for the list of allowed domains.
func (req articleRequest) WithDomains(domains ...string) articleRequest {
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
func (req articleRequest) WithDomainExcluded(domains ...string) articleRequest {
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
func (req articleRequest) WithPriorityDomain(priorityDomain string) articleRequest {
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
func (req articleRequest) WithDomainUrls(domainUrls ...string) articleRequest {
	if len(domainUrls) == 0 {
		return req
	}
	req.params["domainurl"] = strings.Join(domainUrls, ",")
	return req
}

// WithFieldsExcluded specifies fields to exclude from the response.
func (req articleRequest) WithFieldsExcluded(fields ...string) articleRequest {
	if len(fields) == 0 {
		return req
	}
	req.params["excludefield"] = strings.Join(fields, ",")
	return req
}

// WithTimezone Search the news articles for a specific timezone.  Please refer to [timezones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for the list of allowed timezones.
func (req articleRequest) WithTimezone(timezone string) articleRequest {
	if timezone == "" {
		return req
	}
	req.params["timezone"] = timezone
	return req
}

// WithOnlyFullContent requests only articles with a full content.
func (req articleRequest) WithOnlyFullContent() articleRequest {
	req.params["full_content"] = "1"
	return req
}

// WithNoFullContent requests only articles without a full content.
func (req articleRequest) WithNoFullContent() articleRequest {
	req.params["full_content"] = "0"
	return req
}

// WithOnlyImage requests only articles with an image.
func (req articleRequest) WithOnlyImage() articleRequest {
	req.params["image"] = "1"
	return req
}

// WithNoImage requests only articles without image.
func (req articleRequest) WithNoImage() articleRequest {
	req.params["image"] = "0"
	return req
}

// WithOnlyVideo requests only articles with a video.
func (req articleRequest) WithOnlyVideo() articleRequest {
	req.params["video"] = "1"
	return req
}

// WithNoVideo requests only articles without video.
func (req articleRequest) WithNoVideo() articleRequest {
	req.params["video"] = "0"
	return req
}

// WithFromDate sets the start date for the article search.
func (req articleRequest) WithFromDate(date time.Time) articleRequest {
	req.params["from_date"] = date.Format("2006-01-02")
	return req
}

// WithToDate sets the end date for the article search.
func (req articleRequest) WithToDate(date time.Time) articleRequest {
	req.params["to_date"] = date.Format("2006-01-02")
	return req
}

// WithTimeframe sets a time window for the article search.
func (req articleRequest) WithTimeframe(hours int, minutes int) articleRequest {
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
func (req articleRequest) WithSentiment(sentiment string) articleRequest {
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
func (req articleRequest) WithTags(tags ...string) articleRequest {
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
func (req articleRequest) WithRemoveDuplicates() articleRequest {
	if req.service == NewsArchive {
		req.logger.Warn(fmt.Sprintf("newsdata: remove duplicates is not supported for %s", req.service.String()))
		return req
	}
	req.params["removeduplicate"] = "1"
	return req
}

// WithCoins adds cryptocurrency coin filters to the article request
func (req articleRequest) WithCoins(coins ...string) articleRequest {
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
func (req articleRequest) WithSize(size int) articleRequest {
	if size < 1 || size > 50 {
		req.logger.Error("newsdata: size must be between 1 and 50")
		return req
	}
	req.params["size"] = fmt.Sprintf("%d", size)
	return req
}

// WithPage sets the page for paginated results.
func (req articleRequest) WithPage(page string) articleRequest {
	if page == "" {
		return req
	}
	req.params["page"] = page
	return req
}

type sourceRequest struct {
	context context.Context
	params  map[string]string
	logger  *slog.Logger
}

// NewSourceRequest creates a new request for news sources.
func (c *NewsdataClient) NewSourceRequest() sourceRequest {
	req := sourceRequest{
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.logger,
	}
	return req
}

// WithContext sets the context for the source request.
func (req sourceRequest) WithContext(ctx context.Context) sourceRequest {
	req.context = ctx
	return req
}

// WithCountries adds country filter to the source request.
func (req sourceRequest) WithCountry(country string) sourceRequest {
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
func (req sourceRequest) WithCategory(category string) sourceRequest {
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
func (req sourceRequest) WithLanguage(language string) sourceRequest {
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
func (req sourceRequest) WithPriorityDomain(priorityDomain string) sourceRequest {
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
func (req sourceRequest) WithDomainUrl(domainUrl string) sourceRequest {
	if domainUrl == "" {
		return req
	}
	req.params["domainurl"] = domainUrl
	return req
}
