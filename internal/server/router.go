package server

import (
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/lucasg04/nuntium/internal/handler"
	"github.com/lucasg04/nuntium/internal/repository"
	"github.com/lucasg04/nuntium/internal/service"
)

func NewRouter() *chi.Mux {
	dbUrl := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", dbUrl)
	if err != nil {
		panic(err)
	}
	articleRepo := repository.NewArticleRepository(db)
	articleService := service.NewArticleService(articleRepo)
	articleHandler := handler.NewArticleHandler(articleService)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", articleHandler.GetAll)
			r.Get("/{id}", articleHandler.GetByID)
		})
	})

	return r
}
