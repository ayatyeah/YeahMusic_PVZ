package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client
var db *mongo.Database

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"-"`
	Name     string             `bson:"name" json:"name"`
	Bio      string             `bson:"bio" json:"bio"`
	Role     string             `bson:"role" json:"role"`
}

type Track struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Artist    string             `bson:"artist" json:"artist"`
	ArtistID  primitive.ObjectID `bson:"artist_id" json:"artist_id"`
	AlbumID   primitive.ObjectID `bson:"album_id,omitempty" json:"album_id"`
	CoverURL  string             `bson:"cover_url" json:"cover_url"`
	AudioURL  string             `bson:"audio_url" json:"audio_url"`
	Lyrics    string             `bson:"lyrics" json:"lyrics"`
	Duration  int                `bson:"duration" json:"duration"`
	IsSingle  bool               `bson:"is_single" json:"is_single"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Album struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Artist      string             `bson:"artist" json:"artist"`
	ArtistID    primitive.ObjectID `bson:"artist_id" json:"artist_id"`
	CoverURL    string             `bson:"cover_url" json:"cover_url"`
	IsSingle    bool               `bson:"is_single" json:"is_single"`
	ReleaseDate time.Time          `bson:"release_date" json:"release_date"`
}

type Playlist struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title     string               `bson:"title" json:"title"`
	Creator   string               `bson:"creator" json:"creator"`
	CreatorID primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	CoverURL  string               `bson:"cover_url" json:"cover_url"`
	Tracks    []primitive.ObjectID `bson:"tracks" json:"tracks"`
}

type generateLyricsReq struct {
	Genre       string `json:"genre"`
	About       string `json:"about"`
	Language    string `json:"language"`
	Extra       string `json:"extra"`
	MinLines    int    `json:"min_lines"`
	Prompt      string `json:"prompt"`
	Description string `json:"description"`
}

type geminiResp struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func main() {
	_ = os.MkdirAll("public/uploads", 0755)
	loadDotEnv(".env")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(getMongoURI()))
	if err != nil {
		log.Fatal(err)
	}
	db = client.Database(getMongoDBName())

	initIndexes()

	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/update-profile", updateProfileHandler)

	http.HandleFunc("/api/upload-track", uploadTrackHandler)
	http.HandleFunc("/api/delete-track", deleteTrackHandler)
	http.HandleFunc("/api/create-album", createAlbumHandler)
	http.HandleFunc("/api/create-playlist", createPlaylistHandler)
	http.HandleFunc("/api/add-to-playlist", addToPlaylistHandler)

	http.HandleFunc("/api/update-lyrics", updateLyricsHandler)

	http.HandleFunc("/api/content", contentHandler)
	http.HandleFunc("/api/search", searchHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/album", getAlbumHandler)
	http.HandleFunc("/api/playlist", getPlaylistHandler)
	http.HandleFunc("/api/user-playlists", getUserPlaylistsHandler)
	http.HandleFunc("/api/artist-albums", getArtistAlbumsHandler)

	http.HandleFunc("/api/generate-lyrics", generateLyricsHandler)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getMongoURI() string {
	if v := strings.TrimSpace(os.Getenv("MONGODB_URI")); v != "" {
		return v
	}
	return "mongodb://localhost:27017"
}

func getMongoDBName() string {
	if v := strings.TrimSpace(os.Getenv("DB_NAME")); v != "" {
		return v
	}
	return "YeahMusicDiamond"
}

func loadDotEnv(path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		v = strings.Trim(v, `"'`)
		if k == "" {
			continue
		}
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func initIndexes() {
	ctx := context.Background()

	_, _ = db.Collection("tracks").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: "text"}, {Key: "artist", Value: "text"}},
	})

	_, _ = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	_, _ = db.Collection("albums").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "artist_id", Value: 1}, {Key: "is_single", Value: 1}},
	})
}

func jsonOut(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func allow(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func getAnyString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			}
		}
	}
	return ""
}

