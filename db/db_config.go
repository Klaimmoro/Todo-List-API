package db

import (
	"database/sql"
	"fmt"
	"os"
)

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
}

func (db_conf *DBConfig) SetConfig() {
	db_conf.Host = os.Getenv("DB_HOST")
	db_conf.Port = os.Getenv("DB_PORT")
	db_conf.User = os.Getenv("DB_USER")
	db_conf.Password = os.Getenv("DB_PASSWORD")
	db_conf.Name = os.Getenv("DB_NAME")
}

func (db_conf *DBConfig) InitTables(db *sql.DB) {
	init_users := `CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL
	)`
	init_tasks := `CREATE TABLE IF NOT EXISTS tasks(
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id),
		title VARCHAR(255) NOT NULL,
		description VARCHAR(255) NOT NULL
	)`
	_, err_users := db.Exec(init_users)
	if err_users != nil {
		panic(fmt.Sprintf("Error to initialize users table: %s", err_users))
	}
	_, err_tasks := db.Exec(init_tasks)
	if err_tasks != nil {
		panic(fmt.Sprintf("Error to initialize users table: %s", err_tasks))
	}
}
