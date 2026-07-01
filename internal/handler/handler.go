package handler

import "github.com/Nino-Prog/shortr/internal/store"

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}