func updateLyricsHandler(w http.ResponseWriter, r *http.Request) {
	allow(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)

	trackID := strings.TrimSpace(r.URL.Query().Get("track_id"))
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	lyrics := ""

	if trackID == "" {
		trackID = getAnyString(body, "track_id", "trackId", "id", "track")
	}
	if userID == "" {
		userID = getAnyString(body, "user_id", "userId", "artist_id", "artistId", "user")
	}

	if v, ok := body["lyrics"]; ok {
		if s, ok2 := v.(string); ok2 {
			lyrics = s
		}
	}
	if lyrics == "" {
		if v, ok := body["text"]; ok {
			if s, ok2 := v.(string); ok2 {
				lyrics = s
			}
		}
	}

	if trackID == "" {
		http.Error(w, "track_id required", 400)
		return
	}

	tid, err := primitive.ObjectIDFromHex(trackID)
	if err != nil {
		http.Error(w, "bad track_id", 400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var t Track
	if err := db.Collection("tracks").FindOne(ctx, bson.M{"_id": tid}).Decode(&t); err != nil {
		http.Error(w, "track not found", 404)
		return
	}

	if userID != "" {
		uid, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			if t.ArtistID != uid {
				http.Error(w, "forbidden", 403)
				return
			}
		}
	}

	_, err = db.Collection("tracks").UpdateOne(
		ctx,
		bson.M{"_id": tid},
		bson.M{"$set": bson.M{"lyrics": lyrics}},
	)
	if err != nil {
		http.Error(w, "update failed", 500)
		return
	}

	jsonOut(w, 200, map[string]any{"ok": true})
}

func generateLyricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req generateLyricsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	genre := strings.TrimSpace(req.Genre)
	about := strings.TrimSpace(req.About)
	lang := strings.TrimSpace(req.Language)
	extra := strings.TrimSpace(req.Extra)

	if lang == "" {
		lang = "ru"
	}

	userPrompt := strings.TrimSpace(req.Prompt)
	desc := strings.TrimSpace(req.Description)

	var prompt string
	if userPrompt != "" {
		prompt = userPrompt
	} else {
		base := about
		if base == "" {
			base = desc
		}

		prompt = fmt.Sprintf(
			"You are a professional songwriter. Output ONLY lyrics. No explanations. No notes.\n\n"+
				"Language: %s.\n"+
				"Genre: %s.\n"+
				"Theme/Story: %s.\n"+
				"Extra context: %s.\n\n"+
				"ABSOLUTE RULES (must follow or you failed):\n"+
				"1) Use ONLY these section tags, exactly and in this order:\n"+
				"[Intro]\n"+
				"[Chorus]\n"+
				"[Verse]\n"+
				"[Bridge]\n"+
				"[Chorus]\n"+
				"[Outro]\n"+
				"2) Each lyric line MUST end with an adlib in parentheses, exactly like: (u) (yeah) (tss) (fa) (e) (e-e-e) (baby)\n"+
				"3) RHYME RULE: every block of 4 lines must rhyme with each other (AAAA). So lines 1-4 rhyme, 5-8 rhyme, etc.\n"+
				"4) LENGTH RULES:\n"+
				"- [Intro] exactly 2 lines. One of these lines MUST contain the exact phrase: \"special for yeah music buddy\".\n"+
				"- [Chorus] exactly 4 lines.\n"+
				"- [Verse] exactly 8 lines and each line must be long (at least 8 words).\n"+
				"- [Bridge] exactly 4 lines.\n"+
				"- Second [Chorus] exactly 8 lines total: the first 4 lines must be EXACTLY the same as the first [Chorus], and then add 4 new lines that are a logical and rhyming continuation.\n"+
				"- [Outro] exactly 2 lines that release the song.\n"+
				"5) No emojis. No profanity. No numbering like Verse 1. No extra sections.\n"+
				"6) Do not leave any section incomplete. Do not cut off. Make sure the output contains ALL sections.\n\n"+
				"Now generate the lyrics with strong, obvious rhymes and clean structure.",
			lang, genre, base, extra,
		)
	}

	text, err := callGeminiGenerateWithRetry(prompt, 3)
	if err != nil {
		http.Error(w, "AI error: "+err.Error(), 500)
		return
	}

	text = sanitizeGeminiText(text)

	ok := validateLyricsStructure(text)
	if !ok {
		text2, err2 := callGeminiGenerateWithRetry(prompt+"\n\nIMPORTANT: Your previous output violated structure. Rewrite and strictly follow ALL rules.", 2)
		if err2 == nil {
			text2 = sanitizeGeminiText(text2)
			if validateLyricsStructure(text2) {
				text = text2
			} else {
				http.Error(w, "AI error: invalid structure output", 500)
				return
			}
		} else {
			http.Error(w, "AI error: invalid structure output", 500)
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]string{"lyrics": text})
}

