package newsdata

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"time"
)

func APIKey(t *testing.T) string {
	file, err := os.OpenFile(".env", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Error opening .env file: %v", err)
	}
	defer file.Close()

	apiKey := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "API_KEY=") {
			apiKey = strings.TrimPrefix(line, "API_KEY=")
		}
	}
	if apiKey == "" {
		t.Fatalf("API_KEY not found in .env file")
	}

	return apiKey
}

func TestGetBreakingNews(t *testing.T) {
	client := NewClient(APIKey(t))
	query := BreakingNewsQuery{
		Query:     "artificial intelligence",
		Languages: []string{"en"},
		Categories: []string{
			"technology",
		},
		ExcludeFields:    []string{"Title"},
		RemoveDuplicates: "1",
	}
	Articles, err := client.GetBreakingNews(query, 88)
	if err != nil {
		t.Fatalf("Error fetching Breaking News: %v", err)
	}
	if len(*Articles) == 0 || len(*Articles) != 88 {
		t.Fatalf("Invalid number of Articles: %d - should 88", len(*Articles))
	}
	for _, Article := range *Articles {
		if Article.Title != "" {
			t.Fatalf("Article title field is not exluded")
		}
	}
}

func TestGenerateBreakingNews(t *testing.T) {
	client := NewClient(APIKey(t))
	query := BreakingNewsQuery{
		Query:     "artificial intelligence",
		Languages: []string{"en"},
		Categories: []string{
			"technology",
		},
		ExcludeFields:    []string{"Title"},
		RemoveDuplicates: "1",
	}
	out, errChan := client.GenerateBreakingNews(query, 88)
	Articles := []Article{}
	var generationError error

	go func() {
		for err := range errChan {
			generationError = err
		}
	}()

	for article := range out {
		Articles = append(Articles, article)
	}

	if generationError != nil {
		t.Fatalf("Error fetching Breaking News: %v", generationError)
	}

	if len(Articles) == 0 || len(Articles) != 88 {
		t.Fatalf("Invalid number of Articles: %d - should 88", len(Articles))
	}
}

func TestGetHistoricalNews(t *testing.T) {
	client := NewClient(APIKey(t))
	// client.SetTimeout(1 * time.Second)
	query := HistoricalNewsQuery{
		Query: "artificial intelligence",
		From:  DateTime{Time: time.Date(2024, 12, 01, 0, 0, 0, 0, time.UTC)},
		To:    DateTime{Time: time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC)},
	}
	Articles, err := client.GetHistoricalNews(query, 100)
	if err != nil {
		t.Fatalf("Error fetching History News: %v", err)
	}
	if len(*Articles) == 0 || len(*Articles) != 100 {
		t.Fatalf("Invalid number of Articles: %d - should 100", len(*Articles))
	}
}

func TestGetSources(t *testing.T) {
	client := NewClient(APIKey(t))
	options := SourcesQuery{
		Country: "us",
	}
	sources, err := client.GetSources(options)
	if err != nil {
		t.Fatalf("Error fetching Sources: %v", err)
	}
	if len(*sources) == 0 {
		t.Fatalf("No sources found")
	}
}
