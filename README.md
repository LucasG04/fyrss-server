# ForYou RSS Server

A Go backend for automated curation of news and blog articles via RSS. Content is categorized, prioritized, and made accessible via a REST API. The backend is fully stateless and uses an external database (e.g., PostgreSQL in a container).

## Features

- Periodic fetching of RSS feeds
- Duplicate detection via content hash
- Storage of all content in an external PostgreSQL database
- REST API for querying, filtering, and displaying content
- Configuration via ENV variables
- Completely stateless

## ENV Configuration

| Variable               | Description                       |
| ---------------------- | --------------------------------- |
| `RSS_FEED_INTERVAL_MS` | Scraping interval in milliseconds |
| `DATABASE_URL`         | PostgreSQL connection URL         |
| `PORT`                 | Port for the REST API server      |

## Docker Setup

TODO