func callGeminiGenerateWithRetry(prompt string, attempts int) (string, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		txt, err := callGeminiGenerate(prompt)
		if err == nil {
			return txt, nil
		}
		lastErr = err
		time.Sleep(time.Duration(600+300*i) * time.Millisecond)
	}
	return "", lastErr
}

func callGeminiGenerate(prompt string) (string, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)

	body := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.9,
			"topP":            0.9,
			"maxOutputTokens": 4096,
		},
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)

	c := &http.Client{Timeout: 45 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("gemini status %d: %s", res.StatusCode, string(raw))
	}

	var out geminiResp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response")
	}

	txt := out.Candidates[0].Content.Parts[0].Text
	return strings.TrimSpace(txt), nil
}

func sanitizeGeminiText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimRight(ln, " \t")
		out = append(out, ln)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func validateLyricsStructure(text string) bool {
	if !strings.Contains(text, "[Intro]") ||
		!strings.Contains(text, "[Chorus]") ||
		!strings.Contains(text, "[Verse]") ||
		!strings.Contains(text, "[Bridge]") ||
		!strings.Contains(text, "[Outro]") {
		return false
	}

	order := []string{"[Intro]", "[Chorus]", "[Verse]", "[Bridge]", "[Chorus]", "[Outro]"}
	pos := 0
	for _, tag := range order {
		i := strings.Index(text[pos:], tag)
		if i < 0 {
			return false
		}
		pos += i + len(tag)
	}

	if !strings.Contains(text, "special for yeah music buddy") {
		return false
	}

	sections := splitSections(text)
	if len(sections) != 6 {
		return false
	}

	intro := sections[0].lines
	chorus1 := sections[1].lines
	verse := sections[2].lines
	bridge := sections[3].lines
	chorus2 := sections[4].lines
	outro := sections[5].lines

	if len(intro) != 2 || len(chorus1) != 4 || len(verse) != 8 || len(bridge) != 4 || len(chorus2) != 8 || len(outro) != 2 {
		return false
	}

	for _, ln := range append(append(append(append(append(intro, chorus1...), verse...), bridge...), chorus2...), outro...) {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			return false
		}
		if !strings.HasSuffix(ln, ")") {
			return false
		}
		open := strings.LastIndex(ln, "(")
		close := strings.LastIndex(ln, ")")
		if open < 0 || close < 0 || close <= open {
			return false
		}
	}

	for i := 0; i < 4; i++ {
		if strings.TrimSpace(chorus2[i]) != strings.TrimSpace(chorus1[i]) {
			return false
		}
	}

	return true
}

type section struct {
	tag   string
	lines []string
}

func splitSections(text string) []section {
	lines := strings.Split(text, "\n")
	var sections []section
	var cur *section

	isTag := func(s string) bool {
		s = strings.TrimSpace(s)
		return s == "[Intro]" || s == "[Chorus]" || s == "[Verse]" || s == "[Bridge]" || s == "[Outro]"
	}

	for _, raw := range lines {
		ln := strings.TrimSpace(raw)
		if ln == "" {
			continue
		}
		if isTag(ln) {
			sections = append(sections, section{tag: ln, lines: []string{}})
			cur = &sections[len(sections)-1]
			continue
		}
		if cur == nil {
			continue
		}
		cur.lines = append(cur.lines, raw)
	}

	if len(sections) != 6 {
		return nil
	}
	return sections
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	u.Email = strings.TrimSpace(u.Email)
	u.Name = strings.TrimSpace(u.Name)
	u.Role = strings.TrimSpace(u.Role)

	if u.Email == "" || u.Password == "" {
		http.Error(w, "Email and password required", 400)
		return
	}
	if u.Name == "" {
		u.Name = "User"
	}
	if u.Role == "" {
		u.Role = "user"
	}

	count, _ := db.Collection("users").CountDocuments(context.Background(), bson.M{"email": u.Email})
	if count > 0 {
		http.Error(w, "Email already taken", 409)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	u.Password = string(hash)
	u.ID = primitive.NewObjectID()

	_, err := db.Collection("users").InsertOne(context.Background(), u)
	if err != nil {
		http.Error(w, "Server error", 500)
		return
	}
	jsonOut(w, 200, u)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var u User
	err := db.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&u)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid password", 401)
		return
	}
	jsonOut(w, 200, u)
}

