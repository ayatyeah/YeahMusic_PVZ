package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"YeahMusic/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (a *App) ListArtists(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, a.CatalogS.ListArtists())
}

func (a *App) ListAlbums(w http.ResponseWriter, r *http.Request) {
	var id primitive.ObjectID
	if q := r.URL.Query().Get("artist_id"); q != "" {
		id, _ = primitive.ObjectIDFromHex(q)
	}
	writeJSON(w, 200, a.CatalogS.ListAlbums(id))
}

func (a *App) ListTracks(w http.ResponseWriter, r *http.Request) {
	var id primitive.ObjectID
	if q := r.URL.Query().Get("album_id"); q != "" {
		id, _ = primitive.ObjectIDFromHex(q)
	}
	writeJSON(w, 200, a.CatalogS.ListTracks(id))
}

func (a *App) UploadTrack(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	if title == "" {
		writeErr(w, 400, "title required")
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		writeErr(w, 400, "audio required")
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "audio", ext)
	dst, _ := os.Create(filepath.Join("public", "uploads", filename))
	io.Copy(dst, file)
	dst.Close()
	audioURL := "/uploads/" + filename

	coverURL := ""
	cFile, cHeader, err := r.FormFile("cover")
	if err == nil {
		defer cFile.Close()
		cExt := filepath.Ext(cHeader.Filename)
		cFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "cover", cExt)
		cDst, _ := os.Create(filepath.Join("public", "uploads", cFilename))
		io.Copy(cDst, cFile)
		cDst.Close()
		coverURL = "/uploads/" + cFilename
	}

	lyricsContent := ""
	lFile, _, err := r.FormFile("lyrics")
	if err == nil {
		defer lFile.Close()
		bytes, _ := io.ReadAll(lFile)
		lyricsContent = string(bytes)
	}

	track := &models.Track{Title: title, ArtistName: artist, AudioURL: audioURL, CoverURL: coverURL, Lyrics: lyricsContent}
	writeJSON(w, 201, a.CatalogS.AddTrack(track))
}

func (a *App) UpdateTrackHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeErr(w, 400, "bad path")
		return
	}
	id, err := primitive.ObjectIDFromHex(parts[3])
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}

	r.ParseMultipartForm(10 << 20)
	coverURL := ""
	cFile, cHeader, err := r.FormFile("cover")
	if err == nil {
		defer cFile.Close()
		cExt := filepath.Ext(cHeader.Filename)
		cFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "cover_upd", cExt)
		cDst, _ := os.Create(filepath.Join("public", "uploads", cFilename))
		io.Copy(cDst, cFile)
		cDst.Close()
		coverURL = "/uploads/" + cFilename
	}

	lyricsContent := ""
	lFile, _, err := r.FormFile("lyrics")
	if err == nil {
		defer lFile.Close()
		bytes, _ := io.ReadAll(lFile)
		lyricsContent = string(bytes)
	}

	updated, err := a.CatalogS.UpdateTrack(id, coverURL, lyricsContent)
	if err != nil {
		writeErr(w, 404, "track not found")
		return
	}
	writeJSON(w, 200, updated)
}

func (a *App) DeleteTrackHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeErr(w, 400, "bad path")
		return
	}
	id, err := primitive.ObjectIDFromHex(parts[3])
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}

	if err := a.CatalogS.DeleteTrack(id); err != nil {
		writeErr(w, 404, "track not found")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (a *App) CreateAlbumHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	title := r.FormValue("title")
	if title == "" {
		writeErr(w, 400, "title required")
		return
	}
	year, _ := strconv.Atoi(r.FormValue("release_year"))

	coverURL := ""
	cFile, cHeader, err := r.FormFile("cover")
	if err == nil {
		defer cFile.Close()
		cExt := filepath.Ext(cHeader.Filename)
		cFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "album", cExt)
		cDst, _ := os.Create(filepath.Join("public", "uploads", cFilename))
		io.Copy(cDst, cFile)
		cDst.Close()
		coverURL = "/uploads/" + cFilename
	}

	album := &models.Album{Title: title, ReleaseYear: year, CoverURL: coverURL}
	writeJSON(w, 201, a.CatalogS.CreateAlbum(album))
}
