package worker

import (
	"database/sql"
	"time"
	"sync"
)

type Job struct {
	BookingID int
	EventID int
	UserID int
}

type Pool struct {
	jobs chan Job
	db *sql.DB
	mu sync.Mutex
}

func NewPool(workers int, buffer int, db *sql.DB) *Pool {
	p := &Pool{
		jobs: make(chan Job, buffer),
		db: db,
	}

	// As long as there are jobs in the channel and we have workers, process them
	for i := 0; i < workers; i++ {
		go p.work()
	}

	return p
}

// Add a job to the pool
func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

// Process jobs
func (p *Pool) work() {
	for job := range p.jobs {
		p.process(job)
	}
}

func (p *Pool) process(job Job) {
	// Lock
	p.mu.Lock()

	// Unlock later
	defer p.mu.Unlock()

	var seats int

	// Get available seats
	p.db.QueryRow(
		"SELECT available_seats FROM events WHERE id=?",
		job.EventID,
	).Scan(&seats)

	// If there are no available seats
	if seats <= 0 {
		p.db.Exec(
			"UPDATE bookings SET status='failed' WHERE id=?",
			job.BookingID,
		)
		return
	}

	// Decrement available seats
	p.db.Exec(
		"UPDATE events SET available_seats=availabale_seats-1 WHERE id=?",
		job.EventID,
	)

	// Confirm booking
	p.db.Exec(
		"UPDATE bookings SET status='confirmed' WHERE id=?",
		job.BookingID,
	)

	time.Sleep(5 * time.Second)
}