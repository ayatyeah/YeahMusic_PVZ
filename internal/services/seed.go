package services

import (
	"YeahMusic/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func Seed(store *Store) {
	if len(store.ListTracks(primitive.NilObjectID)) > 0 {
		return
	}

	a1 := store.AddArtist(&models.Artist{Name: "OG Buda", Bio: "Demo artist"})
	a2 := store.AddArtist(&models.Artist{Name: "Ernar Beats", Bio: "Demo artist"})

	al1 := store.AddAlbum(&models.Album{ArtistID: a1.ID, Title: "Скучаю но Работаю", CoverURL: "https://via.placeholder.com/300", ReleaseYear: 2023})
	al2 := store.AddAlbum(&models.Album{ArtistID: a2.ID, Title: "Скучаю но Ещё Работаю", CoverURL: "https://via.placeholder.com/300", ReleaseYear: 2025})

	store.AddTrack(&models.Track{AlbumID: al1.ID, Title: "Ссора Я", ArtistName: "OG Buda", DurationSec: 180, AudioURL: "/audio/1.mp3", Lyrics: "Lyrics...", CoverURL: "https://via.placeholder.com/300"})
	store.AddTrack(&models.Track{AlbumID: al1.ID, Title: "Вода", ArtistName: "OG Buda", DurationSec: 210, AudioURL: "/audio/2.mp3", Lyrics: "Lyrics...", CoverURL: "https://via.placeholder.com/300"})
	store.AddTrack(&models.Track{AlbumID: al2.ID, Title: "Всё Норм", ArtistName: "Ernar Beats", DurationSec: 200, AudioURL: "/audio/3.mp3", Lyrics: "", CoverURL: "https://via.placeholder.com/300"})
}

func SessionCleanupWorker(store *Store, every time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case now := <-t.C:
			store.CleanupExpiredSessions(now)
		case <-stop:
			return
		}
	}
}
