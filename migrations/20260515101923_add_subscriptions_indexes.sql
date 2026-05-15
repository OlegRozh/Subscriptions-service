-- +goose Up
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX idx_subscriptions_start_month ON subscriptions(start_month);
CREATE INDEX idx_subscriptions_end_month ON subscriptions(end_month);

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP INDEX IF EXISTS idx_subscriptions_service_name;
DROP INDEX IF EXISTS idx_subscriptions_start_month;
DROP INDEX IF EXISTS idx_subscriptions_end_month;
