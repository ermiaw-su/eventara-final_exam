package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"final/db"
	"final/handler"
	"final/middleware"
	"final/worker"
)

func main() {
	// Connect to database
	database, err := sql.Open("sqlite3", "./eventara.db")

	if err != nil {
		log.Fatal(err)
	}

	// Initialize database
	db.InitDB(database)

	// Seed events
	db.SeedEvents(database)

	// Create worker pool
	pool := worker.NewPool(5, 100, database)

	authHandler := handler.NewAuthHandler(database)
	eventHandler := handler.NewEventHandler(database)
	bookingHandler := handler.NewBookingHandler(database, pool)

	// Create router
	mux := http.NewServeMux()

	// AUTH
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)

	// EVENTS
	mux.HandleFunc("/api/events", eventHandler.GetEvents)

	mux.HandleFunc("/api/events/", func(w http.ResponseWriter, r *http.Request) {

		// BOOKING
		if strings.HasSuffix(r.URL.Path, "/book") {

			middleware.AuthMiddleware(
				http.HandlerFunc(bookingHandler.BookEvent),
			).ServeHTTP(w, r)

			return
		}

		// DETAIL EVENT
		eventHandler.GetEventByID(w, r)
	})

	// BOOKINGS
	mux.Handle(
		"/api/bookings",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.GetBookings)),
	)

	mux.Handle(
		"/api/bookings/",
		middleware.AuthMiddleware(http.HandlerFunc(bookingHandler.GetBookingByID)),
	)

	rl := middleware.NewRateLimiter(10, 10*time.Second)

	log.Println("Server running on :8080")

	http.ListenAndServe(
		":8080",
		middleware.CORSMiddleware(
			middleware.RateLimitMiddleware(rl, mux),
		),
	)
}