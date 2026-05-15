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
	GetList(ctx context.Context, userId, serviceName string) ([]models.Subscription, error)
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
	if startDate == "" && endDate == "" {
		query := `
            SELECT COALESCE(SUM(price), 0)
            FROM subscriptions
            WHERE user_id = $1
              AND ($2 = '' OR service_name = $2)
        `
		var sum int
		err := s.pool.QueryRow(ctx, query, userID, serviceName).Scan(&sum)
		return sum, err
	}
	query := `
        SELECT COALESCE(SUM(price * months_count), 0)
        FROM (
            SELECT
                s.price,
                COUNT(*) as months_count
            FROM subscriptions s
            CROSS JOIN LATERAL generate_series(
                date_trunc('month', GREATEST(s.start_month, $3::DATE)),
                date_trunc('month', LEAST(COALESCE(s.end_month, $4::DATE), $4::DATE)),
                '1 month'
            ) AS month_series
            WHERE s.user_id = $1
              AND ($2 = '' OR s.service_name = $2)
              AND s.start_month <= $4::DATE
              AND (s.end_month IS NULL OR s.end_month >= $3::DATE)
            GROUP BY s.id, s.price
        ) AS active
    `
	var sum int
	err := s.pool.QueryRow(ctx, query, userID, serviceName, startDate, endDate).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("failed to sum subscriptions: %w", err)
	}
	return sum, nil
}

func (s *Storage) GetList(ctx context.Context, userId, serviceName string) ([]models.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_month, end_month, created_at
        FROM subscriptions
        WHERE ($1 = '' OR user_id = $1)
          AND ($2 = '' OR service_name = $2)
        ORDER BY start_month DESC
    `
	rows, err := s.pool.Query(ctx, query, userId, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(&sub.Id, &sub.ServiceName, &sub.Price,
			&sub.UserId, &sub.StartMonth, &sub.EndMonth, &sub.CreatedAt)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *Storage) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $1, price = $2, user_id = $3, start_month = $4, end_month = $5
        WHERE id = $6
    `
	_, err := s.pool.Exec(ctx, query,
		sub.ServiceName, sub.Price, sub.UserId, sub.StartMonth, sub.EndMonth, sub.Id)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
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
