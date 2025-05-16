package routes

import (
	"database/sql"
	"kickin/config"
	"kickin/handlers"
	"kickin/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRoutes(db *sql.DB, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// === CORS Middleware ===
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigins},	
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-CSRF-Token"},
		AllowCredentials: true,
	}))
	

	// Public routes
	r.Get("/", handlers.RootHandlerWithDB(db))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", handlers.Login(db))
		r.Post("/register", handlers.Register(db))
		r.Post("/logout", handlers.Logout)
		r.Post("/refresh", handlers.RefreshToken)
	})


	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RefreshMiddleware)
		r.Use(middleware.AuthMiddleware)

		r.Get("/me", handlers.GetMe)
	})

	return r
}
