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

````go
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

## Returned Data Structures

### Article Structure

The `article` struct represents a news article with the following fields:

| Field          | Type     | Description                          |
| -------------- | -------- | ------------------------------------ |
| Id             | string   | Unique article identifier            |
| Title          | string   | Article title                        |
| Link           | string   | URL to the full article              |
| Keywords       | []string | Keywords associated with the article |
| Creator        | []string | Article authors                      |
| VideoURL       | string   | URL to associated video content      |
| Description    | string   | Brief article description            |
| Content        | string   | Full article content                 |
| PubDate        | DateTime | Publication date and time            |
| ImageURL       | string   | URL to article's featured image      |
| SourceId       | string   | News source identifier               |
| SourcePriority | int      | Priority level of the source         |
| SourceName     | string   | Name of the news source              |
| SourceURL      | string   | URL of the news source               |
| SourceIconURL  | string   | URL to source's icon                 |
| Language       | string   | Article language code                |
| Countries      | []string | Related country codes                |
| Categories     | []string | Article categories                   |
| AiTag          | string   | AI-generated topic tag               |
| Sentiment      | string   | Article sentiment analysis           |
| SentimentStats | string   | Detailed sentiment statistics        |
| AiRegion       | string   | AI-detected geographical region      |
| AiOrganization | string   | AI-detected organization             |
| Duplicate      | bool     | Indicates if article is a duplicate  |

### newsResponse Structure

The `newsResponse` struct represents the API response:

| Field        | Type      | Description                                     |
| ------------ | --------- | ----------------------------------------------- |
| Status       | string    | Response status ("success" or error message)    |
| TotalResults | int       | Total number of articles matching the query     |
| Articles     | []article | Array of articles (see Article Structure above) |
| NextPage     | string    | Token for fetching the next page of results     |

→ For details, see the [Response Object API documentation](https://newsdata.io/documentation/#http_response)

## Query Options & Parameters

### Latest News

#### Advanced Search Options (NewsQueryOptions)

| Parameter         | Type     | Description                                           |
| ----------------- | -------- | ----------------------------------------------------- |
| QueryInTitle      | string   | Search term in article title only                     |
| QueryInMetadata   | string   | Search in metadata (titles, URL, keywords)            |
| Timeframe         | string   | Filter by hours (e.g., "24") or minutes (e.g., "30m") |
| Categories        | []string | Filter by categories (max 5)                          |
| ExcludeCategories | []string | Categories to exclude (max 5)                         |
| Countries         | []string | Filter by country codes (max 5)                       |
| Languages         | []string | Filter by language codes (max 5)                      |

#### Direct API Parameters (NewsQueryParams)

| Parameter         | Type     | Description                                           |
| ----------------- | -------- | ----------------------------------------------------- |
| Query             | string   | Main search term                                      |
| QueryInTitle      | string   | Search term in article title only                     |
| QueryInMetadata   | string   | Search in metadata (titles, URL, keywords)            |
| Timeframe         | string   | Filter by hours (e.g., "24") or minutes (e.g., "30m") |
| Categories        | []string | Filter by categories (max 5)                          |
| ExcludeCategories | []string | Categories to exclude (max 5)                         |
| Countries         | []string | Filter by country codes (max 5)                       |
| Languages         | []string | Filter by language codes (max 5)                      |
| Domains           | []string | Include specific domains (max 5)                      |
| ExcludeDomains    | []string | Exclude specific domains (max 5)                      |
| PriorityDomain    | string   | Filter by domain priority ("Top", "Medium", "Low")    |
| RemoveDuplicates  | string   | Remove duplicate articles ("1" for true)              |
| Size              | int      | Results per page (max 50)                             |

→ For details, see the [Latest News API documentation](https://newsdata.io/documentation/#latest-news)

### Crypto News

#### Advanced Search Options (CryptoQueryOptions)

| Parameter       | Type     | Description                                             |
| --------------- | -------- | ------------------------------------------------------- |
| QueryInTitle    | string   | Search term in article title only                       |
| QueryInMetadata | string   | Search in metadata                                      |
| Timeframe       | string   | Filter by time period                                   |
| Languages       | []string | Filter by language codes (max 5)                        |
| Tags            | []string | Filter by crypto-specific tags                          |
| Sentiment       | string   | Filter by sentiment ("positive", "negative", "neutral") |

#### Direct API Parameters (CryptoQueryParams)

| Parameter        | Type     | Description                                             |
| ---------------- | -------- | ------------------------------------------------------- |
| Query            | string   | Main search term                                        |
| Coins            | []string | Filter by specific cryptocurrencies                     |
| QueryInTitle     | string   | Search term in article title only                       |
| QueryInMetadata  | string   | Search in metadata                                      |
| Timeframe        | string   | Filter by time period                                   |
| Languages        | []string | Filter by language codes (max 5)                        |
| Tags             | []string | Filter by crypto-specific tags                          |
| Sentiment        | string   | Filter by sentiment ("positive", "negative", "neutral") |
| RemoveDuplicates | string   | Remove duplicate articles ("1" for true)              |
| Size             | int      | Results per page (max 50)                               |

→ For details, see the [Crypto News API documentation](https://newsdata.io/documentation/#crypto-news)

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

The library uses Go's `slog` package for logging. You can customize logging by specifying an output writer and log level:

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

The client's logger can be:

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
````
