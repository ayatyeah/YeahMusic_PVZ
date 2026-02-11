package handlers

import (
	"net/http"
)

func NewRouter(app *App) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.Health)

	mux.HandleFunc("POST /api/register", app.Register)
	mux.HandleFunc("POST /api/login", app.Login)

	mux.HandleFunc("GET /api/artists", app.ListArtists)
	mux.HandleFunc("GET /api/albums", app.ListAlbums)
	mux.HandleFunc("GET /api/tracks", app.ListTracks)

	mux.Handle("POST /api/upload", app.Auth(http.HandlerFunc(app.UploadTrack)))
	mux.Handle("POST /api/tracks/", app.Auth(http.HandlerFunc(app.UpdateTrackHandler)))
	mux.Handle("DELETE /api/tracks/", app.Auth(http.HandlerFunc(app.DeleteTrackHandler)))

	mux.Handle("POST /api/albums", app.Auth(http.HandlerFunc(app.CreateAlbumHandler)))

	mux.Handle("POST /api/playlists", app.Auth(http.HandlerFunc(app.CreatePlaylist)))
	mux.Handle("GET /api/playlists", app.Auth(http.HandlerFunc(app.ListPlaylists)))
	mux.Handle("POST /api/playlists/", app.Auth(http.HandlerFunc(app.PlaylistSubroutes)))
	mux.Handle("DELETE /api/playlists/", app.Auth(http.HandlerFunc(app.PlaylistSubroutes)))

	mux.HandleFunc("GET /api/admin/ping", app.AdminPing)

	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("GET /", http.StripPrefix("/", fs))
	mux.Handle("GET /css/", http.StripPrefix("/", fs))
	mux.Handle("GET /js/", http.StripPrefix("/", fs))
	mux.Handle("GET /audio/", http.StripPrefix("/", fs))
	mux.Handle("GET /uploads/", http.StripPrefix("/", fs))
	mux.Handle("GET /manifest.json", http.StripPrefix("/", fs))

	return mux
}
