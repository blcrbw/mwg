package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mmoviecom/pkg/logging"
	"mmoviecom/rating/configs"
	"mmoviecom/rating/pkg/model"

	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const tracerID = "metadata-repository-mysql"

// Repository defines a MySQL-based rating repository.
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// New creates a new MySQL-based rating repository.
func New(config configs.MysqlConfig, logger *zap.Logger) (*Repository, error) {
	logger = logger.With(
		zap.String(logging.FieldComponent, "repository"),
		zap.String(logging.FieldType, "mysql"),
	)
	logger.Info("Connecting to mysql")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name))
	if err != nil {
		return nil, err
	}
	return &Repository{db: db, logger: logger}, nil
}

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, recordId model.RecordId, recordType model.RecordType) ([]model.Rating, error) {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Get")
	defer span.End()
	r.logger.Info("Trying to get rating from MySQL", zap.String("record", fmt.Sprintf("%v/%v", recordType, recordId)))
	rows, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?", recordId, recordType)
	if err != nil {
		r.logger.Warn("Failed to get rating from MySQL", zap.String("record", fmt.Sprintf("%v/%v", recordType, recordId)), zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int32
		if err := rows.Scan(&userID, &value); err != nil {
			r.logger.Warn("Failed to get rating items from MySQL", zap.String("record", fmt.Sprintf("%v/%v", recordType, recordId)), zap.Error(err))
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
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Put")
	defer span.End()
	if rating == nil {
		return errors.New("rating is nil")
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?, ?, ?, ?)",
		recordId, recordType, rating.UserId, rating.Value)
	return err
}
