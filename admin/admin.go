// Package admin implements the optional web panel and HTTP API that lets an
// operator inspect and reconfigure a running wireproxy.
package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	wireproxy "github.com/bropines/awg-wireproxy"
)

//go:embed web
var webFS embed.FS

const (
	sessionCookie  = "wireproxy_session"
	sessionTTL     = 24 * time.Hour
	csrfHeader     = "X-Requested-With"
	csrfValue      = "wireproxy-admin"
	maxBody        = 1 << 20
	loginWindow    = time.Minute
	loginMaxFailed = 10
)

// Options wires the panel to the running daemon.
type Options struct {
	ConfigPath string
	Token      string
	Version    string
	Logs       *LogBuffer
	// Tun returns the running tunnel, or nil when it is not running.
	Tun func() *wireproxy.VirtualTun
	// StartError returns why the tunnel failed to start, if it did.
	StartError func() string
	// Listeners describes the configured proxies and tunnels.
	Listeners func() []string
	// Reload restarts the daemon so a new config takes effect.
	Reload func()
}

// Server is the admin HTTP handler.
type Server struct {
	opts    Options
	started time.Time

	mu       sync.Mutex
	sessions map[string]time.Time
	failures []time.Time
}

// New creates the admin server.
func New(opts Options) *Server {
	if opts.Logs == nil {
		opts.Logs = NewLogBuffer(1)
	}
	return &Server{opts: opts, started: time.Now(), sessions: map[string]time.Time{}}
}

// GenerateToken returns a random access token.
func GenerateToken() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Handler returns the HTTP handler serving the panel and its API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	web, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServerFS(web))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})

	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.auth(s.logout))
	mux.HandleFunc("GET /api/status", s.auth(s.status))
	mux.HandleFunc("GET /api/config", s.auth(s.getConfig))
	mux.HandleFunc("PUT /api/config", s.auth(s.putConfig))
	mux.HandleFunc("POST /api/validate", s.auth(s.validate))
	mux.HandleFunc("POST /api/reload", s.auth(s.reload))
	mux.HandleFunc("POST /api/config/restore", s.auth(s.restore))
	mux.HandleFunc("GET /api/logs", s.auth(s.logs))

	return secureHeaders(mux)
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// ---- auth ----

func (s *Server) tokenOK(candidate string) bool {
	a := sha256.Sum256([]byte(candidate))
	b := sha256.Sum256([]byte(s.opts.Token))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") && s.tokenOK(strings.TrimPrefix(h, "Bearer ")) {
			next(w, r)
			return
		}
		c, err := r.Cookie(sessionCookie)
		if err != nil || !s.validSession(c.Value) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		// Cookie-authenticated requests that change state must come from
		// the panel itself (CSRF defence on top of SameSite=Strict).
		if r.Method != http.MethodGet {
			if r.Header.Get(csrfHeader) != csrfValue || !sameOrigin(r) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "bad origin"})
				return
			}
		}
		next(w, r)
	}
}

func sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		return true
	}
	return strings.HasSuffix(o, "://"+r.Host)
}

func (s *Server) validSession(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.sessions[id]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(s.sessions, id)
		return false
	}
	return true
}

func (s *Server) tooManyFailures() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-loginWindow)
	kept := s.failures[:0]
	for _, t := range s.failures {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	s.failures = kept
	return len(s.failures) >= loginMaxFailed
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(csrfHeader) != csrfValue || !sameOrigin(r) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "bad origin"})
		return
	}
	if s.tooManyFailures() {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts, try again in a minute"})
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := decodeBody(w, r, &req); err != nil || !s.tokenOK(req.Token) {
		s.mu.Lock()
		s.failures = append(s.failures, time.Now())
		s.mu.Unlock()
		time.Sleep(500 * time.Millisecond)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "wrong token"})
		return
	}

	id := GenerateToken()
	s.mu.Lock()
	s.sessions[id] = time.Now().Add(sessionTTL)
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: id, Path: "/", MaxAge: int(sessionTTL.Seconds()),
		HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		s.mu.Lock()
		delete(s.sessions, c.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---- API ----

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"version":     s.opts.Version,
		"uptime":      int64(time.Since(s.started).Seconds()),
		"config_path": s.opts.ConfigPath,
		"running":     false,
		"peers":       []PeerStatus{},
		"listeners":   []string{},
		"backup":      fileExists(s.opts.ConfigPath + ".bak"),
	}
	if s.opts.StartError != nil {
		resp["start_error"] = s.opts.StartError()
	}
	if s.opts.Listeners != nil {
		resp["listeners"] = s.opts.Listeners()
	}
	if s.opts.Tun != nil {
		if tun := s.opts.Tun(); tun != nil {
			resp["running"] = true
			if dump, err := tun.Dev.IpcGet(); err == nil {
				resp["peers"] = ParsePeers(dump)
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	text, exists, err := readConfig(s.opts.ConfigPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	reveal := r.URL.Query().Get("reveal") == "1"
	if !reveal {
		text = MaskConfig(text)
	}
	writeJSON(w, http.StatusOK, map[string]any{"config": text, "exists": exists, "masked": !reveal, "path": s.opts.ConfigPath})
}

type configRequest struct {
	Config string `json:"config"`
	Apply  bool   `json:"apply"`
}

// prepare restores masked secrets and validates the candidate config.
func (s *Server) prepare(text string) (string, error) {
	old, _, err := readConfig(s.opts.ConfigPath)
	if err != nil {
		return "", err
	}
	full, err := UnmaskConfig(text, old)
	if err != nil {
		return "", err
	}
	if err := validateConfig(s.opts.ConfigPath, full); err != nil {
		return "", err
	}
	return full, nil
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if _, err := s.prepare(req.Config); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) putConfig(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	full, err := s.prepare(req.Config)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := writeConfig(s.opts.ConfigPath, full); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.maybeReload(req.Apply)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "applying": req.Apply})
}

func (s *Server) restore(w http.ResponseWriter, r *http.Request) {
	bak, err := os.ReadFile(s.opts.ConfigPath + ".bak")
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no backup available"})
		return
	}
	if err := validateConfig(s.opts.ConfigPath, string(bak)); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "backup is not valid either: " + err.Error()})
		return
	}
	if err := writeConfig(s.opts.ConfigPath, string(bak)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.maybeReload(true)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "applying": true})
}

func (s *Server) reload(w http.ResponseWriter, r *http.Request) {
	if _, ok, _ := readConfig(s.opts.ConfigPath); !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "there is no config file to load yet"})
		return
	}
	s.maybeReload(true)
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "applying": true})
}

func (s *Server) maybeReload(apply bool) {
	if apply && s.opts.Reload != nil {
		// Give the HTTP response time to reach the browser first.
		time.AfterFunc(400*time.Millisecond, s.opts.Reload)
	}
}

func (s *Server) logs(w http.ResponseWriter, r *http.Request) {
	after, _ := strconv.ParseUint(r.URL.Query().Get("after"), 10, 64)
	lines, next := s.opts.Logs.Since(after)
	writeJSON(w, http.StatusOK, map[string]any{"lines": lines, "next": next})
}

// ---- helpers ----

func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return errors.New("invalid request body")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
