package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/Nino-Prog/shortr/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrCodeConflict = errors.New("code already taken")

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Users

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// Links

func (s *Store) CreateLink(ctx context.Context, userID int64, originalURL, code string, expiresAt *time.Time) (*model.Link, error) {
	if code == "" {
		code = randomCode(7)
	}
	l := &model.Link{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO links (user_id, code, original_url, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, code, original_url, expires_at, created_at`,
		userID, code, originalURL, expiresAt,
	).Scan(&l.ID, &l.UserID, &l.Code, &l.OriginalURL, &l.ExpiresAt, &l.CreatedAt)
	return l, err
}

func (s *Store) GetLinkByCode(ctx context.Context, code string) (*model.Link, error) {
	l := &model.Link{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, code, original_url, expires_at, created_at
		 FROM links WHERE code = $1`, code,
	).Scan(&l.ID, &l.UserID, &l.Code, &l.OriginalURL, &l.ExpiresAt, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return l, err
}

func (s *Store) ListLinksByUser(ctx context.Context, userID int64) ([]model.Link, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, code, original_url, expires_at, created_at
		 FROM links WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var l model.Link
		if err := rows.Scan(&l.ID, &l.UserID, &l.Code, &l.OriginalURL, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *Store) DeleteLink(ctx context.Context, code string, userID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM links WHERE code = $1 AND user_id = $2`, code, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Clicks

func (s *Store) RecordClick(ctx context.Context, linkID int64, country, city, ipHash string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO clicks (link_id, country, city, ip_hash) VALUES ($1, $2, $3, $4)`,
		linkID, country, city, ipHash,
	)
	return err
}

func (s *Store) GetAnalytics(ctx context.Context, code string, userID int64) (*model.Analytics, error) {
	l := &model.Link{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, code, original_url, expires_at, created_at
		 FROM links WHERE code = $1 AND user_id = $2`, code, userID,
	).Scan(&l.ID, &l.UserID, &l.Code, &l.OriginalURL, &l.ExpiresAt, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var total int64
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM clicks WHERE link_id = $1`, l.ID).Scan(&total)

	rows, _ := s.db.Query(ctx,
		`SELECT country, city, clicked_at, ip_hash FROM clicks WHERE link_id = $1 ORDER BY clicked_at DESC LIMIT 100`, l.ID,
	)
	defer rows.Close()

	byCountry := map[string]int64{}
	byCity := map[string]int64{}
	var recent []model.Click
	for rows.Next() {
		var c model.Click
		c.LinkID = l.ID
		rows.Scan(&c.Country, &c.City, &c.ClickedAt, &c.IPHash)
		byCountry[c.Country]++
		byCity[c.City]++
		if len(recent) < 20 {
			recent = append(recent, c)
		}
	}

	return &model.Analytics{
		Link:        *l,
		TotalClicks: total,
		ByCountry:   byCountry,
		ByCity:      byCity,
		Recent:      recent,
	}, nil
}

func randomCode(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}
