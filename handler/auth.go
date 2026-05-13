package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"final/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler have DB which is connected with SQLite database
type AuthHandler struct {
	DB *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

type AuthRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	// Make sure request method is POST
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	// Create variable 'req' from AuthRequest
	var req AuthRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		writeError(w, 400, "invalid json")
		return
	}

	// Make sure all fields are filled
	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeError(w, 400, "all fields required")
		return
	}

	// Make sure email is valid
	if !strings.Contains(req.Email, "@") {
		writeError(w, 400, "invalid email")
		return
	}

	// Make sure password is at least 6 chars
	if len(req.Password) < 6 {
		writeError(w, 400, "password min 6 chars")
		return
	}

	// Create variable 'exists'
	var exists int

	// Search the email
	h.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email=?",
		req.Email,
	).Scan(&exists)

	// If email already exists
	if exists > 0 {
		writeError(w, 409, "email already registered")
		return
	}

	// Hash the password
	hash, _ := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	// Insert the user
	res, err := h.DB.Exec(
		"INSERT INTO users(name,email,password) VALUES(?,?,?)",
		req.Name,
		req.Email,
		string(hash),
	)

	if err != nil {
		writeError(w, 500, "database error")
		return
	}

	// Get the user id
	id, _ := res.LastInsertId()

	// Generate JWT
	token := generateJWT(int(id), req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    id,
			"name":  req.Name,
			"email": req.Email,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req AuthRequest

	json.NewDecoder(r.Body).Decode(&req)

	var id int
	var name string
	var email string
	var password string

	err := h.DB.QueryRow(
		"SELECT id,name,email,password FROM users WHERE email=?",
		req.Email,
	).Scan(&id, &name, &email, &password)

	if err != nil {
		writeError(w, 401, "invalid credentials")
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(password),
		[]byte(req.Password),
	)

	if err != nil {
		writeError(w, 401, "invalid credentials")
		return
	}

	token := generateJWT(id, email)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    id,
			"name":  name,
			"email": email,
		},
	})
}

func generateJWT(id int, email string) string {

	claims := jwt.MapClaims{
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, _ := token.SignedString(middleware.JwtSecret)

	return signed
}

func writeError(w http.ResponseWriter, code int, msg string) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}