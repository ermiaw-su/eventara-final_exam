package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"final/middleware"
	"final/worker"
)

type BookingHandler struct {
	DB   *sql.DB
	Pool *worker.Pool
}

func NewBookingHandler(
	db *sql.DB,
	pool *worker.Pool,
) *BookingHandler {

	return &BookingHandler{
		DB:   db,
		Pool: pool,
	}
}

type BookingResponse struct {
	ID        int         `json:"id"`
	Status    string      `json:"status"`
	Event     interface{} `json:"event"`
	CreatedAt string      `json:"created_at"`
}

func (h *BookingHandler) BookEvent(w http.ResponseWriter, r *http.Request) {

	// Make sure request method is POST
	if r.Method != http.MethodPost {
		return
	}

	// Make sure request path is /api/events/:id/book
	if !strings.HasSuffix(r.URL.Path, "/book") {
		return
	}

	// Get event id from middleware
	userID := r.Context().Value(middleware.UserIDKey).(int)

	// Get event id from request path
	path := strings.TrimSuffix(r.URL.Path, "/book") // /api/events/:id
	path = strings.TrimPrefix(path, "/api/events/") // :id

	// Convert event id to int
	eventID, err := strconv.Atoi(path)

	if err != nil {
		writeError(w, 400, "invalid event id")
		return
	}

	// Insert booking
	res, err := h.DB.Exec(
		"INSERT INTO bookings(user_id,event_id,status) VALUES(?,?,?)",
		userID,
		eventID,
		"pending",
	)

	if err != nil {
		writeError(w, 500, "database error")
		return
	}

	// Get last inserted id
	bookingID, _ := res.LastInsertId()

	// Submit job
	h.Pool.Submit(worker.Job{
		BookingID: int(bookingID),
		EventID:   eventID,
		UserID:    userID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"booking_id": bookingID,
		"status":     "pending",
	})
}

func (h *BookingHandler) GetBookings(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/api/bookings" {
		http.NotFound(w, r)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)

	rows, err := h.DB.Query(`
	SELECT
	b.id,
	b.status,
	e.title,
	e.date,
	e.location,
	b.created_at
	FROM bookings b
	JOIN events e ON b.event_id = e.id
	WHERE b.user_id = ?
	ORDER BY b.id DESC
	`,
		userID,
	)

	if err != nil {
		writeError(w, 500, "database error")
		return
	}

	defer rows.Close()

	result := []BookingResponse{}

	for rows.Next() {

		var b BookingResponse

		var title string
		var date string
		var location string

		rows.Scan(
			&b.ID,
			&b.Status,
			&title,
			&date,
			&location,
			&b.CreatedAt,
		)

		b.Event = map[string]interface{}{
			"title":    title,
			"date":     date,
			"location": location,
		}

		result = append(result, b)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(result)
}

func (h *BookingHandler) GetBookingByID(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value(middleware.UserIDKey).(int)

	idStr := strings.TrimPrefix(r.URL.Path, "/api/bookings/")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		writeError(w, 400, "invalid booking id")
		return
	}

	var b BookingResponse

	var ownerID int
	var title string
	var date string
	var location string

	err = h.DB.QueryRow(`
	SELECT
	b.id,
	b.user_id,
	b.status,
	e.title,
	e.date,
	e.location,
	b.created_at
	FROM bookings b
	JOIN events e ON b.event_id = e.id
	WHERE b.id = ?
	`,
		id,
	).Scan(
		&b.ID,
		&ownerID,
		&b.Status,
		&title,
		&date,
		&location,
		&b.CreatedAt,
	)

	if err != nil {
		writeError(w, 404, "booking not found")
		return
	}

	if ownerID != userID {
		writeError(w, 403, "forbidden")
		return
	}

	b.Event = map[string]interface{}{
		"title":    title,
		"date":     date,
		"location": location,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(b)
}