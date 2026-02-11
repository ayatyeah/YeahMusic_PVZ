package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (a *App) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	r.ParseMultipartForm(10 << 20)
	title := r.FormValue("title")
	if title == "" {
		writeErr(w, 400, "title required")
		return
	}

	coverURL := ""
	cFile, cHeader, err := r.FormFile("cover")
	if err == nil {
		defer cFile.Close()
		cExt := filepath.Ext(cHeader.Filename)
		cFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "playlist", cExt)
		cDst, _ := os.Create(filepath.Join("public", "uploads", cFilename))
		io.Copy(cDst, cFile)
		cDst.Close()
		coverURL = "/uploads/" + cFilename
	}

	writeJSON(w, 201, a.PlayS.Create(u.ID, title, coverURL))
}

func (a *App) ListPlaylists(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	writeJSON(w, 200, a.PlayS.List(u.ID))
}

type addTrackReq struct {
	TrackID string `json:"track_id"`
}

func (a *App) PlaylistSubroutes(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	path := strings.TrimPrefix(r.URL.Path, "/api/playlists/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		writeErr(w, 404, "not found")
		return
	}
	playlistID, err := primitive.ObjectIDFromHex(parts[0])
	if err != nil {
		writeErr(w, 400, "invalid id")
		return
	}

	if r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "tracks" {
		var req addTrackReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, 400, "bad json")
			return
		}
		trackID, err := primitive.ObjectIDFromHex(req.TrackID)
		if err != nil {
			writeErr(w, 400, "invalid track id")
			return
		}

		if err := a.PlayS.AddTrack(u.ID, playlistID, trackID); err != nil {
			writeErr(w, 500, "error")
			return
		}
		writeJSON(w, 200, map[string]any{"status": "added"})
		return
	}
	if r.Method == http.MethodDelete && len(parts) == 3 && parts[1] == "tracks" {
		trackID, err := primitive.ObjectIDFromHex(parts[2])
		if err != nil {
			writeErr(w, 400, "invalid track id")
			return
		}

		if err := a.PlayS.RemoveTrack(u.ID, playlistID, trackID); err != nil {
			writeErr(w, 500, "error")
			return
		}
		writeJSON(w, 200, map[string]any{"status": "removed"})
		return
	}
	writeErr(w, 404, "not found")
}
