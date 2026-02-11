package handlers

import "net/http"

func (a *App) AdminPing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, a.AdminS.Ping())
}
