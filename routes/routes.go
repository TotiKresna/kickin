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
		r.Get("/allcourts", handlers.GetCourt(db))

		r.Route("/mybookings", func(r chi.Router) {

			r.Post("/", handlers.CreateBooking(db))
			r.Get("/", handlers.GetMyBookings(db))
			r.Get("/{id}", handlers.GetMyBookingByID(db))
		})
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

		r.Route("/courts", func(r chi.Router) { // Create, Update, Delete Court only superadmin
			
			r.Post("/create", handlers.CreateCourt(db))
			r.Put("/update/{id}", handlers.UpdateCourt(db))
			r.Delete("/delete/{id}", handlers.DeleteCourt(db))
		})

		r.Route("/bookings", func(r chi.Router) {
			r.Get("/", handlers.GetAllBookings(db))
			r.Get("/{id}", handlers.GetBookingByID(db))
			r.Put("/{id}", handlers.UpdateBooking(db))
			r.Delete("/{id}", handlers.DeleteBooking(db))
		})
	})

	r.Handle("/image/*", http.StripPrefix("/image/", http.FileServer(http.Dir("./assets/image"))))

	return r
}
