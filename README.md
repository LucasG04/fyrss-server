# ForYou RSS Server

A Go backend for automated curation of news and blog articles via RSS and web scraping. Content is categorized, prioritized, and made accessible via a REST API. The backend is fully stateless and uses an external database (e.g., PostgreSQL in a container).

## Features

- Periodic fetching of RSS feeds and websites (scraping)
- LLM-based categorization, tagging, and summarization (OpenAI GPT)
- Duplicate detection via content hash
- Storage of all content in an external PostgreSQL database
- REST API for querying, filtering, and displaying content
- Configuration via ENV variables (feeds, URLs, API keys, etc.)
- Completely stateless backend (suitable for containerization)

## ENV Configuration

| Variable              | Description                          |
| --------------------- | ------------------------------------ |
| `SCRAPER_RSS_FEEDS`   | Comma-separated list of RSS URLs     |
| `SCRAPER_RAW_URLS`    | Comma-separated list of website URLs |
| `OPENAI_API_KEY`      | API key for GPT categorization       |
| `DATABASE_URL`        | PostgreSQL connection URL            |
| `SCRAPE_INTERVAL_MIN` | Scraping interval in minutes         |

## REST API

Base route: `/api`

| Method | Route                | Description                          |
| ------ | -------------------- | ------------------------------------ |
| GET    | `/api/articles`      | Returns all articles                 |
| GET    | `/api/articles/:id`  | Returns a single article             |
| GET    | `/api/articles/feed` | Returns a filtered feed              |
| POST   | `/api/trigger`       | Manually triggers a scraping process |

## Docker Setup (recommended)

TODO
