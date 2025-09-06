package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mmoviecom/rating/configs"
	"mmoviecom/rating/pkg/model"

	_ "github.com/go-sql-driver/mysql"
)

// Repository defines a MySQL-based rating repository.
type Repository struct {
	db *sql.DB
}

// New creates a new MySQL-based rating repository.
func New(config configs.MysqlConfig) (*Repository, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name))
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, recordId model.RecordId, recordType model.RecordType) ([]model.Rating, error) {
	log.Printf("Trying to get rating from MySQL for record: %s/%s", recordType, recordId)
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?", recordId, recordType)
	if err != nil {
		log.Printf("Failed to get rating from MySQL for record: %s/%s. Error: %v", recordType, recordId, err)
		return nil, err
	}
	defer rows.Close()
	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int32
		if err := rows.Scan(&userID, &value); err != nil {
			log.Printf("Failed to get rating items from MySQL for record: %s/%s. Error: %v", recordType, recordId, err)
			return nil, err
		}
		res = append(res, model.Rating{
			UserId: model.UserId(userID),
			Value:  model.RatingValue(value),
		})
	}
	return res, nil
}

// Put adds a rating for a given record.
func (r *Repository) Put(ctx context.Context, recordId model.RecordId, recordType model.RecordType, rating *model.Rating) error {
	if rating == nil {
		return errors.New("rating is nil")
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?, ?, ?, ?)",
		recordId, recordType, rating.UserId, rating.Value)
	return err
}
