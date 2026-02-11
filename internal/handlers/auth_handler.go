package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"YeahMusic/internal/services"
)

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "bad json")
		return
	}
	role := strings.TrimSpace(req.Role)
	if role != "artist" {
		role = "user"
	}

	u, err := a.AuthS.Register(req.Email, req.Password, req.Name, role)
	if err != nil {
		if err == services.ErrAlreadyExists {
			writeErr(w, 409, "email exists")
			return
		}
		writeErr(w, 500, "server error")
		return
	}
	writeJSON(w, 201, u)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "bad json")
		return
	}
	sec, u, err := a.AuthS.Login(req.Email, req.Password)
	if err != nil {
		writeErr(w, 401, "invalid credentials")
		return
	}
	writeJSON(w, 200, map[string]any{
		"token": sec.Token, "expires_at": sec.ExpiresAt, "user": u,
	})
}