func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Bio  string `json:"bio"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	oid, _ := primitive.ObjectIDFromHex(req.ID)

	update := bson.M{
		"name": strings.TrimSpace(req.Name),
		"bio":  strings.TrimSpace(req.Bio),
	}
	_, _ = db.Collection("users").UpdateOne(context.Background(), bson.M{"_id": oid}, bson.M{"$set": update})

	var u User
	_ = db.Collection("users").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&u)
	jsonOut(w, 200, u)
}

func createAlbumHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(10 << 20)
	title := strings.TrimSpace(r.FormValue("title"))
	artistName := strings.TrimSpace(r.FormValue("artist_name"))
	artistID, _ := primitive.ObjectIDFromHex(r.FormValue("artist_id"))
	coverURL := saveFile(r, "cover")

	if title == "" {
		http.Error(w, "title required", 400)
		return
	}

	album := Album{
		ID: primitive.NewObjectID(), Title: title, Artist: artistName, ArtistID: artistID,
		CoverURL: coverURL, IsSingle: false, ReleaseDate: time.Now(),
	}
	_, _ = db.Collection("albums").InsertOne(context.Background(), album)
	jsonOut(w, 200, album)
}

func uploadTrackHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(50 << 20)
	title := strings.TrimSpace(r.FormValue("title"))
	artistName := strings.TrimSpace(r.FormValue("artist_name"))
	artistID, _ := primitive.ObjectIDFromHex(r.FormValue("artist_id"))
	albumIDStr := strings.TrimSpace(r.FormValue("album_id"))
	lyrics := r.FormValue("lyrics")

	if title == "" {
		http.Error(w, "title required", 400)
		return
	}

	var albumID primitive.ObjectID
	var coverURL string
	var isSingle bool

	if albumIDStr != "" && albumIDStr != "single" {
		albumID, _ = primitive.ObjectIDFromHex(albumIDStr)
		var alb Album
		_ = db.Collection("albums").FindOne(context.Background(), bson.M{"_id": albumID}).Decode(&alb)
		coverURL = alb.CoverURL
		isSingle = false
	} else {
		albumID = primitive.NewObjectID()
		coverURL = saveFile(r, "cover")
		isSingle = true

		album := Album{
			ID: albumID, Title: title, Artist: artistName, ArtistID: artistID,
			CoverURL: coverURL, IsSingle: true, ReleaseDate: time.Now(),
		}
		_, _ = db.Collection("albums").InsertOne(context.Background(), album)
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio required", 400)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, _ := os.Create("public/uploads/" + name)
	_, _ = io.Copy(dst, file)
	_ = dst.Close()

	track := Track{
		ID: primitive.NewObjectID(), Title: title, Artist: artistName, ArtistID: artistID,
		AlbumID: albumID, CoverURL: coverURL, AudioURL: "/uploads/" + name,
		Lyrics: lyrics, IsSingle: isSingle, CreatedAt: time.Now(),
	}
	_, _ = db.Collection("tracks").InsertOne(context.Background(), track)

	jsonOut(w, 200, map[string]string{"status": "ok"})
}

func deleteTrackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	oid, _ := primitive.ObjectIDFromHex(req.ID)

	_, _ = db.Collection("tracks").DeleteOne(context.Background(), bson.M{"_id": oid})
	_, _ = db.Collection("playlists").UpdateMany(context.Background(), bson.M{}, bson.M{"$pull": bson.M{"tracks": oid}})

	jsonOut(w, 200, map[string]string{"status": "deleted"})
}

func createPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(10 << 20)
	title := strings.TrimSpace(r.FormValue("title"))
	creator := strings.TrimSpace(r.FormValue("creator"))
	creatorID, _ := primitive.ObjectIDFromHex(r.FormValue("creator_id"))
	coverURL := saveFile(r, "cover")

	if title == "" {
		http.Error(w, "title required", 400)
		return
	}

	p := Playlist{
		ID: primitive.NewObjectID(), Title: title, Creator: creator, CreatorID: creatorID,
		CoverURL: coverURL, Tracks: []primitive.ObjectID{},
	}
	_, _ = db.Collection("playlists").InsertOne(context.Background(), p)
	jsonOut(w, 200, p)
}

func addToPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlaylistID string `json:"playlist_id"`
		TrackID    string `json:"track_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	pID, _ := primitive.ObjectIDFromHex(req.PlaylistID)
	tID, _ := primitive.ObjectIDFromHex(req.TrackID)

	_, _ = db.Collection("playlists").UpdateOne(context.Background(),
		bson.M{"_id": pID},
		bson.M{"$addToSet": bson.M{"tracks": tID}},
	)
	jsonOut(w, 200, map[string]string{"status": "added"})
}

