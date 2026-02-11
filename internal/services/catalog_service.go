package services

import (
	"YeahMusic/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatalogService struct {
	store *Store
}

func NewCatalogService(store *Store) *CatalogService {
	return &CatalogService{store: store}
}

func (c *CatalogService) ListArtists() []*models.Artist {
	return c.store.ListArtists()
}

func (c *CatalogService) ListAlbums(artistID primitive.ObjectID) []*models.Album {
	return c.store.ListAlbums(artistID)
}

func (c *CatalogService) ListTracks(albumID primitive.ObjectID) []*models.Track {
	return c.store.ListTracks(albumID)
}

func (c *CatalogService) AddTrack(t *models.Track) *models.Track {
	return c.store.AddTrack(t)
}

func (c *CatalogService) UpdateTrack(id primitive.ObjectID, coverURL, lyrics string) (*models.Track, error) {
	return c.store.UpdateTrack(id, coverURL, lyrics)
}

func (c *CatalogService) DeleteTrack(id primitive.ObjectID) error {
	return c.store.DeleteTrack(id)
}

func (c *CatalogService) CreateAlbum(a *models.Album) *models.Album {
	return c.store.AddAlbum(a)
}
