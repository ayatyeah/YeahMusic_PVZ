package handlers

import (
	"YeahMusic/internal/services"
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"YeahMusic/internal/models"
)

type ctxKey string

const ctxUserKey ctxKey = "user"

func (a *App) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, http.StatusUnauthorized, "missing token")
			return
		}
		tok := strings.TrimPrefix(h, "Bearer ")
		sec, err := a.Store.GetSession(tok)
		if err != nil || !sec.ExpiresAt.After(time.Now()) {
			writeErr(w, http.StatusUnauthorized, "expired session")
			return
		}
		u, err := a.Store.GetUserByID(sec.UserID)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "user not found")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userFromCtx(r *http.Request) *models.User {
	u, _ := r.Context().Value(ctxUserKey).(*models.User)
	return u
}

func mapServiceErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, services.ErrNotFound):
		writeErr(w, http.StatusNotFound, "not found")
		return true
	case errors.Is(err, services.ErrForbidden):
		writeErr(w, http.StatusForbidden, "forbidden")
		return true
	case errors.Is(err, services.ErrAlreadyExists):
		writeErr(w, http.StatusConflict, "already exists")
		return true
	case errors.Is(err, services.ErrUnauthorized):
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return true
	default:
		writeErr(w, http.StatusInternalServerError, "server error")
		return true
	}
}
