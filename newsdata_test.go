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

func TestClientInitialization(t *testing.T) {
	client := NewClient("xxx")
	if client.LatestNews.endpoint != "/latest" {
		t.Fatalf("LatestNews endpoint is not /latest")
	}
	if client.CryptoNews.endpoint != "/crypto" {
		t.Fatalf("CryptoNews endpoint is not /crypto")
	}
	if client.NewsArchive.endpoint != "/archive" {
		t.Fatalf("NewsArchive endpoint is not /archive")
	}
	if client.Sources.endpoint != "/sources" {
		t.Fatalf("Sources endpoint is not /sources")
	}
}

func TestSearchLatestNews(t *testing.T) {
	client := NewClient(APIKey(t))
	client.LimitResultsToFirst(1)

	articles, err := client.LatestNews.Search("")
	if err != nil {
		t.Fatalf("Error fetching latest news: %v", err)
	}
	if len(*articles) == 0 {
		t.Fatalf("No articles found")
	}
	if len(*articles) > 1 {
		t.Fatalf("More than one article found")
	}
}

func TestAdvancedSearchLatestNews(t *testing.T) {
	client := NewClient(APIKey(t))
	client.LimitResultsToFirst(1)
	options := NewsQueryOptions{
		Languages:  []string{"en"},
		Categories: []string{"technology"},
	}
	articles, err := client.LatestNews.AdvancedSearch("artificial intelligence", options)
	if err != nil {
		t.Fatalf("Error fetching latest news: %v", err)
	}
	if len(*articles) == 0 {
		t.Fatalf("No articles found")
	}
	if len(*articles) > 1 {
		t.Fatalf("More than one article found")
	}
}
