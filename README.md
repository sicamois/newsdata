# newsdata - Go Client for newsdata.io API

[![Go Reference](https://pkg.go.dev/badge/github.com/sicamois/newsdata.svg)](https://pkg.go.dev/github.com/sicamois/newsdata)
[![Go Report Card](https://goreportcard.com/badge/github.com/sicamois/newsdata)](https://goreportcard.com/report/github.com/sicamois/newsdata)

A Go client library for accessing the [newsdata.io](https://newsdata.io) API.

## Key Features

- Support for Latest News, Crypto News, Historical News and Sources endpoints
- Asynchronous article processing through pipe functions
- Automatic pagination handling
- Async article processing with custom actions
- Customizable logging
- Request timeout configuration
- Result limiting
- Input validation
- Full access to raw API parameters

## Installation

```go
go get github.com/sicamois/newsdata
```

## Requirement

You need a [newsdata.io](https://newsdata.io) API key to use this library.

→ To get an API key, you can [sign up for a free account](https://newsdata.io/register).

## Usage

```go
// Create a new client
client := newsdata.NewClient("your-api-key")

// Get breaking news about climate change
query := BreakingNewsQuery{
    Query: "climate change",
    Languages: []string{"en", "fr"},
    Categories: []string{"environment", "science"},
    Countries: []string{"us", "gb", "fr"},
}

// Get the first 100 breaking news about climate change
Articles, err := client.GetBreakingNews(query, 100)

// Get US news sources
Sources, err := client.GetSources(SourcesQuery{
    Country: "us",
})

// Get crypto news about Bitcoin
cryptoQuery := CryptoNewsQuery{
    Query: "Bitcoin",
    Languages: []string{"en"},
    Timeframe: "24",
}

CryptoArticles, err := client.GetCryptoNews(cryptoQuery, 50)

// Search news archive
historicalQuery := HistoricalNewsQuery{
    Query: "Olympics",
    Languages: []string{"en"},
    From: "2021-07-23",
    To: "2021-08-08",
}

HistoricalArticles, err := client.GetHistoricalNews(historicalQuery, 100)
```

## Asynchronous Processing with Pipe Functions

The library provides efficient asynchronous processing capabilities through its Pipe functions. These functions return channels that allow for non-blocking article processing:

```go
// Create a new client
client := newsdata.NewClient("your-api-key")

// Configure query
query := BreakingNewsQuery{
    Query:     "artificial intelligence",
    Languages: []string{"en"},
}

// Get channels for articles and errors
articleChan, errChan := client.PipeBreakingNews(query, 100)

// Process articles asynchronously
for {
    select {
    case article, ok := <-articleChan:
        if !ok {
            // Channel is closed, all articles have been processed
            return
        }
        // Process each article asynchronously
        fmt.Printf("Processing article: %s\n", article.Title)
    case err := <-errChan:
        if err != nil {
            log.Fatal(err)
        }
    }
}
```

Similar pipe functions are available for other endpoints:

- `PipeBreakingNews`: Process latest news articles asynchronously
- `PipeHistoricalNews`: Process historical news articles asynchronously
- `PipeCryptoNews`: Process cryptocurrency news articles asynchronously

These functions are particularly useful when:

- Processing large datasets efficiently
- Implementing non-blocking article processing
- Building asynchronous processing pipelines
- Handling articles concurrently

## Advanced Client Configuration

### Setting Timeout

```go
client := newsdata.NewClient("your-api-key", 0)
client.SetTimeout(20 * time.Second)
```

### Debug Logging

```go
client := newsdata.NewClient("your-api-key", 0)
client.EnableDebug() // Enable debug logging
// ... perform operations ...
client.DisableDebug() // Disable debug logging
```

### Custom Logging

The library uses Go `slog` package for logging. You can customize logging by specifying an output writer and log level:

```go
// Create or open a log file
file, err := os.OpenFile("logs", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
defer file.Close()

// Configure custom logging
client.CustomizeLogging(file, slog.LevelInfo)

// Get the logger for use in your application
logger := client.Logger

// Use the logger for your own logging needs
logger.Info("Starting news search...")

// The logger will also be used internally by the client
Articles, err := client.GetBreakingNews(query)
if err != nil {
    logger.Error(err.Error())
    return
}

logger.Info("Articles retrieved", "count", len(*Articles))
```

The client logger can be:

- Customized with any `io.Writer` (file, stdout, network writer, etc.)
- Set to different log levels: `slog.LevelDebug`, `slog.LevelInfo`, `slog.LevelWarn`, `slog.LevelError`
- Retrieved for use in your application via `client.Logger()`

## Complete Example

```go
func main() {
    client := newsdata.NewClient("your-api-key")

    // Configure client
    client.SetTimeout(15 * time.Second)

    // Perform an advanced search
    query := BreakingNewsQuery{
        Query:     "artificial intelligence",
        Languages:  []string{"en"},
        Categories: []string{"technology"},
        Timeframe:  "24",
    }

    Articles, err := client.GetBreakingNews(query, 100)
    if err != nil {
        log.Fatal(err)
    }

    // Process results
    for _, Article := range *Articles {
        fmt.Printf("Title: %s\nSource: %s\nPublished: %s\n\n",
            Article.Title,
            Article.SourceName,
            Article.PubDate)
    }
}
```

## License

[MIT License](LICENSE)
