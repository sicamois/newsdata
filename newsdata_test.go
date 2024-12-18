package newsdata

import (
	"bufio"
	"os"
	"strings"
	"testing"
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
	client := NewClient(APIKey(t), 1)
	query := BreakingNewsQuery{
		Query:     "artificial intelligence",
		Languages: []string{"en"},
		Categories: []string{
			"technology",
		},
	}
	articles, err := client.GetBreakingNews(query)
	if err != nil {
		t.Fatalf("Error fetching latest news: %v", err)
	}
	if len(*articles) == 0 || len(*articles) > 1 {
		t.Fatalf("Invalid number of articles: %d - should 1", len(*articles))
	}
}

func TestGetSources(t *testing.T) {
	client := NewClient(APIKey(t), 0)
	options := SourcesQuery{
		Country: "us",
	}
	sources, err := client.GetSources(options)
	if err != nil {
		t.Fatalf("Error fetching sources: %v", err)
	}
	if len(*sources) == 0 {
		t.Fatalf("No sources found")
	}
}
