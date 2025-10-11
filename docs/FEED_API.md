# Feed Management API

The Feed Management API allows you to manage RSS feed URLs through HTTP endpoints. **New feature**: The API now validates that URLs actually return valid RSS/Atom feeds before saving them.

## Base URL

All endpoints are available under `/api/feeds`

## RSS Feed Validation

When creating or updating feeds, the system will:

1. Validate the URL format (must be http/https with valid host)
2. **Fetch and parse the RSS/Atom feed** to ensure it's valid
3. Check that the feed has required elements (title, feed type)
4. Ensure no duplicate URLs exist

## Endpoints

### GET /api/feeds

Get all feeds.

**Response:** Array of feed objects

```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Example News",
    "url": "https://example.com/rss.xml",
    "createdAt": "2023-10-11T10:00:00Z",
    "updatedAt": "2023-10-11T10:00:00Z"
  }
]
```

### GET /api/feeds/{id}

Get a specific feed by ID.

**Response:** Feed object

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Example News",
  "url": "https://example.com/rss.xml",
  "createdAt": "2023-10-11T10:00:00Z",
  "updatedAt": "2023-10-11T10:00:00Z"
}
```

### POST /api/feeds

Create a new feed.

**Request Body:**

```json
{
  "name": "Example News",
  "url": "https://example.com/rss.xml"
}
```

**Response:** Created feed object (201 Created)

**Validation:** The URL will be validated to ensure it returns a valid RSS/Atom feed.

### PUT /api/feeds/{id}

Update an existing feed.

**Request Body:**

```json
{
  "name": "Updated News Name",
  "url": "https://updated.example.com/rss.xml"
}
```

**Response:** Updated feed object

**Validation:** The URL will be validated to ensure it returns a valid RSS/Atom feed.

### DELETE /api/feeds/{id}

Delete a feed.

**Response:** 204 No Content on success

## Testing with curl

```bash
# Get all feeds
curl http://localhost:8080/api/feeds

# Create a new feed (will validate RSS)
curl -X POST http://localhost:8080/api/feeds \
  -H "Content-Type: application/json" \
  -d '{"name": "Tagesschau", "url": "https://www.tagesschau.de/index~rss2.xml"}'

# Try to create an invalid feed (will fail validation)
curl -X POST http://localhost:8080/api/feeds \
  -H "Content-Type: application/json" \
  -d '{"name": "Invalid", "url": "https://example.com/not-an-rss-feed"}'

# Get a specific feed
curl http://localhost:8080/api/feeds/{feed-id}

# Update a feed (will validate RSS)
curl -X PUT http://localhost:8080/api/feeds/{feed-id} \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "url": "https://updated-url.com/rss.xml"}'

# Delete a feed
curl -X DELETE http://localhost:8080/api/feeds/{feed-id}
```

## Error Responses

- **400 Bad Request**:
  - Invalid request body or parameters
  - Invalid feed name (empty)
  - Invalid URL format
  - **URL does not return a valid RSS/Atom feed**
- **404 Not Found**: Feed not found
- **409 Conflict**: Duplicate feed URL
- **500 Internal Server Error**: Server error

## RSS Validation Details

The validation process:

1. **Format Check**: Ensures URL is valid http/https with proper host
2. **Feed Fetch**: Actually fetches the URL with a 30-second timeout
3. **Parse Check**: Uses gofeed parser to validate RSS/Atom format
4. **Content Check**: Ensures feed has title and determinable feed type
5. **User Agent**: Requests include a proper user agent for better compatibility

Supported feed formats:

- RSS 1.0, 2.0
- Atom
- JSON Feed
- Any format supported by the gofeed library

## Integration Notes

The RSS reader automatically uses feeds from the database instead of environment variables. When feeds are added/updated/deleted through the API, they will be automatically picked up in the next RSS reading cycle.

**Performance Note**: RSS validation adds a network request during feed creation/update. This ensures feed quality but may add 1-5 seconds to the API response time depending on the target RSS server response time.
