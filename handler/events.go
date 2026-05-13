package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type EventHandler struct {
	DB *sql.DB
}

func NewEventHandler(db *sql.DB) *EventHandler {
	return &EventHandler{DB: db}
}

type Event struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Date           string `json:"date"`
	Location       string `json:"location"`
	Category       string `json:"category"`
	Capacity       int    `json:"capacity"`
	AvailableSeats int    `json:"available_seats"`
	Price          int    `json:"price"`
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	// Make sure request path is /api/events
	if r.URL.Path != "/api/events" {
		http.NotFound(w, r)
		return
	}

	// Get all events
	rows, err := h.DB.Query(`
		SELECT
		id,
		title,
		description,
		date,
		location,
		category,
		capacity,
		available_seats,
		price
		FROM events
	`)

	if err != nil {
		writeError(w, 500, "database error")
		return
	}

	// Close rows
	defer rows.Close()

	// Create events slice
	events := []Event{}

	// Loop through rows for input the event into slice
	for rows.Next() {

		var e Event

		rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Date,
			&e.Location,
			&e.Category,
			&e.Capacity,
			&e.AvailableSeats,
			&e.Price,
		)

		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(events)
}

func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {

	// Make sure request path is /api/events
	if !strings.HasPrefix(r.URL.Path, "/api/events/") {
		http.NotFound(w, r)
		return
	}

	// Make sure request path is /api/events/:id
	if strings.HasSuffix(r.URL.Path, "/book") {
		return
	}

	// Get event id from request path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/events/")

	// Convert event id to int
	id, err := strconv.Atoi(idStr)

	if err != nil {
		writeError(w, 400, "invalid id")
		return
	}

	var e Event

	err = h.DB.QueryRow(`
	SELECT
	id,
	title,
	description,
	date,
	location,
	category,
	capacity,
	available_seats,
	price
	FROM events
	WHERE id=?
	`,
		id,
	).Scan(
		&e.ID,
		&e.Title,
		&e.Description,
		&e.Date,
		&e.Location,
		&e.Category,
		&e.Capacity,
		&e.AvailableSeats,
		&e.Price,
	)

	if err != nil {
		writeError(w, 404, "event not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(e)
}