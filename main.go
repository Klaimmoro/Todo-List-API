package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"todo-list-api/authentication"
	"todo-list-api/db"
	"todo-list-api/todo_list"
	"todo-list-api/token"

	_ "github.com/lib/pq"
	"github.com/lpernett/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error to load `.env` file with config for DB connection")
	}
	token := &token.Token{
		JwtSecret: []byte(os.Getenv("JWT_SECRET")),
	}
	var db_config db.DBConfig
	db_config.SetConfig()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", db_config.Host, db_config.Port, db_config.User, db_config.Password, db_config.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("Error to open DB with dsn: %s, error: %s", dsn, err))
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic("Error to connect to DB")
	}
	db_config.InitTables(db)
	authenticator := &authentication.Authenticator{
		Db:    db,
		Token: token,
	}
	http.HandleFunc("/register", authenticator.Registration)
	http.HandleFunc("/login", authenticator.Login)
	todo_list := todo_list.ToDoList{
		DB:    db,
		Token: token,
	}
	http.HandleFunc("POST /todos", todo_list.Create)
	http.HandleFunc("PUT /todos/{id}", todo_list.Update)
	http.HandleFunc("DELETE /todos/{id}", todo_list.Delete)
	http.HandleFunc("GET /todos", todo_list.Get)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic("Error to set connection")
	}
}
