package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

type Event struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Date time.Time `json:"date"`
	Location string `json:"location"`
	Category string `json:"category"`
	Capacity int `json:"capacity"`
	AvailableSeats int `json:"available_seats"`
	Price int `json:"price"`
}

type BookingEvent struct {
    Title    string    `json:"title"`
    Date     time.Time `json:"date"`
    Location string    `json:"location"`
}

type Booking struct {
    ID        int          `json:"id"`
    Status    string       `json:"status"`
    Event     BookingEvent `json:"event"`
    CreatedAt time.Time    `json:"created_at"`
}

type Job struct {
	BookingID int
	EventID   int
	UserID    int
}