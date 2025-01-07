package newsdata

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func Loadenv(t *testing.T) {
	file, err := os.OpenFile(".env", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Error opening .env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, "=")
		os.Setenv(values[0], values[1])
	}
}

func TestGetLatestNews(t *testing.T) {
	Loadenv(t)
	client := NewClient()
	articles, err := client.LatestNews.Get(context.Background(), "artificial intelligence", 88, WithLanguages("en"), WithCategories("technology"), WithFieldsExcluded("Title"), WithRemoveDuplicates())
	if err != nil {
		t.Fatalf("Error fetching Breaking News: %v", err)
	}
	if len(articles) != 88 {
		t.Fatalf("Invalid number of Articles: %d - should 88", len(articles))
	}
	for _, Article := range articles {
		if Article.Title != "" {
			t.Fatalf("Article title field is not exluded")
		}
	}
}

func TestCancelGetLatestNews(t *testing.T) {
	Loadenv(t)
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	_, err := client.LatestNews.Get(ctx, "", 500, WithCountries("us"))
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Error fetching Latest News: %v", err)
		}
	}
}

func TestGetNewsFromArchive(t *testing.T) {
	Loadenv(t)
	client := NewClient()
	articles, err := client.NewsArchive.Get(context.Background(), "", 72, WithCountries("us"))
	if err != nil {
		t.Fatalf("Error fetching Latest News: %v", err)
	}
	if len(articles) != 72 {
		t.Fatalf("Invalid number of Articles: %d - should 72", len(articles))
	}
}

func TestGetSources(t *testing.T) {
	Loadenv(t)
	client := NewClient(WithLogLevel(slog.LevelWarn))
	sources, err := client.Sources.Get(context.Background())
	if err != nil {
		t.Fatalf("Error fetching Sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatalf("No sources found")
	}
}
