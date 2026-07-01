package model

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Link struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Code        string     `json:"code"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Click struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	ClickedAt time.Time `json:"clicked_at"`
	Country   string    `json:"country"`
	City      string    `json:"city"`
	IPHash    string    `json:"ip_hash"`
}

type Analytics struct {
	Link       Link            `json:"link"`
	TotalClicks int64          `json:"total_clicks"`
	ByCountry  map[string]int64 `json:"by_country"`
	ByCity     map[string]int64 `json:"by_city"`
	Recent     []Click         `json:"recent"`
}
