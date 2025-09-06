package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mmoviecom/metadata/configs"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"

	_ "github.com/go-sql-driver/mysql"
)

// Repository defines a MySQL-based movie metadata repository.
type Repository struct {
	db *sql.DB
}

// New creates a new MySQL-based repository.
func New(config configs.MysqlConfig) (*Repository, error) {
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name)
	log.Printf("connecting to mysql %s", connString)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// Get retrieves movie metadata by a movie id.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	var title, description, director string
	log.Println("Trying to get metadata from MySQL by id: ", id)
	row := r.db.QueryRowContext(ctx, "SELECT title, description, director FROM movies WHERE id=?", id)
	if err := row.Scan(&title, &description, &director); err != nil {
		log.Printf("Failed to get metadata from MySQL by id: %s. Error: %v", id, err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &model.Metadata{
		ID:          id,
		Title:       title,
		Description: description,
		Director:    director,
	}, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(ctx context.Context, id string, m *model.Metadata) error {
	log.Println("Trying to put metadata to MySQL by id: ", id)
	_, err := r.db.ExecContext(ctx, "INSERT INTO movies (id, title, description, director) VALUES (?, ?, ?, ?)",
		id, m.Title, m.Description, m.Director)
	if err != nil {
		log.Printf("Failed to get metadata ti MySQL by id: %s. Error: %v", id, err)
	}
	return err
}
