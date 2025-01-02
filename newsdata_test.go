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

func TestGetArticles(t *testing.T) {
	client := NewClient(APIKey(t))
	req := client.NewArticleRequest(LatestNews, "artificial intelligence").
		WithLanguages("en").
		WithCategories("technology").
		WithFieldsExcluded("Title").
		WithRemoveDuplicates()

	Articles, err := client.GetArticles(req, 88)
	if err != nil {
		t.Fatalf("Error fetching Breaking News: %v", err)
	}
	if len(Articles) == 0 || len(Articles) != 88 {
		t.Fatalf("Invalid number of Articles: %d - should 88", len(Articles))
	}
	for _, Article := range Articles {
		if Article.Title != "" {
			t.Fatalf("Article title field is not exluded")
		}
	}
}

func TestGetArticlesFromArchive(t *testing.T) {
	client := NewClient(APIKey(t))
	req := client.NewArticleRequest(NewsArchive, "artificial intelligence").
		WithFromDate(time.Date(2024, 12, 01, 0, 0, 0, 0, time.UTC)).
		WithToDate(time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC))
	Articles, err := client.GetArticles(req, 100)
	if err != nil {
		t.Fatalf("Error fetching History News: %v", err)
	}
	if len(Articles) == 0 || len(Articles) != 100 {
		t.Fatalf("Invalid number of Articles: %d - should 100", len(Articles))
	}
}

func TestGetSources(t *testing.T) {
	client := NewClient(APIKey(t))
	req := client.NewSourceRequest().
		WithCategory("technology").
		WithLanguage("fr")
	sources, err := client.GetSources(req)
	if err != nil {
		t.Fatalf("Error fetching Sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatalf("No sources found")
	}
}
