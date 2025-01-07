package newsdata

import (
	"context"
	"encoding/json"
	"fmt"
)

// SourcesService handles operations related to news sources from the NewsData API.
// It provides methods to retrieve information about available news sources.
type SourcesService struct {
	client *NewsdataClient
}

func (c *NewsdataClient) newSourcesService() *SourcesService {
	return &SourcesService{
		client: c,
	}
}

// Source represents a news source from the NewsData API.
// It contains metadata about the news provider including its identity,
// content categories, and geographical coverage.
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
// The method supports filtering by country and other criteria through SourceParams.
func (s *SourcesService) Get(ctx context.Context, params ...SourceParams) ([]*Source, error) {
	sources := make([]*Source, 0, 100)
	reqParams := newRequestParams("", s.client.logger, endpointSources, params...)
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
