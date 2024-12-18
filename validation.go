package newsdata

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// isValidCategory checks if a category is in the allowed list.
func isValidCategory(category string) bool {
	for _, allowed := range allowedCategories {
		if category == allowed {
			return true
		}
	}
	return false
}

// isValidCountry checks if a country code is in the allowed list.
func isValidCountry(countryCode string) bool {
	for _, allowed := range allowedCountries {
		if countryCode == allowed {
			return true
		}
	}
	return false
}

// isValidLanguage checks if a language code is in the allowed list.
func isValidLanguage(languageCode string) bool {
	for _, allowed := range allowedLanguages {
		if languageCode == allowed {
			return true
		}
	}
	return false
}

// isValidField checks if a field exists in the article struct.
func isValidField(field string) bool {
	articleFields := make([]string, 0)
	t := reflect.TypeOf(article{})
	for i := 0; i < t.NumField(); i++ {
		articleFields = append(articleFields, t.Field(i).Name)
	}
	for _, allowed := range articleFields {
		if field == allowed {
			return true
		}
	}
	return false
}

// isValidPriorityDomain checks if a priority domain is in the allowed list.
func isValidPriorityDomain(priorityDomain string) bool {
	for _, allowed := range allowedPriorityDomain {
		if priorityDomain == allowed {
			return true
		}
	}
	return false
}

// isValidSentiment checks if a sentiment is in the allowed list.
func isValidSentiment(sentiment string) bool {
	for _, allowed := range allowedSentiment {
		if sentiment == allowed {
			return true
		}
	}
	return false
}

// isValidTag checks if a tag is in the allowed list.
func isValidTag(tag string) bool {
	for _, allowed := range allowedTags {
		if tag == allowed {
			return true
		}
	}
	return false
}

// Validate validates the BreakingNewsQuery struct, ensuring all fields are valid.
func (p *BreakingNewsQuery) Validate() error {
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

// Validate validates the CryptoNewsQuery struct, ensuring all fields are valid.
func (p *CryptoNewsQuery) Validate() error {
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

// Validate validates the HistoricalNewsQuery struct, ensuring all fields are valid.
func (p *HistoricalNewsQuery) Validate() error {
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

// Validate validates the HistoricalNewsQuery struct, ensuring all fields are valid.
func (p *SourcesQuery) Validate() error {
	if p.Country != "" && !isValidCountry(p.Country) {
		return fmt.Errorf("invalid country code: %s", p.Country)
	}
	if p.Language != "" && !isValidLanguage(p.Language) {
		return fmt.Errorf("invalid language code: %s", p.Language)
	}
	if p.Category != "" && !isValidCategory(p.Category) {
		return fmt.Errorf("invalid category: %s", p.Category)
	}
	if p.PriorityDomain != "" && !isValidPriorityDomain(p.PriorityDomain) {
		return fmt.Errorf("%s is not an available priority domain. Possible options are: %v", p.PriorityDomain, strings.Join(allowedPriorityDomain, ","))
	}
	return nil
}