func contentHandler(w http.ResponseWriter, r *http.Request) {
	var albums []Album
	var singles []Album
	var playlists []Playlist

	cur, _ := db.Collection("albums").Find(context.Background(), bson.M{"is_single": false})
	_ = cur.All(context.Background(), &albums)

	cur2, _ := db.Collection("albums").Find(context.Background(), bson.M{"is_single": true})
	_ = cur2.All(context.Background(), &singles)

	cur3, _ := db.Collection("playlists").Find(context.Background(), bson.M{})
	_ = cur3.All(context.Background(), &playlists)

	resp := map[string]any{
		"albums": func() []Album {
			if albums == nil {
				return []Album{}
			}
			return albums
		}(),
		"singles": func() []Album {
			if singles == nil {
				return []Album{}
			}
			return singles
		}(),
		"playlists": func() []Playlist {
			if playlists == nil {
				return []Playlist{}
			}
			return playlists
		}(),
	}
	jsonOut(w, 200, resp)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		jsonOut(w, 200, []Track{})
		return
	}

	filter := bson.M{"$text": bson.M{"$search": q}}

	var tracks []Track
	cur, err := db.Collection("tracks").Find(context.Background(), filter)
	if err != nil {
		filter = bson.M{"title": bson.M{"$regex": q, "$options": "i"}}
		cur, _ = db.Collection("tracks").Find(context.Background(), filter)
	}
	_ = cur.All(context.Background(), &tracks)
	if tracks == nil {
		tracks = []Track{}
	}
	jsonOut(w, 200, tracks)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$artist"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 5}},
	}

	cursor, _ := db.Collection("tracks").Aggregate(context.Background(), pipeline)
	var topArtists []bson.M
	_ = cursor.All(context.Background(), &topArtists)

	usersCount, _ := db.Collection("users").CountDocuments(context.Background(), bson.M{})
	tracksCount, _ := db.Collection("tracks").CountDocuments(context.Background(), bson.M{})

	jsonOut(w, 200, map[string]any{
		"users":       usersCount,
		"tracks":      tracksCount,
		"top_artists": topArtists,
	})
}

func getAlbumHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("id"))
	var album Album
	_ = db.Collection("albums").FindOne(context.Background(), bson.M{"_id": id}).Decode(&album)

	var tracks []Track
	cur, _ := db.Collection("tracks").Find(context.Background(), bson.M{"album_id": id})
	_ = cur.All(context.Background(), &tracks)

	if tracks == nil {
		tracks = []Track{}
	}
	jsonOut(w, 200, map[string]any{"info": album, "tracks": tracks})
}

func getPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("id"))
	var pl Playlist
	_ = db.Collection("playlists").FindOne(context.Background(), bson.M{"_id": id}).Decode(&pl)

	var tracks []Track
	if len(pl.Tracks) > 0 {
		cur, _ := db.Collection("tracks").Find(context.Background(), bson.M{"_id": bson.M{"$in": pl.Tracks}})
		_ = cur.All(context.Background(), &tracks)
	}
	if tracks == nil {
		tracks = []Track{}
	}
	jsonOut(w, 200, map[string]any{"info": pl, "tracks": tracks})
}

func getUserPlaylistsHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("user_id"))
	var playlists []Playlist
	cur, _ := db.Collection("playlists").Find(context.Background(), bson.M{"creator_id": id})
	_ = cur.All(context.Background(), &playlists)
	if playlists == nil {
		playlists = []Playlist{}
	}
	jsonOut(w, 200, playlists)
}

func getArtistAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := primitive.ObjectIDFromHex(r.URL.Query().Get("artist_id"))
	var albums []Album
	cur, _ := db.Collection("albums").Find(context.Background(), bson.M{"artist_id": id, "is_single": false})
	_ = cur.All(context.Background(), &albums)
	if albums == nil {
		albums = []Album{}
	}
	jsonOut(w, 200, albums)
}

func saveFile(r *http.Request, key string) string {
	file, header, err := r.FormFile(key)
	if err != nil {
		return ""
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	name := fmt.Sprintf("%d_img%s", time.Now().UnixNano(), ext)

	dst, err := os.Create("public/uploads/" + name)
	if err != nil {
		return ""
	}
	_, _ = io.Copy(dst, file)
	_ = dst.Close()

	return "/uploads/" + name
}
