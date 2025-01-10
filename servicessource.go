package newsdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// SourcesService handles operations related to news sources from the NewsData API.
// It provides methods to retrieve information about available news sources.
type SourcesService struct {
	client *NewsDataClient
}

func (c *NewsDataClient) newSourcesService() *SourcesService {
	return &SourcesService{
		client: c,
	}
}

// Source represents a news source from the NewsData API.
//
// See https://newsdata.io/documentation/#news-sources for more details.
type Source struct {
	Id          string   `json:"id"`          // Unique identifier for the source
	Name        string   `json:"name"`        // Display name of the source
	Url         string   `json:"url"`         // Website URL of the source
	IconUrl     string   `json:"icon"`        // URL to source's icon/logo
	Priority    int      `json:"priority"`    // Source priority ranking
	Description string   `json:"description"` // Brief description of the source
	Categories  []string `json:"category"`    // Content categories covered
	Languages   []string `json:"language"`    // Languages supported
	Countries   []string `json:"country"`     // Countries covered
	LastFetch   DateTime `json:"last_fetch"`  // Timestamp of last content fetch
}

// sourcesResponse represents the news sources API response.
//
// See https://newsdata.io/documentation/#news-sources
type sourcesResponse struct {
	Status       string   `json:"status"`       // Response status ("success" or error message)
	TotalResults int      `json:"totalResults"` // Total number of news sources matching the query
	Sources      []Source `json:"results"`      // Array of news sources
}

// Get retrieves a list of news sources matching the provided parameters.
// It returns all available sources if no parameters are specified.
//
// The method supports filtering by country and other criteria through SourceRequestParams.
func (s *SourcesService) Get(ctx context.Context, params ...SourceRequestParams) ([]*Source, error) {
	start := time.Now()
	sources := make([]*Source, 0, 100)
	reqParams := newRequestParams("", s.client.logger, endpointSources, params...)

	s.client.logger.Debug("retrieving sources started", "service", endpointSources.String(), "params", reqParams.String())
	defer func() {
		// Closure are evaluated when the function is executed, not when defer is defined. Hence, articlesCount & duration will have the correct value.
		s.client.logger.Debug("retrieving sources ended", "service", endpointSources.String(), "params", reqParams.String(), "sourcesCount", len(sources), "duration", time.Since(start))
	}()

	body, err := s.client.fetch(ctx, endpointSources, reqParams)
	if err != nil {
		return nil, fmt.Errorf("newsdata: getSources - error fetching sources - error: %w", err)
	}

	// Decode the JSON response.
	var res sourcesResponse
	if err := json.Unmarshal(body, &res); err != nil { // Parse []byte to go struct pointer
		return nil, fmt.Errorf("newsdata: getSources - error unmarshalling sources response - error: %w", err)
	}
	resSources := res.Sources
	for i := 0; i < len(resSources); i++ {
		sources = append(sources, &resSources[i])
	}

	return sources, nil
}
