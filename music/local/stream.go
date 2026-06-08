package local

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	localAudioRoutePrefix = "/local-audio/"
	localCoverRoutePrefix = "/local-cover/"
)

// AudioServer serves local library files over loopback HTTP with range request support.
// A dedicated server avoids Wails WebView host-check failures on <audio> subresource requests.
type AudioServer struct {
	store  *LocalLibraryStore
	secret []byte

	mu      sync.RWMutex
	baseURL string
	server  *http.Server
}

func NewAudioServer(store *LocalLibraryStore) *AudioServer {
	s := &AudioServer{
		store: store,
		secret: func() []byte {
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				panic(fmt.Sprintf("local audio server: generate secret: %v", err))
			}
			return secret
		}(),
	}
	if err := s.startLoopbackServer(); err != nil {
		panic(fmt.Sprintf("local audio server: start loopback server: %v", err))
	}
	return s
}

func (s *AudioServer) startLoopbackServer() error {
	mux := http.NewServeMux()
	mux.Handle(localAudioRoutePrefix, http.HandlerFunc(s.serveAudioHTTP))
	mux.Handle(localCoverRoutePrefix, http.HandlerFunc(s.serveCoverHTTP))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	baseURL := "http://" + ln.Addr().String()
	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(ln)
	}()

	s.mu.Lock()
	s.baseURL = baseURL
	s.server = server
	s.mu.Unlock()
	return nil
}

func (s *AudioServer) loopbackBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseURL
}

// StreamURL returns the loopback HTTP URL for local playback.
func (s *AudioServer) StreamURL(filePath string) (string, error) {
	if s.store == nil {
		return "", fmt.Errorf("local library is not initialized")
	}
	if err := s.store.ValidateLibraryPath(filePath); err != nil {
		return "", err
	}

	abs, err := filepath.Abs(strings.TrimSpace(filePath))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}

	base := s.loopbackBaseURL()
	if base == "" {
		return "", fmt.Errorf("local audio server is not ready")
	}
	return base + localAudioRoutePrefix + s.signToken(abs), nil
}

// CoverURL returns the loopback HTTP URL for a deduplicated cover key.
func (s *AudioServer) CoverURL(coverKey string) (string, error) {
	coverKey = strings.TrimSpace(coverKey)
	if coverKey == "" {
		return "", nil
	}
	if s.store == nil || s.store.coverFiles == nil {
		return "", fmt.Errorf("local cover store is not initialized")
	}
	if s.store.coverFiles.FilePath(coverKey) == "" {
		return "", nil
	}
	base := s.loopbackBaseURL()
	if base == "" {
		return "", fmt.Errorf("local audio server is not ready")
	}
	return base + localCoverRoutePrefix + s.signToken(coverKey), nil
}

func (s *AudioServer) signToken(value string) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(value))
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

func (s *AudioServer) verifyToken(token string) (string, bool) {
	dot := strings.LastIndex(token, ".")
	if dot <= 0 {
		return "", false
	}

	payload := token[:dot]
	sigPart := token[dot+1:]

	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	expected := mac.Sum(nil)

	got, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil || !hmac.Equal(got, expected) {
		return "", false
	}

	valueBytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", false
	}
	value := string(valueBytes)
	if value == "" {
		return "", false
	}
	return value, true
}

func (s *AudioServer) serveAudioHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, localAudioRoutePrefix)
	if token == "" || strings.Contains(token, "/") {
		http.NotFound(w, r)
		return
	}

	abs, ok := s.verifyToken(token)
	if !ok {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(abs)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "file stat failed", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	if ct := audioMIME(filepath.Ext(abs)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (s *AudioServer) serveCoverHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, localCoverRoutePrefix)
	if token == "" || strings.Contains(token, "/") {
		http.NotFound(w, r)
		return
	}

	coverKey, ok := s.verifyToken(token)
	if !ok || s.store == nil || s.store.coverFiles == nil {
		http.NotFound(w, r)
		return
	}

	path := s.store.coverFiles.FilePath(coverKey)
	if path == "" {
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "file stat failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mimeFromExt(filepath.Ext(path)))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}
