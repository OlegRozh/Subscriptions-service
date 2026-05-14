package storage

import (
	"context"
	"fmt"

	"github.com/OlegRozh/subscriptions-service/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) (int64, error)
	Get(ctx context.Context, id int64) (*models.Subscription, error)
	GetSum(ctx context.Context, userID, serviceName, start, end string) (int, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id int64) error
	Close() error
}

func NewStorage(databaseURL string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	storage := &Storage{pool: pool}

	return storage, nil
}

func (s *Storage) Create(ctx context.Context, subscription *models.Subscription) (int64, error) {
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_month, end_month)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`
	var id int64
	err := s.pool.QueryRow(ctx, query, subscription.ServiceName, subscription.Price, subscription.UserId, subscription.StartMonth, subscription.EndMonth).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create subscription: %w", err)
	}
	return id, nil
}

func (s *Storage) Get(ctx context.Context, id int64) (*models.Subscription, error) {
	var subscription models.Subscription
	query := `SELECT id, service_name, price, user_id, start_month, end_month, created_at
	FROM subscriptions
	WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&subscription.Id, &subscription.ServiceName, &subscription.Price,
		&subscription.UserId, &subscription.StartMonth, &subscription.EndMonth, &subscription.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (s *Storage) GetSum(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error) {
	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE user_id = $1
          AND ($2 = '' OR service_name = $2)
          AND ($3 = '' OR start_month >= $3::DATE)
          AND ($4 = '' OR (end_month IS NULL OR end_month <= $4::DATE))
    `
	var sum int
	err := s.pool.QueryRow(ctx, query, userID, serviceName, startDate, endDate).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("failed to sum subscriptions: %w", err)
	}
	return sum, nil
}

func (s *Storage) Update(ctx context.Context, subscription *models.Subscription) error {
	query := `UPDATE subscriptions
        SET service_name = $1, price = $2, end_month = $3
        WHERE id = $4
    `
	res, err := s.pool.Exec(ctx, query, subscription.ServiceName, subscription.Price, subscription.EndMonth, subscription.Id)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (s *Storage) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM subscriptions WHERE id = $1"
	res, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

func (s *Storage) Close() error {
	s.pool.Close()
	return nil
}
