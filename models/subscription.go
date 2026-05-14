package models

import "time"

type Subscription struct {
	Id          int64      `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserId      string     `json:"user_id"`
	StartMonth  time.Time  `json:"start_month"`
	EndMonth    *time.Time `json:"end_month,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
