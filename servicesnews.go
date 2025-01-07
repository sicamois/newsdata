package newsdata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NewsService handles operations related to news articles from the NewsData API.
//
// It provides methods to fetch latest news, news archives, and crypto news.
type NewsService struct {
	client   *NewsDataClient
	endpoint endpoint
}

func (c *NewsDataClient) newLatestNewsService() *NewsService {
	return &NewsService{
		client:   c,
		endpoint: endpointLatestNews,
	}
}

func (c *NewsDataClient) newNewsArchiveService() *NewsService {
	return &NewsService{
		client:   c,
		endpoint: endpointNewsArchive,
	}
}

func (c *NewsDataClient) newCryptoNewsService() *NewsService {
	return &NewsService{
		client:   c,
		endpoint: endpointCoinNews,
	}
}

// DateTime is a wrapper around time.Time that implements custom JSON unmarshaling
// to handle the date format used by the NewsData API.
type DateTime struct {
	time.Time
}

// SentimentStats represents the sentiment analysis statistics for a news article.
// Each field represents the probability score for that sentiment category.
type SentimentStats struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}

// Tags is a wrapper around []string for handling article tags.
// It implements custom JSON unmarshaling to handle API restriction messages
// and null values gracefully.
type Tags []string

// NewsArticle represents a single news article from the NewsData API.
//
// See https://newsdata.io/documentation/#http_response for more details.
type NewsArticle struct {
	Id             string         `json:"Article_id"`      // Unique identifier for the article
	Title          string         `json:"title"`           // Article headline
	Link           string         `json:"link"`            // URL to the original article
	Keywords       []string       `json:"keywords"`        // Keywords associated with the article
	Creator        []string       `json:"creator"`         // Authors of the article
	VideoURL       string         `json:"video_url"`       // URL to associated video content
	Description    string         `json:"description"`     // Brief summary of the article
	Content        string         `json:"content"`         // Full article content
	PubDate        DateTime       `json:"pubDate"`         // Publication date and time
	PubDateTZ      string         `json:"pubDateTZ"`       // Timezone of publication date
	ImageURL       string         `json:"image_url"`       // URL to article's main image
	SourceId       string         `json:"source_id"`       // Unique identifier of the news source
	SourcePriority int            `json:"source_priority"` // Priority ranking of the source
	SourceName     string         `json:"source_name"`     // Name of the news source
	SourceURL      string         `json:"source_url"`      // URL of the news source
	SourceIconURL  string         `json:"source_icon"`     // URL to source's icon
	Language       string         `json:"language"`        // Article's language code
	Countries      []string       `json:"country"`         // Countries associated with the article
	Categories     []string       `json:"category"`        // Article categories
	AiTags         Tags           `json:"ai_tag"`          // AI-generated topic tags
	Sentiment      string         `json:"sentiment"`       // Overall sentiment classification
	SentimentStats SentimentStats `json:"sentiment_stats"` // Detailed sentiment analysis scores
	AiRegions      Tags           `json:"ai_region"`       // AI-detected geographical regions
	Coin           []string       `json:"coin"`            // Cryptocurrency coins mentioned
	Duplicate      bool           `json:"duplicate"`       // Whether article is a duplicate
}

// newsResponse represents the news API response.
//
// See https://newsdata.io/documentation/#http_response
type newsResponse struct {
	Status       string        `json:"status"`       // Response status ("success" or error message)
	TotalResults int           `json:"totalResults"` // Total number of NewsArticles matching the query
	Articles     []NewsArticle `json:"results"`      // Array of NewsArticles
	NextPage     string        `json:"nextPage"`     // Next page token
}

func (s *NewsService) fetch(ctx context.Context, params requestParams) (*newsResponse, error) {
	body, err := s.client.fetch(ctx, s.endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("fetchNews - error fetching news - error: %w", err)
	}
	// Decode the JSON response.
	var data newsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("fetchNews - error unmarshalling news response - error: %w", err)
	}
	return &data, nil
}

// Stream returns a channel that streams news articles matching the given query and parameters.
//
// It handles pagination automatically and continues streaming until all matching articles
// are retrieved or the context is cancelled. Errors are sent on the error channel.
func (s *NewsService) Stream(ctx context.Context, query string, params ...NewsRequestParams) (<-chan *NewsArticle, <-chan error) {
	out := make(chan *NewsArticle)
	errChan := make(chan error, 1)

	go func() {
		start := time.Now()
		defer close(out)
		defer close(errChan)
		articlesCount := 0
		reqParams := newRequestParams(query, s.client.logger, s.endpoint, params...)
		defer func() {
			// Closure are evaluated when the function is executed, not when defer is defined. Hence, articlesCount & duration will have the correct value.
			s.client.logger.Debug("newsdata: Stream - done", "service", s.endpoint.String(), "last_params", reqParams, "articlesCount", articlesCount, "duration", time.Since(start))
		}()
		for {
			res, err := s.fetch(ctx, reqParams)
			if err != nil {
				errChan <- fmt.Errorf("newsdata: Stream: %w", err)
				return
			}
			for _, article := range res.Articles {
				select {
				case out <- &article:
					articlesCount++
				case <-ctx.Done():
					errChan <- fmt.Errorf("newsdata: Stream - context done: %w", ctx.Err())
					return
				}
			}
			if articlesCount == res.TotalResults {
				return
			}
			if res.NextPage != "" {
				reqParams["page"] = res.NextPage
			} else {
				return
			}
		}
	}()
	return out, errChan
}

// Get retrieves a specified number of news articles matching the given query and parameters.
//
// It returns at most maxResults articles. If maxResults is 0, it returns all matching articles.
func (s *NewsService) Get(ctx context.Context, query string, maxResults int, params ...NewsRequestParams) ([]*NewsArticle, error) {
	var articles []*NewsArticle
	defer func() {
		s.client.logger.Debug("newsdata: Get - done", "service", s.endpoint.String(), "articlesCount", len(articles))
	}()
	if maxResults > 0 {
		articles = make([]*NewsArticle, 0, maxResults)
	} else {
		articles = make([]*NewsArticle, 0)
	}
	// maxResultsCause := errors.New("maxResults reached")
	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	articlesChan, errChan := s.Stream(newCtx, query, params...)
	for article := range articlesChan {
		articles = append(articles, article)
		if maxResults > 0 && len(articles) == maxResults {
			cancel()
			return articles[:maxResults], nil
		}
	}
	// Check if there was an error in the stream
	for {
		select {
		case err := <-errChan:
			return nil, err
		default:
			return articles, nil
		}
	}
}
