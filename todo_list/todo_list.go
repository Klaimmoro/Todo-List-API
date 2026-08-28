package todo_list

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"todo-list-api/token"
)

type ToDoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ToDoResponse struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetToDoResponse struct {
	Data  []ToDoResponse `json:"data"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Total int            `json:"total"`
}

type ToDoList struct {
	DB    *sql.DB
	Token *token.Token
}

func (tdl *ToDoList) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unknown error", http.StatusMethodNotAllowed)
		return
	}
	auth_header := r.Header.Get("Authorization")
	auth_token := strings.TrimPrefix(auth_header, "Bearer ")
	user_id, err := tdl.Token.ParseToken(auth_token)
	if auth_header == auth_token || err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
		return
	}
	defer r.Body.Close()
	var new_todo ToDoRequest
	err_parse := json.NewDecoder(r.Body).Decode(&new_todo)
	if err_parse != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if new_todo.Title == "" || new_todo.Description == "" {
		http.Error(w, "Invalid values for title or description", http.StatusBadRequest)
		return
	}
	var task_id int
	err_create := tdl.DB.QueryRow("INSERT INTO tasks (user_id, title, description) VALUES ($1,$2, $3) RETURNING id", user_id, new_todo.Title, new_todo.Description).Scan(&task_id)
	if err_create != nil {
		log.Println("Error to insert new task: ", err_create)
		http.Error(w, "Error to insert new task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToDoResponse{ID: task_id, Title: new_todo.Title, Description: new_todo.Description})
}

func (tdl *ToDoList) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Unknown error", http.StatusMethodNotAllowed)
		return
	}
	auth_header := r.Header.Get("Authorization")
	auth_token := strings.TrimPrefix(auth_header, "Bearer ")
	user_id, err := tdl.Token.ParseToken(auth_token)
	if auth_header == auth_token || err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
		return
	}
	task_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
	}
	defer r.Body.Close()
	var new_todo ToDoRequest
	err_parse := json.NewDecoder(r.Body).Decode(&new_todo)
	if err_parse != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if new_todo.Title == "" || new_todo.Description == "" {
		http.Error(w, "Invalid values for title or description", http.StatusBadRequest)
		return
	}
	var owner_id int
	err_finding := tdl.DB.QueryRow("SELECT user_id FROM tasks WHERE id=$1", task_id).Scan(&owner_id)
	if errors.Is(err_finding, sql.ErrNoRows) {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err_finding != nil {
		http.Error(w, "Error to insert new task", http.StatusInternalServerError)
		return
	}
	if owner_id != user_id {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
		return
	}
	_, err_update := tdl.DB.Exec("UPDATE tasks SET title=$1, description=$2 WHERE id=$3", new_todo.Title, new_todo.Description, task_id)
	if err_update != nil {
		log.Println("Error to insert new task: ", err_update)
		http.Error(w, "Error to update task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ToDoResponse{ID: task_id, Title: new_todo.Title, Description: new_todo.Description})
}

func (tdl *ToDoList) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Unknown error", http.StatusMethodNotAllowed)
		return
	}
	auth_header := r.Header.Get("Authorization")
	auth_token := strings.TrimPrefix(auth_header, "Bearer ")
	user_id, err := tdl.Token.ParseToken(auth_token)
	if auth_header == auth_token || err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
		return
	}
	task_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var owner_id int
	err_finding := tdl.DB.QueryRow("SELECT user_id FROM tasks WHERE id=$1", task_id).Scan(&owner_id)
	if errors.Is(err_finding, sql.ErrNoRows) {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err_finding != nil {
		http.Error(w, "Error to delete new task", http.StatusInternalServerError)
		return
	}
	if owner_id != user_id {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "Forbidden"})
		return
	}
	_, err_update := tdl.DB.Exec("DELETE FROM tasks WHERE id=$1", task_id)
	if err_update != nil {
		log.Println("Error to delete new task: ", err_update)
		http.Error(w, "Error to update task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(204)
}

func (tdl *ToDoList) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unknown error", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	auth_header := r.Header.Get("Authorization")
	auth_token := strings.TrimPrefix(auth_header, "Bearer ")
	user_id, err := tdl.Token.ParseToken(auth_token)
	if err != nil {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
		return
	}
	rows, err := tdl.DB.Query("SELECT id, title, description FROM tasks WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3", user_id, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tasks are not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Error to get tasks: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var todos []ToDoResponse
	for rows.Next() {
		var todo ToDoResponse
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Description); err != nil {
			log.Println("Error to scan row: ", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error to scan row: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetToDoResponse{
		Data:  todos,
		Page:  page,
		Limit: limit,
		Total: len(todos),
	})
}
