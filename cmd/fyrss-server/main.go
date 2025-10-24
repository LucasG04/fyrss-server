package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lucasg04/fyrss-server/internal/handler"
	"github.com/lucasg04/fyrss-server/internal/repository"
	"github.com/lucasg04/fyrss-server/internal/service"

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
	articleRepo := repository.NewArticleRepository(db)
	articleService := service.NewArticleService(articleRepo)
	feedRepo := repository.NewFeedRepository(db)
	rssReader := service.NewRssArticleReader(articleService)
	feedService := service.NewFeedService(feedRepo, rssReader, articleService)

	runMigrations(databaseUrl)
	go startReadingRssFeeds(feedService)
	go startDeleteOldArticlesJob(articleService)

	startServer(articleService, feedService)
}

func startServer(articleService *service.ArticleService, feedService *service.FeedService) {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	setupArticleHttpHandler(r, articleService)
	setupFeedHttpHandler(r, feedService, articleService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s\n", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func setupArticleHttpHandler(r *chi.Mux, articleService *service.ArticleService) {
	articleHandler := handler.NewArticleHandler(articleService)

	r.Route("/api/articles", func(r chi.Router) {
		r.Get("/", articleHandler.GetAll)
		r.Get("/history", articleHandler.GetHistory)
		r.Get("/saved", articleHandler.GetSaved)

		r.Get("/{id}", articleHandler.GetByID)
		r.Patch("/{id}/saved", articleHandler.UpdateSavedByID)
		r.Patch("/{id}/read", articleHandler.UpdateReadByID)
	})
}

func setupFeedHttpHandler(r *chi.Mux, feedService *service.FeedService, articleService *service.ArticleService) {
	feedHandler := handler.NewFeedHandler(feedService)
	articleHandler := handler.NewArticleHandler(articleService)

	r.Route("/api/feeds", func(r chi.Router) {
		r.Get("/", feedHandler.GetAll)
		r.Get("/{id}", feedHandler.GetByID)
		r.Post("/", feedHandler.Create)
		r.Put("/{id}", feedHandler.Update)
		r.Delete("/{id}", feedHandler.Delete)
		r.Patch("/{id}/read", feedHandler.UpdateLastReadAt)
		r.Get("/{feedId}/paginated", articleHandler.GetPaginatedByFeedID)
	})
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

func startReadingRssFeeds(feedService *service.FeedService) {
	intervalInMs := os.Getenv("RSS_FEED_INTERVAL_MS")
	if intervalInMs == "" {
		intervalInMs = "7200000" // Default to 2 hours
	}
	interval, err := strconv.Atoi(intervalInMs)
	if err != nil {
		log.Fatalf("Invalid RSS_FEED_INTERVAL_MS: %s", intervalInMs)
	}

	fmt.Printf("Starting RSS feed reader with interval: %d ms\n", interval)
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	processRssFeeds(feedService) // Initial processing before starting the ticker
	for range ticker.C {
		processRssFeeds(feedService)
	}
}

func processRssFeeds(feedService *service.FeedService) {
	// Get all feeds from database
	feeds, err := feedService.GetAll(context.Background())
	if err != nil {
		log.Printf("Error getting feeds from database: %v\n", err)
		return
	}

	if len(feeds) == 0 {
		log.Println("No feeds configured in database")
		return
	}

	fmt.Printf("Starting scheduled processing of %d feeds\n", len(feeds))

	// TODO: add parallel processing for feeds
	for _, feed := range feeds {
		// cancel article read after 30 seconds to avoid blocking
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		err := feedService.ProcessFeedNow(ctx, feed)
		if err != nil {
			log.Printf("Error processing RSS feed %s (%s): %v\n", feed.Name, feed.URL, err)
		}

		cancel()
	}

	fmt.Println("Finished scheduled RSS feed processing cycle")
}

func startDeleteOldArticlesJob(articleService *service.ArticleService) {
	interval := 24 * time.Hour // Default to 24 hours
	ticker := time.NewTicker(interval)

	for range ticker.C {
		err := articleService.DeleteOneWeekOldArticles(context.Background())
		if err != nil {
			log.Printf("Error deleting old articles: %v\n", err)
		} else {
			log.Println("Deleted articles older than one week")
		}
	}
}
