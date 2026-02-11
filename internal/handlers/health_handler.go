package handlers

import "net/http"

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok"})
}
