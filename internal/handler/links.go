package handler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nino-Prog/shortr/internal/geo"
	"github.com/Nino-Prog/shortr/internal/model"
	"github.com/go-chi/chi/v5"
)

type shortenRequest struct {
	URL       string  `json:"url"`
	Code      string  `json:"code,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r.Context())
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		http.Error(w, "url must start with http:// or https://", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "invalid expires_at format (use RFC3339)", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	link, err := h.store.CreateLink(r.Context(), userID, req.URL, req.Code, expiresAt)
	if err != nil {
		http.Error(w, "could not create link (code may be taken)", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	link, err := h.store.GetLinkByCode(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		http.Error(w, "link expired", http.StatusGone)
		return
	}

	ip := realIP(r)
	go func() {
		loc := geo.Lookup(ip)
		h.store.RecordClick(r.Context(), link.ID, loc.Country, loc.City, hashIP(ip))
	}()

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r.Context())
	links, err := h.store.ListLinksByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if links == nil {
		links = []model.Link{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r.Context())
	code := chi.URLParam(r, "code")
	if err := h.store.DeleteLink(r.Context(), code, userID); err != nil {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r.Context())
	code := chi.URLParam(r, "code")
	analytics, err := h.store.GetAnalytics(r.Context(), code, userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.SplitN(ip, ",", 2)[0]
	}
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}

func hashIP(ip string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(ip)))[:16]
}
