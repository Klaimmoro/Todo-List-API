package authentication

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	dberrors "todo-list-api/db/db_errors"
	"todo-list-api/token"

	"golang.org/x/crypto/bcrypt"
)

type RegistrationRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Authenticator struct {
	Db    *sql.DB
	Token *token.Token
}

func (authenticator *Authenticator) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unknown request", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var login_req LoginRequest
	err_parse := json.NewDecoder(r.Body).Decode(&login_req)
	if err_parse != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var user_hashed_pass string
	var user_id int
	err_find_user := authenticator.Db.QueryRow(
		`SELECT id, password FROM users WHERE email =$1`, login_req.Email,
	).Scan(&user_id, &user_hashed_pass)
	if err_find_user != nil {
		if errors.Is(err_find_user, sql.ErrNoRows) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err_password := bcrypt.CompareHashAndPassword([]byte(user_hashed_pass), []byte(login_req.Password))
	if err_password != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	token, err_token := authenticator.Token.GenerateToken(user_id)
	if err_token != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (authenticator *Authenticator) Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unknown request", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var registr_req RegistrationRequest
	err_parse := json.NewDecoder(r.Body).Decode(&registr_req)
	if err_parse != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	hashedPass, err_hashing := bcrypt.GenerateFromPassword([]byte(registr_req.Password), bcrypt.DefaultCost)
	if err_hashing != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	var userID int
	err_registration := authenticator.Db.QueryRow(
		`INSERT INTO users (name,email,password) VALUES ($1, $2, $3) RETURNING id`,
		registr_req.Name, registr_req.Email, string(hashedPass),
	).Scan(&userID)
	if err_registration != nil {
		if dberrors.IsUniqueViolation(err_registration) {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	token, err := authenticator.Token.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
