package models

type Song struct {
	ID          int64  `json:"id"`
	AlbumID     int64  `json:"album_id"`
	Title       string `json:"title"`
	DurationSec int    `json:"duration_sec"`
	AudioURL    string `json:"audio_url"`
}
