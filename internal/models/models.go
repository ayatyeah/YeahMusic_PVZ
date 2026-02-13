package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email        string             `json:"email" bson:"email"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	Name         string             `json:"name" bson:"name"`
	Bio          string             `json:"bio" bson:"bio"`
	Role         string             `json:"role" bson:"role"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}

type Artist struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Bio       string             `json:"bio" bson:"bio"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type Album struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ArtistID    primitive.ObjectID `json:"artist_id" bson:"artist_id"`
	Title       string             `json:"title" bson:"title"`
	CoverURL    string             `json:"cover_url" bson:"cover_url"`
	ReleaseYear int                `json:"release_year" bson:"release_year"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type Track struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AlbumID     primitive.ObjectID `json:"album_id,omitempty" bson:"album_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	ArtistID    primitive.ObjectID `json:"artist_id" bson:"artist_id"`
	ArtistName  string             `json:"artist_name" bson:"artist_name"`
	DurationSec int                `json:"duration_sec" bson:"duration_sec"`
	AudioURL    string             `json:"audio_url" bson:"audio_url"`
	CoverURL    string             `json:"cover_url" bson:"cover_url"`
	Lyrics      string             `json:"lyrics" bson:"lyrics"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type Playlist struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Title     string             `json:"title" bson:"title"`
	CoverURL  string             `json:"cover_url" bson:"cover_url"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type PlaylistTrack struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PlaylistID primitive.ObjectID `json:"playlist_id" bson:"playlist_id"`
	TrackID    primitive.ObjectID `json:"track_id" bson:"track_id"`
	AddedAt    time.Time          `json:"added_at" bson:"added_at"`
}

type Session struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token     string             `json:"token" bson:"token"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
