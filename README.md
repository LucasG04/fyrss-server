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

| Variable               | Description                              |
| ---------------------- | ---------------------------------------- |
| `RSS_FEED_URLS`        | Comma-separated list of RSS URLs         |
| `RSS_FEED_INTERVAL_MS` | Scraping interval in minutes             |
| `DATABASE_URL`         | PostgreSQL connection URL                |
| `PORT`                 | Port for the REST API server             |
| `OPENAI_CHAT_MODEL`    | OpenAI GPT model name for categorization |
| `OPENAI_API_KEY`       | OpenAI API key for LLM-based features    |

## Docker Setup (recommended)

TODO
