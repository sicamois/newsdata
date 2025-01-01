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

func (c *NewsdataClient) NewArticleRequest(service newsService, query string) articleRequest {
	req := articleRequest{
		service: service,
		context: context.Background(),
		params:  make(map[string]string),
		logger:  c.Logger,
	}
	if len(query) > 512 {
		req.logger.Warn("newsdata: query length is greater than 512, truncating to 512")
		query = query[:512]
	}
	req.params["q"] = query
	return req
}

func (req articleRequest) WithContext(context context.Context) articleRequest {
	req.context = context
	return req
}

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

func (req articleRequest) WithCountries(countries ...string) articleRequest {
	if len(countries) == 0 {
		return req
	}

	safeCountries := validateCountries(countries, req.logger)
	req.params["country"] = strings.Join(safeCountries, ",")
	return req
}

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

func (req articleRequest) WithLanguages(languages ...string) articleRequest {
	if len(languages) == 0 {
		return req
	}
	safeLanguages := validateLanguages(languages, req.logger)
	req.params["language"] = strings.Join(safeLanguages, ",")
	return req
}

func (req articleRequest) WithDomains(domains ...string) articleRequest {
	if len(domains) == 0 {
		return req
	}
	req.params["domain"] = strings.Join(domains, ",")
	return req
}

func (req articleRequest) WithDomainExcluded(domains ...string) articleRequest {
	if len(domains) == 0 {
		return req
	}
	req.params["excludedomain"] = strings.Join(domains, ",")
	return req
}

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

func (req articleRequest) WithDomainUrls(domainUrls ...string) articleRequest {
	if len(domainUrls) == 0 {
		return req
	}
	req.params["domainurl"] = strings.Join(domainUrls, ",")
	return req
}

func (req articleRequest) WithFieldsExcluded(fields ...string) articleRequest {
	if len(fields) == 0 {
		return req
	}
	req.params["excludefield"] = strings.Join(fields, ",")
	return req
}

func (req articleRequest) WithTimezone(timezone string) articleRequest {
	if timezone == "" {
		return req
	}
	req.params["timezone"] = timezone
	return req
}

func (req articleRequest) WithOnlyFullContent() articleRequest {
	req.params["full_content"] = "1"
	return req
}

func (req articleRequest) WithNoFullContent() articleRequest {
	req.params["full_content"] = "0"
	return req
}

func (req articleRequest) WithOnlyImage() articleRequest {
	req.params["image"] = "1"
	return req
}

func (req articleRequest) WithNoImage() articleRequest {
	req.params["image"] = "0"
	return req
}

func (req articleRequest) WithOnlyVideo() articleRequest {
	req.params["video"] = "1"
	return req
}

func (req articleRequest) WithNoVideo() articleRequest {
	req.params["video"] = "0"
	return req
}

func (req articleRequest) WithFromDate(date time.Time) articleRequest {
	req.params["from_date"] = date.Format("2006-01-02")
	return req
}

func (req articleRequest) WithToDate(date time.Time) articleRequest {
	req.params["to_date"] = date.Format("2006-01-02")
	return req
}

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

func (req articleRequest) WithRemoveDuplicates() articleRequest {
	if req.service == NewsArchive {
		req.logger.Warn(fmt.Sprintf("newsdata: remove duplicates is not supported for %s", req.service.String()))
		return req
	}
	req.params["removeduplicate"] = "1"
	return req
}

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

func (req articleRequest) WithSize(size int) articleRequest {
	if size < 1 || size > 50 {
		req.logger.Error("newsdata: size must be between 1 and 50")
		return req
	}
	req.params["size"] = fmt.Sprintf("%d", size)
	return req
}

func (req articleRequest) WithPage(page string) articleRequest {
	if page == "" {
		return req
	}
	req.params["page"] = page
	return req
}
