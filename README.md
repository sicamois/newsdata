# newsdata - Go Client for newsdata.io API

A Go client library for accessing the [newsdata.io](https://newsdata.io) API.

## Key Features

- Support for Latest News, Crypto News, Historical News and Sources endpoints
- Automatic pagination handling
- Customizable logging
- Request timeout configuration
- Result limiting
- Input validation
- Full access to raw API parameters

## Available Methods

### GetBreakingNews

Get the latest news articles in real-time from various sources worldwide. Filter by categories, countries, languages and more.

### GetCryptoNews

Get cryptocurrency-related news with additional filters like coin symbols, sentiment analysis, and specialized crypto tags.

### GetHistoricalNews

Search through news archives with date range filters while maintaining all filtering capabilities of breaking news.

### GetSources

Get information about available news sources with filters for country, category, language and priority level.

## Installation

```go
go get github.com/sicamois/newsdata
```

## Requirement

You need a [newsdata.io](https://newsdata.io) API key to use this library.

→ To get an API key, you can [sign up for a free account](https://newsdata.io/register).

## Usage

### Basic Usage

```go
// Create a new client
client := newsdata.NewClient("your-api-key", 0) // 0 means no limit on results

// Get breaking news about climate change
query := BreakingNewsQuery{
    Query: "climate change",
    Languages: []string{"en", "fr"},
    Categories: []string{"environment", "science"},
    Countries: []string{"us", "gb", "fr"},
}
articles, err := client.GetBreakingNews(query)
```

3. **Direct API Access** - For complete control over API parameters
   - Returns: `newsResponse` (see newsResponse Structure below)

```go
params := BreakingNewsQuery{
    Query: "bitcoin",
    Languages: []string{"en"},
    Size: 50,
    RemoveDuplicates: "1",
}
response, err := client.LatestNews.Get(&params)
```

## Advanced Client Configuration

### Setting Timeout

```go
client := newsdata.NewClient("your-api-key", 0)
client.SetTimeout(20 * time.Second)
```

### Limiting Results

```go
client := newsdata.NewClient("your-api-key", 100)
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
logger := client.Logger()

// Use the logger for your own logging needs
logger.Info("Starting news search...")

// The logger will also be used internally by the client
articles, err := client.GetBreakingNews(query)
if err != nil {
    logger.Error(err.Error())
    return
}

logger.Info("Articles retrieved", "count", len(*articles))
```

The client logger can be:

- Customized with any `io.Writer` (file, stdout, network writer, etc.)
- Set to different log levels: `slog.LevelDebug`, `slog.LevelInfo`, `slog.LevelWarn`, `slog.LevelError`
- Retrieved for use in your application via `client.Logger()`

## Complete Example

```go
func main() {
    client := newsdata.NewClient("your-api-key", 0)

    // Configure client
    client.SetTimeout(15 * time.Second)
    client.LimitResultsToFirst(50)

    // Perform an advanced search
    options := NewsQueryOptions{
        Languages:  []string{"en"},
        Categories: []string{"technology"},
        Timeframe:  "24",
    }

    articles, err := client.LatestNews.AdvancedSearch("artificial intelligence", options)
    if err != nil {
        log.Fatal(err)
    }

    // Process results
    for _, article := range *articles {
        fmt.Printf("Title: %s\nSource: %s\nPublished: %s\n\n",
            article.Title,
            article.SourceName,
            article.PubDate)
    }
}
```

## License

[MIT License](LICENSE)
