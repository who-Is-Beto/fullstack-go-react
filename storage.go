package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateUser(user *User) error
	GetUser(id int) (*User, error)
	GetUsers() ([]*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}

type PostgresStorage struct {
	db *sql.DB
}

func (s *PostgresStorage) CreateUserTable() error {
	query := `create table if not exists users (
			id serial primary key,
			username varchar(50),
			createdAt timestamp,
			updatedAt timestamp
		)
	`

		_, err := s.db.Exec(query)
		return err
}

func (s *PostgresStorage) Start() error {
	return s.CreateUserTable()
}



func (s *PostgresStorage) CreateUser(user *User) error {
	query := `insert into users (username, createdAt, updatedAt)
							values ($1, $2, $3)`

	response, err := s.db.Query(query, user.Username, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	fmt.Printf("response: %v\n", response)

	return nil
}

func (s *PostgresStorage) GetUser(id int) (*User, error) {
	rows, err := s.db.Query("select * from users where id = $1", id)
	if err != nil {
		return nil, err
	}

  for rows.Next() {
		return scanIntoUser(rows)
	}

	return nil, fmt.Errorf("user with id %d not found", id)
}

func (s *PostgresStorage) UpdateUser(user *User) error {
	return nil
}

func (s *PostgresStorage) DeleteUser(id int) error {
	_, err := s.db.Query("delete from users where id = $1", id)
	return err
}

func (s *PostgresStorage) GetUsers() ([]*User, error) {
	rows, err := s.db.Query("select * from users")
	
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for rows.Next() {
		user, err := scanIntoUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	return users, nil
}

func NewPostgresStorage() (*PostgresStorage, error) {
	// connectionString := fmt.Sprintf("user=%s password=admin123 dbname=marketplace port=5432 sslmode=disable", os.Getenv("DATABASE_USERNAME"))
	connectionString := "user=whoisbeto password=admin123 dbname=marketplace port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func scanIntoUser(rows *sql.Rows) (*User, error) {
	user := new(User)
	err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}