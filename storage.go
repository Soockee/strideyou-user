package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	CreateUser(*User) (string, error)
	GetUser(string) (*User, error)
	GetUsers() ([]*User, error)
	UpdateUser(string, []*User) (*User, error)
	DeleteUser(string) (*User, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "host=postgres-user user=postgres dbname=postgres password=strideyouuserdb sslmode=disable"
	//debug locally
	//connStr := "user=postgres dbname=postgres password=golifeUserer sslmode=disable"

	var db *sql.DB
	var err error
	maxRetries := 5
	retryInterval := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			break
		}

		log.Error().Msgf("Failed to connect to the database: %v. Retrying in %v...\n", err, retryInterval)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database after %d retries: %v", maxRetries, err)
	}

	// Ping the database to ensure the connection is successful
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err == nil {
			break
		}

		log.Error().Msgf("Failed to ping the database. Retrying in %v...\n", retryInterval)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to ping the database after %d retries: %v", maxRetries, err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	var err error

	err = s.createUserTable()
	if err != nil {
		return err

	}
	return nil
}

func (s *PostgresStore) createUserTable() error {
	query := `	
		CREATE TABLE IF NOT EXISTS "users" (
			sub VARCHAR(255) PRIMARY KEY
		);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) CreateUser(u *User) (string, error) {
	query := `
		INSERT INTO "users" (sub)
		VALUES($1);
	`
	_, err := s.db.Exec(query, u.Sub)
	if err != nil {
		return "", err
	}
	return u.Sub, nil
}
func (s *PostgresStore) GetUser(sub string) (*User, error) {
	query := `
		SELECT sub 
		FROM "users"
		WHERE sub = $1;
	`
	rows, err := s.db.Query(query, sub)
	log.Debug().Msg("Executed Query")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	log.Debug().Msg("Executed Query")
	user := &User{}
	if err := rows.Scan(&user.Sub); err != nil {
		return nil, err
	}

	return user, nil
}
func (s *PostgresStore) UpdateUser(string, []*User) (*User, error) {
	return nil, nil
}
func (s *PostgresStore) DeleteUser(sub string) (*User, error) {
	user, err := s.GetUser(sub)
	if err != nil {
		return nil, err
	}

	query := `
		DELETE FROM "users"
		WHERE sub = $1
	`
	_, err = s.db.Exec(query, sub)
	if err != nil {
		return nil, err
	}

	return user, nil
}
func (s *PostgresStore) GetUsers() ([]*User, error) {
	userQuery := `
		SELECT sub
		FROM "users"
	`

	rows, err := s.db.Query(userQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.Sub); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
