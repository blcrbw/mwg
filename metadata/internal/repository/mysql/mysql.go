package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mmoviecom/metadata/configs"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/logging"

	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const tracerID = "metadata-repository-mysql"

// Repository defines a MySQL-based movie metadata repository.
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// New creates a new MySQL-based repository.
func New(config configs.MysqlConfig, logger *zap.Logger) (*Repository, error) {
	logger = logger.With(
		zap.String(logging.FieldComponent, "repository"),
		zap.String(logging.FieldType, "mysql"),
	)
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Pass, config.Host, config.Port, config.Name)
	logger.Info("Connecting to mysql")
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return nil, err
	}
	return &Repository{db: db, logger: logger}, nil
}

// Get retrieves movie metadata by a movie id.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Get")
	defer span.End()
	var title, description, director string
	r.logger.Info("Trying to get metadata from MySQL", zap.String("id", id))
	row := r.db.QueryRowContext(ctx, "SELECT title, description, director FROM movies WHERE id=?", id)
	if err := row.Scan(&title, &description, &director); err != nil {
		r.logger.Warn("Failed to get metadata from MySQL", zap.String("id", id), zap.Error(err))
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
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Put")
	defer span.End()
	r.logger.Info("Trying to put metadata to MySQL", zap.String("id", id))
	_, err := r.db.ExecContext(ctx, "INSERT INTO movies (id, title, description, director) VALUES (?, ?, ?, ?)",
		id, m.Title, m.Description, m.Director)
	if err != nil {
		r.logger.Warn("Failed to get metadata ti MySQL", zap.String("id", id), zap.Error(err))
	}
	return err
}
