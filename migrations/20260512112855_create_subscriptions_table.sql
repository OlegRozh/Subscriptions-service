-- +goose Up
CREATE TABLE subscriptions (
                               id            BIGSERIAL PRIMARY KEY,
                               service_name  TEXT NOT NULL,
                               price         INT NOT NULL CHECK (price >= 0),
                               user_id       UUID NOT NULL,
                               start_month   DATE NOT NULL,
                               end_month     DATE,
                               created_at    TIMESTAMP DEFAULT NOW()
);
-- +goose Down
