package db

// Import SQL database
import "database/sql"

// Initialize database
func InitDB(db *sql.DB) {
    // State the user schema
    user := `
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT,
            email TEXT UNIQUE,
            password TEXT
        );`

    // State the events schema
    events := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        description TEXT,
        date TEXT,
        location TEXT,
        category TEXT,
        capacity INTEGER,
        available_seats INTEGER,
        price INTEGER
    );`

    // State the bookings schema
    bookings := `
    CREATE TABLE IF NOT EXISTS bookings (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        event_id INTEGER,
        status TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

    // Execute the schema
    db.Exec(user)
    db.Exec(events)
    db.Exec(bookings)
}

// Default events
func SeedEvents(db *sql.DB) {

    // Insert default events
    db.Exec(`
        INSERT OR IGNORE INTO events
        (title, description, date, location, category, capacity, available_seats, price)
        VALUES
        (
            'Go Conference 2025',
            'Go concurrency',
            '2025-08-15T09:00:00Z',
            'Jakarta',
            'Technology',
            200,
            1,
            150000
        ),
        (
            'UI Workshop',
            'Mobile Design',
            '2025-08-22T10:00:00Z',
            'SGU',
            'Design',
            50,
            0,
            75000
        ),
        (
            'API Security',
            'REST Security',
            '2025-09-05T13:00:00Z',
            'Zoom',
            'Security',
            500,
            0,
            0
        )
    `)
}