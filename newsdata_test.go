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
	Articles, err := client.GetBreakingNews(query, 1)
	if err != nil {
		t.Fatalf("Error fetching Breaking News: %v", err)
	}
	if len(*Articles) == 0 || len(*Articles) > 1 {
		t.Fatalf("Invalid number of Articles: %d - should 1", len(*Articles))
	}
	for _, Article := range *Articles {
		if Article.Title != "" {
			t.Fatalf("Article title field is not exluded")
		}
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
