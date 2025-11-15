package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/reco-engine/internal/models"
	"github.com/yourusername/reco-engine/internal/util/config"
	"github.com/yourusername/reco-engine/internal/util/logger"
	"go.uber.org/zap"
)

// PostgresStore handles PostgreSQL operations
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a new Postgres store
func NewPostgresStore(cfg config.PostgresConfig) (*PostgresStore, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode, cfg.MaxOpenConns,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))

	return &PostgresStore{pool: pool}, nil
}

// Close closes the database connection pool
func (p *PostgresStore) Close() {
	p.pool.Close()
}

// InsertEvent inserts an event into the database
func (p *PostgresStore) InsertEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (user_id, item_id, event_type, session_id, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return p.pool.QueryRow(ctx, query,
		event.UserID,
		event.ItemID,
		event.EventType,
		event.SessionID,
		event.Metadata,
		event.Timestamp,
	).Scan(&event.ID)
}

// GetItem retrieves an item by ID
func (p *PostgresStore) GetItem(ctx context.Context, itemID int64) (*models.Item, error) {
	query := `
		SELECT id, sku, title, category, price, stock, metadata, created_at, updated_at
		FROM items
		WHERE id = $1
	`
	var item models.Item
	err := p.pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID,
		&item.SKU,
		&item.Title,
		&item.Category,
		&item.Price,
		&item.Stock,
		&item.Metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetItems retrieves multiple items by IDs
func (p *PostgresStore) GetItems(ctx context.Context, itemIDs []int64) ([]models.Item, error) {
	query := `
		SELECT id, sku, title, category, price, stock, metadata, created_at, updated_at
		FROM items
		WHERE id = ANY($1)
	`
	rows, err := p.pool.Query(ctx, query, itemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.SKU,
			&item.Title,
			&item.Category,
			&item.Price,
			&item.Stock,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetItemsByCategory retrieves items by category
func (p *PostgresStore) GetItemsByCategory(ctx context.Context, category string, limit int) ([]models.Item, error) {
	query := `
		SELECT id, sku, title, category, price, stock, metadata, created_at, updated_at
		FROM items
		WHERE category = $1
		LIMIT $2
	`
	rows, err := p.pool.Query(ctx, query, category, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.SKU,
			&item.Title,
			&item.Category,
			&item.Price,
			&item.Stock,
			&item.Metadata,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// InsertModel inserts a new model metadata
func (p *PostgresStore) InsertModel(ctx context.Context, model *models.Model) error {
	query := `
		INSERT INTO models (model_name, version, model_type, metrics, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return p.pool.QueryRow(ctx, query,
		model.ModelName,
		model.Version,
		model.ModelType,
		model.Metrics,
		model.Config,
	).Scan(&model.ID, &model.CreatedAt)
}
