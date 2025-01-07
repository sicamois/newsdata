# NewsData.io Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/sicamois/newsdata.svg)](https://pkg.go.dev/github.com/sicamois/newsdata)
[![Go Report Card](https://goreportcard.com/badge/github.com/sicamois/newsdata)](https://goreportcard.com/report/github.com/sicamois/newsdata)

A fully-featured Go client library for the [NewsData.io](https://newsdata.io) API that provides easy access to news articles and sources with robust error handling and extensive configuration options.

## Features

- **Multiple News Services**
  - Latest News
  - News Archive
  - Cryptocurrency News
  - News Sources
- **Flexible Article Retrieval**
  - Stream-based processing for efficient handling
  - Pagination support
  - Customizable result limits
- **Rich Filtering Options**
  - Categories and languages
  - Countries and domains
  - Date ranges and timeframes
  - Sentiment analysis
  - AI-powered tags and regions
- **Advanced Configuration**
  - Configurable HTTP client
  - Structured logging with multiple levels
  - Customizable timeouts
  - Extensive error handling

## Installation

```bash
go get github.com/sicamois/newsdata
```

## Quick Start

You'll need an API key from [NewsData.io](https://newsdata.io/api/register) to use this client.

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "time"

    "github.com/sicamois/newsdata"
)

func main() {
    // Create a new client
    client := newsdata.NewClient(
        newsdata.WithAPIKey("your-api-key"),
        newsdata.WithLogLevel(slog.LevelDebug),
    )

    // Get latest news articles about AI
    articles, err := client.LatestNews.Get(
        context.Background(),
        "artificial intelligence",
        10, // limit to 10 articles
        newsdata.WithLanguages("en"),
        newsdata.WithCategories("technology"),
        newsdata.WithRemoveDuplicates(),
    )
    if err != nil {
        panic(err)
    }

    // Process the articles
    for _, article := range articles {
        fmt.Printf("Title: %s\n", article.Title)
        fmt.Printf("Source: %s\n", article.SourceName)
        fmt.Printf("Published: %s\n", article.PubDate.Time.Format(time.RFC3339))
        if article.Sentiment != "" {
            fmt.Printf("Sentiment: %s\n", article.Sentiment)
            fmt.Printf("Sentiment Stats: +%.2f/=%.2f/-%.2f\n",
                article.SentimentStats.Positive,
                article.SentimentStats.Neutral,
                article.SentimentStats.Negative,
            )
        }
        fmt.Println()
    }
}
```

## Streaming Articles

For efficient processing of large result sets:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/sicamois/newsdata"
)

func main() {
    // Create a new client with the API key from the environment variable NEWSDATA_API_KEY
    client := newsdata.NewClient()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    articleChan, errChan := client.LatestNews.Stream(
        ctx,
        "cryptocurrency",
        newsdata.WithCategories("business"),
        newsdata.WithLanguages("en"),
    )

    for {
        select {
        case article, ok := <-articleChan:
            if !ok {
                return // Channel closed, all articles processed
            }
            fmt.Printf("Title: %s\n", article.Title)
        case err := <-errChan:
            if err != nil {
                fmt.Printf("Error: %v\n", err)
                return
            }
            return
        case <-ctx.Done():
            fmt.Printf("Context done: %v\n", ctx.Err())
            return
        }
    }
}
```

## Advanced Configuration

### Client Options

```go
package main

import (
    "log/slog"
    "os"
    "time"

    "github.com/sicamois/newsdata"
)

func main() {
    // Create a log file
    logFile, err := os.Create("log_" + time.Now().Format("2006-01-02_15-04-05") + ".txt")
    if err != nil {
        panic(err)
    }
    defer logFile.Close()

    // Create a new client
    client := newsdata.NewClient(
        newsdata.WithTimeout(10*time.Second),
        newsdata.WithCustomLogWriter(logFile),
        newsdata.WithLogLevel(slog.LevelDebug),
    )

    // Use client...
}
```

### Article Filtering

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/sicamois/newsdata"
)

func main() {
    client := newsdata.NewClient(
        newsdata.WithAPIKey("your-api-key"),
    )

    ctx := context.Background()
    articles, err := client.LatestNews.Get(
        ctx,
        "climate change",
        20,
        newsdata.WithLanguages("en", "fr"),
        newsdata.WithCountries("us", "gb", "fr"),
        newsdata.WithCategories("environment", "science"),
        newsdata.WithFromDate(time.Now().AddDate(0, 0, -7)),
        newsdata.WithToDate(time.Now()),
        newsdata.WithSentiment("positive"),
        newsdata.WithRemoveDuplicates(),
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    // Process articles...
}
```

### Cryptocurrency News

```go
package main

import (
    "context"
    "fmt"

    "github.com/sicamois/newsdata"
)

    func main() {
    client := newsdata.NewClient()

    ctx := context.Background()
    articles, err := client.CryptoNews.Get(
        ctx,
        "",
        10,
        newsdata.WithCoins("BTC", "ETH"),
        newsdata.WithTags("blockchain", "technology"),
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    // Process articles...
}
```

### News Sources

```go
package main

import (
    "context"
    "fmt"

    "github.com/sicamois/newsdata"
)

func main() {
    client := newsdata.NewClient()

    ctx := context.Background()
    sources, err := client.Sources.Get(
        ctx,
        newsdata.WithCountry("us"),
        newsdata.WithCategory("technology"),
        newsdata.WithLanguage("en"),
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    // Process sources...
}
```

## Streaming News Articles to S3

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/sicamois/newsdata"
)

func main() {
    // Initialize NewsData client
    client := newsdata.NewClient()

    // Initialize AWS config and S3 client
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        log.Fatalf("unable to load AWS SDK config, %v", err)
    }
    s3Client := s3.NewFromConfig(cfg)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    // Create channels for streaming articles
    articleChan, errChan := client.LatestNews.Stream(
        ctx,
        "technology",
        newsdata.WithLanguages("en"),
        newsdata.WithCategories("technology"),
    )

    const (
        bucketName = "your-bucket-name"
        objectPrefix = "news-articles/"
    )

    // Process articles and upload to S3
    for {
        select {
        case article, ok := <-articleChan:
            if !ok {
                return // Channel closed, all articles processed
            }

            // Convert article to JSON
            articleJSON, err := json.Marshal(article)
            if err != nil {
                log.Printf("Error marshaling article: %v", err)
                continue
            }

            // Generate S3 key using article ID or timestamp
            objectKey := fmt.Sprintf("%s%s-%s.json",
                objectPrefix,
                article.PubDate.Time.Format("2006-01-02"),
                article.ArticleID,
            )

            // Upload to S3
            _, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
                Bucket: &bucketName,
                Key:    &objectKey,
                Body:   bytes.NewReader(articleJSON),
            })
            if err != nil {
                log.Printf("Error uploading to S3: %v", err)
                continue
            }

            fmt.Printf("Uploaded article %s to S3\n", article.Title)

        case err := <-errChan:
            if err != nil {
                log.Printf("Stream error: %v", err)
                return
            }
            return

        case <-ctx.Done():
            log.Printf("Context done: %v", ctx.Err())
            return
        }
    }
}
```

## Available Services

### LatestNews Service

- Real-time news articles
- Streaming support
- Full filtering capabilities

### NewsArchive Service

- Historical news articles
- Date range filtering
- Comprehensive search options

### CryptoNews Service

- Cryptocurrency-specific news
- Coin filtering
- Blockchain and crypto tags

### Sources Service

- News source metadata
- Source filtering by country/language
- Priority domain support

## Article Features

Articles include rich metadata:

- Basic Information
  - Title, description, and content
  - Publication date with timezone
  - Keywords and categories
- Source Details
  - Name and URL
  - Priority ranking
  - Icon URL
- Media Content
  - Image URLs
  - Video URLs
- Classification
  - Language
  - Countries
  - Categories
- AI Features
  - Topic tags
  - Geographic regions
  - Sentiment analysis with statistics
- Cryptocurrency
  - Coin mentions
  - Crypto-specific tags
- Quality Control
  - Duplicate detection
  - Content availability flags

## Error Handling

The client provides detailed error information:

- API-specific error codes and messages
- HTTP transport errors
- Context cancellation
- Parameter validation errors
- Rate limiting information

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
