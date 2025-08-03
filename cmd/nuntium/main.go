package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lucasg04/nuntium/internal/handler"
	"github.com/lucasg04/nuntium/internal/repository"
	"github.com/lucasg04/nuntium/internal/service"
	"github.com/openai/openai-go"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.dev")
	godotenv.Load(".env.secrets")
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", databaseUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize services
	openAiClient := openai.NewClient() // retrieving api key defaults to os.LookupEnv("OPENAI_API_KEY")
	aiService := service.NewAiService(&openAiClient)
	articleRepo := repository.NewArticleRepository(db)
	articleService := service.NewArticleService(articleRepo, aiService)
	rssReader := service.NewRssArticleReader(articleService)

	runMigrations(databaseUrl)
	startReadingRssFeeds(rssReader, articleService)
	startServer(articleService)
}

func startServer(articleService *service.ArticleService) {
	setupArticleHttpHandler(articleService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func setupArticleHttpHandler(articleService *service.ArticleService) {
	articleHandler := handler.NewArticleHandler(articleService)

	http.HandleFunc("/api/v1/articles", articleHandler.GetAll)
	http.HandleFunc("/api/v1/articles/", articleHandler.GetByID)
}

func runMigrations(dbUrl string) {
	m, err := migrate.New(
		"file://db/migrations", dbUrl,
	)
	if err != nil {
		log.Fatal("Failed to create migration instance:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed to apply migrations:", err)
	}
	log.Println("Database migrations applied successfully")
}

func startReadingRssFeeds(rssReader *service.RssArticleReader, articleService *service.ArticleService) {
	feedUrlString := os.Getenv("RSS_FEED_URLS")
	feedUrls := strings.Split(feedUrlString, ",")
	intervalInMs := os.Getenv("RSS_FEED_INTERVAL_MS")
	if intervalInMs == "" {
		intervalInMs = "7200000" // Default to 2 hours
	}
	interval, err := strconv.Atoi(intervalInMs)
	if err != nil {
		log.Fatalf("Invalid RSS_FEED_INTERVAL_MS: %s", intervalInMs)
	}

	fmt.Printf("Starting RSS feed reader with interval: %d ms and feeds: %v\n", interval, feedUrls)
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()
	go processRssFeeds(rssReader, articleService, feedUrls) // Initial processing before starting the ticker
	go func() {
		for range ticker.C {
			processRssFeeds(rssReader, articleService, feedUrls)
		}
	}()
}

func processRssFeeds(rssReader *service.RssArticleReader, articleService *service.ArticleService, feedUrls []string) {
	skippedDuplicates := 0
	savedArticles := 0
	for _, feedUrl := range feedUrls {
		articles, err := rssReader.ReadArticleFeed(context.Background(), feedUrl)
		if err != nil {
			log.Printf("Error reading RSS feed from %s: %v\n", feedUrl, err)
			return
		}
		for _, article := range articles {
			err := articleService.Save(context.Background(), article)
			if err == service.ErrDuplicateArticle {
				skippedDuplicates++
				continue
			}
			if err != nil {
				log.Printf("Error saving article %s: %v\n", article.Title, err)
				continue
			}
			savedArticles++
		}
	}
	if skippedDuplicates > 0 {
		log.Printf("Skipped %d duplicate articles\n", skippedDuplicates)
	}
	fmt.Printf("Finished processing RSS feeds. Saved %d new articles.\n", savedArticles)
}
