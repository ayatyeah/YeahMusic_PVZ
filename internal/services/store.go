package services

import (
	"context"
	"errors"
	"log"
	"time"

	"YeahMusic/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrAlreadyExists = errors.New("already exists")
)

type Store struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewStore(uri string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Store{client: client, db: client.Database("YeahMusic")}, nil
}

func (s *Store) Now() time.Time { return time.Now() }

func (s *Store) CreateUser(u *models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, _ := s.db.Collection("users").CountDocuments(ctx, bson.M{"email": u.Email})
	if count > 0 {
		return nil, ErrAlreadyExists
	}

	u.ID = primitive.NewObjectID()
	u.CreatedAt = s.Now()

	_, err := s.db.Collection("users").InsertOne(ctx, u)
	if err != nil {
		log.Println("CreateUser error:", err)
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}

func (s *Store) GetUserByID(id primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u models.User
	err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}

func (s *Store) CreateSession(userID primitive.ObjectID, token string, ttl time.Duration) *models.Session {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sec := &models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Token:     token,
		CreatedAt: s.Now(),
		ExpiresAt: s.Now().Add(ttl),
	}
	s.db.Collection("sessions").InsertOne(ctx, sec)
	return sec
}

func (s *Store) GetSession(token string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sec models.Session
	err := s.db.Collection("sessions").FindOne(ctx, bson.M{"token": token}).Decode(&sec)
	if err != nil {
		return nil, ErrNotFound
	}
	return &sec, nil
}

func (s *Store) CleanupExpiredSessions(now time.Time) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, _ := s.db.Collection("sessions").DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lte": now}})
	return int(res.DeletedCount)
}

func (s *Store) AddArtist(a *models.Artist) *models.Artist {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing models.Artist
	err := s.db.Collection("artists").FindOne(ctx, bson.M{"name": a.Name}).Decode(&existing)
	if err == nil {
		return &existing
	}

	a.ID = primitive.NewObjectID()
	s.db.Collection("artists").InsertOne(ctx, a)
	return a
}

func (s *Store) AddAlbum(a *models.Album) *models.Album {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.ID = primitive.NewObjectID()
	s.db.Collection("albums").InsertOne(ctx, a)
	return a
}

func (s *Store) AddTrack(t *models.Track) *models.Track {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.ID = primitive.NewObjectID()
	s.db.Collection("tracks").InsertOne(ctx, t)
	return t
}

func (s *Store) UpdateTrack(id primitive.ObjectID, coverURL, lyrics string) (*models.Track, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if coverURL != "" {
		update["cover_url"] = coverURL
	}
	if lyrics != "" {
		update["lyrics"] = lyrics
	}

	if len(update) == 0 {
		var t models.Track
		s.db.Collection("tracks").FindOne(ctx, bson.M{"_id": id}).Decode(&t)
		return &t, nil
	}

	res := s.db.Collection("tracks").FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var t models.Track
	if err := res.Decode(&t); err != nil {
		return nil, ErrNotFound
	}
	return &t, nil
}

func (s *Store) DeleteTrack(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := s.db.Collection("tracks").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}

	s.db.Collection("playlist_tracks").DeleteMany(ctx, bson.M{"track_id": id})
	return nil
}

func (s *Store) ListArtists() []*models.Artist {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res []*models.Artist
	cur, _ := s.db.Collection("artists").Find(ctx, bson.M{})
	defer cur.Close(ctx)
	cur.All(ctx, &res)
	if res == nil {
		return []*models.Artist{}
	}
	return res
}

func (s *Store) ListAlbums(artistID primitive.ObjectID) []*models.Album {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res []*models.Album
	filter := bson.M{}
	if !artistID.IsZero() {
		filter["artist_id"] = artistID
	}

	cur, _ := s.db.Collection("albums").Find(ctx, filter)
	defer cur.Close(ctx)
	cur.All(ctx, &res)
	if res == nil {
		return []*models.Album{}
	}
	return res
}

func (s *Store) ListTracks(albumID primitive.ObjectID) []*models.Track {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res []*models.Track
	filter := bson.M{}
	if !albumID.IsZero() {
		filter["album_id"] = albumID
	}

	opts := options.Find().SetSort(bson.M{"_id": -1})
	cur, _ := s.db.Collection("tracks").Find(ctx, filter, opts)
	defer cur.Close(ctx)
	cur.All(ctx, &res)
	if res == nil {
		return []*models.Track{}
	}
	return res
}

func (s *Store) CreatePlaylist(userID primitive.ObjectID, title, coverURL string) *models.Playlist {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p := &models.Playlist{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Title:     title,
		CoverURL:  coverURL,
		CreatedAt: s.Now(),
	}
	s.db.Collection("playlists").InsertOne(ctx, p)
	return p
}

func (s *Store) ListPlaylists(userID primitive.ObjectID) []*models.Playlist {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res []*models.Playlist
	cur, _ := s.db.Collection("playlists").Find(ctx, bson.M{"user_id": userID})
	defer cur.Close(ctx)
	cur.All(ctx, &res)
	if res == nil {
		return []*models.Playlist{}
	}
	return res
}

func (s *Store) AddTrackToPlaylist(userID, playlistID, trackID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p models.Playlist
	err := s.db.Collection("playlists").FindOne(ctx, bson.M{"_id": playlistID}).Decode(&p)
	if err != nil {
		return ErrNotFound
	}
	if p.UserID != userID {
		return ErrForbidden
	}

	pt := &models.PlaylistTrack{
		ID:         primitive.NewObjectID(),
		PlaylistID: playlistID,
		TrackID:    trackID,
		AddedAt:    s.Now(),
	}
	_, err = s.db.Collection("playlist_tracks").InsertOne(ctx, pt)
	return err
}

func (s *Store) RemoveTrackFromPlaylist(userID, playlistID, trackID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p models.Playlist
	err := s.db.Collection("playlists").FindOne(ctx, bson.M{"_id": playlistID}).Decode(&p)
	if err != nil {
		return ErrNotFound
	}
	if p.UserID != userID {
		return ErrForbidden
	}

	res, _ := s.db.Collection("playlist_tracks").DeleteOne(ctx, bson.M{"playlist_id": playlistID, "track_id": trackID})
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
