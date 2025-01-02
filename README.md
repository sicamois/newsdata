# NewsData.io Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/sicamois/newsdata.svg)](https://pkg.go.dev/github.com/sicamois/newsdata)
[![Go Report Card](https://goreportcard.com/badge/github.com/sicamois/newsdata)](https://goreportcard.com/report/github.com/sicamois/newsdata)

A Go client library for the [NewsData.io](https://newsdata.io) API that provides easy access to news articles and sources.

## Features

- Fetch news articles with customizable filters
- Stream articles for efficient processing
- Access news sources information
- Configurable HTTP client timeout
- Customizable logging with different log levels
- Full support for NewsData.io API v1
- Sentiment analysis support
- AI-powered tags and regions
- Cryptocurrency tags support

## Installation

```bash
go get github.com/sicamois/newsdata
```

## Quick Start

You will need to sign-up for a free account on [NewsData.io](https://newsdata.io) to get an API key [here](https://newsdata.io/api/register).

```go
package main

import (
    "context"
    "fmt"
    "github.com/sicamois/newsdata"
)

func main() {
    // Create a new client with your API key
    client := newsdata.NewClient("your-api-key")

    // Create a request to fetch news articles
    req := newsdata.NewArticleRequest("artificial intelligence").
        WithKeywords("technology").
        WithLanguage("en")

    // Get articles (limited to 10 results)
    articles, err := client.GetArticles(req, 10)
    if err != nil {
        panic(err)
    }

    // Process the articles
    for _, article := range articles {
        fmt.Printf("Title: %s\n", article.Title)
        fmt.Printf("Link: %s\n", article.Link)
        fmt.Printf("Published: %s\n", article.PubDate.Time)
        if article.Sentiment != "" {
            fmt.Printf("Sentiment: %s\n", article.Sentiment)
        }
        fmt.Println()
    }
}
```

## Streaming Articles

For handling large result sets efficiently:

```go
articleChan, errChan := client.StreamArticles(req, 0) // 0 means no limit

for {
    select {
    case article, ok := <-articleChan:
        if !ok {
            return // Channel closed, all articles processed
        }
        // Process each article
        fmt.Printf("Title: %s\n", article.Title)
    case err := <-errChan:
        if err != nil {
            panic(err)
        }
    }
}
```

## Customization

### Setting Timeout

```go
client.SetTimeout(10 * time.Second)
```

### Configuring Logging

```go
// Enable debug logging
client.EnableDebug()

// Customize logging
client.CustomizeLogging(os.Stdout, slog.LevelDebug)
```

## Article Features

Articles include rich metadata:

- Title, description, and content
- Publication date with timezone
- Source information (name, URL, priority)
- Media URLs (images, videos)
- Categories and keywords
- Language and countries
- AI-powered tags and regions
- Sentiment analysis with detailed statistics
- Cryptocurrency-specific tags
- Duplicate detection

## Available Methods

### Articles

- `GetArticles(req ArticleRequest, maxResults int) ([]*Article, error)`
- `StreamArticles(req ArticleRequest, maxResults int) (<-chan *Article, <-chan error)`

### Sources

- `GetSources(req SourceRequest) ([]*Source, error)`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
