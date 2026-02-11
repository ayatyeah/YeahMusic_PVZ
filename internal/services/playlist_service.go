package services

import (
	"YeahMusic/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlaylistService struct {
	store *Store
}

func NewPlaylistService(store *Store) *PlaylistService {
	return &PlaylistService{store: store}
}

func (p *PlaylistService) Create(userID primitive.ObjectID, title, coverURL string) *models.Playlist {
	return p.store.CreatePlaylist(userID, title, coverURL)
}

func (p *PlaylistService) List(userID primitive.ObjectID) []*models.Playlist {
	return p.store.ListPlaylists(userID)
}

func (p *PlaylistService) AddTrack(userID, playlistID, trackID primitive.ObjectID) error {
	return p.store.AddTrackToPlaylist(userID, playlistID, trackID)
}

func (p *PlaylistService) RemoveTrack(userID, playlistID, trackID primitive.ObjectID) error {
	return p.store.RemoveTrackFromPlaylist(userID, playlistID, trackID)
}
