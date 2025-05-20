package routes

import (
	"kickin/config"
	"kickin/handlers"
	"kickin/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) http.Handler {
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
		r.Post("/refresh", handlers.RefreshToken(db))
	})
	
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RefreshMiddleware)
		r.Use(middleware.AuthMiddleware)

		r.Get("/me", handlers.GetMe)
		r.Put("/user/{id}", handlers.UpdateUser(db))
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RefreshMiddleware)
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RoleMiddleware("superadmin"))

		r.Route("/user", func(r chi.Router) {
			r.Get("/", handlers.GetAllUsers(db))
			r.Get("/{id}", handlers.GetUserByID(db))
			r.Delete("/{id}", handlers.DeleteUser(db))
		})

		r.Route("/logs", func(r chi.Router) {
			r.Get("/", handlers.ViewLogs)
			r.Delete("/", handlers.ClearLogs)
		})
	})



	return r
}
