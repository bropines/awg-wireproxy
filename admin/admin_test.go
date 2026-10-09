package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const sampleConf = `[Interface]
Address = 10.5.0.2/32
PrivateKey = LAr1aNSNF9d0MjwUgAVC4020T0N/E5NUtqVv5EnsSz0=

[Peer]
PublicKey = e8LKAc+f9xEzq9Ar7+MfKRrs+gZ/4yzvpRJLRJ/VJ1w=
PresharedKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Endpoint = 94.140.11.15:51820

[Socks5]
BindAddress = 127.0.0.1:25344
Username = u
Password = hunter2
`

func TestMaskUnmaskRoundTrip(t *testing.T) {
	masked := MaskConfig(sampleConf)
	for _, secret := range []string{"LAr1aNSNF9d0", "AAAAAAAAAAAA", "hunter2"} {
		if strings.Contains(masked, secret) {
			t.Fatalf("secret %q leaked:\n%s", secret, masked)
		}
	}
	if !strings.Contains(masked, "PublicKey = e8LKAc") || !strings.Contains(masked, "Username = u") {
		t.Fatalf("non-secret values must stay visible:\n%s", masked)
	}
	back, err := UnmaskConfig(masked, sampleConf)
	if err != nil {
		t.Fatal(err)
	}
	if back != sampleConf {
		t.Fatalf("round trip changed the config:\n%s", back)
	}
}

func TestUnmaskKeepsEditsAndRejectsUnknown(t *testing.T) {
	edited := strings.Replace(MaskConfig(sampleConf), "127.0.0.1:25344", "127.0.0.1:1080", 1)
	back, err := UnmaskConfig(edited, sampleConf)
	if err != nil || !strings.Contains(back, "127.0.0.1:1080") || !strings.Contains(back, "hunter2") {
		t.Fatalf("edit lost or secret not restored: %v\n%s", err, back)
	}
	// a masked value in a section that did not exist before cannot be restored
	added := MaskConfig(sampleConf) + "\n[http]\nBindAddress = 127.0.0.1:8080\nPassword = ********\n"
	if _, err := UnmaskConfig(added, sampleConf); err == nil {
		t.Fatal("expected an error for a masked value without a previous one")
	}
	// a fresh real value is used as typed
	added = strings.Replace(added, "8080\nPassword = ********", "8080\nPassword = newpass", 1)
	if out, err := UnmaskConfig(added, sampleConf); err != nil || !strings.Contains(out, "newpass") {
		t.Fatalf("typed secret rejected: %v", err)
	}
}

func TestParsePeers(t *testing.T) {
	dump := "private_key=deadbeef\nlisten_port=0\npublic_key=" + strings.Repeat("ab", 32) +
		"\npreshared_key=" + strings.Repeat("00", 32) +
		"\nendpoint=1.2.3.4:51820\nlast_handshake_time_sec=1700000000\nrx_bytes=10\ntx_bytes=20\n" +
		"persistent_keepalive_interval=25\nallowed_ip=0.0.0.0/0\nallowed_ip=::/0\n"
	peers := ParsePeers(dump)
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	p := peers[0]
	if p.Endpoint != "1.2.3.4:51820" || p.RxBytes != 10 || p.TxBytes != 20 || p.LastHandshake != 1700000000 ||
		len(p.AllowedIPs) != 2 || p.Keepalive != "25" || len(p.PublicKey) != 44 {
		t.Fatalf("unexpected peer: %+v", p)
	}
}

func TestLogBuffer(t *testing.T) {
	b := NewLogBuffer(3)
	_, _ = b.Write([]byte("one\ntwo\nthr"))
	_, _ = b.Write([]byte("ee\nfour\n"))
	lines, next := b.Since(0)
	if len(lines) != 3 || lines[0].Text != "two" || lines[2].Text != "four" || next != 4 {
		t.Fatalf("got %+v next=%d", lines, next)
	}
	if more, _ := b.Since(next); len(more) != 0 {
		t.Fatal("expected no new lines")
	}
}

func newTestServer(t *testing.T) (*httptest.Server, string, *atomic.Int32) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "wireproxy.conf")
	if err := os.WriteFile(path, []byte(sampleConf), 0o600); err != nil {
		t.Fatal(err)
	}
	reloads := new(atomic.Int32)
	s := New(Options{ConfigPath: path, Token: "secret", Version: "test", Reload: func() { reloads.Add(1) }})
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts, path, reloads
}

func do(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, url, rd)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestAuthRequired(t *testing.T) {
	ts, _, _ := newTestServer(t)
	for _, tok := range []string{"", "wrong"} {
		if code, _ := do(t, "GET", ts.URL+"/api/config", tok, nil); code != http.StatusUnauthorized {
			t.Fatalf("token %q: got %d", tok, code)
		}
	}
	if code, _ := do(t, "GET", ts.URL+"/api/status", "secret", nil); code != http.StatusOK {
		t.Fatalf("valid token rejected: %d", code)
	}
	if code, _ := do(t, "GET", ts.URL+"/healthz", "", nil); code != http.StatusOK {
		t.Fatalf("healthz must be public: %d", code)
	}
}

func TestSessionLoginNeedsCSRFHeader(t *testing.T) {
	ts, _, _ := newTestServer(t)
	req, _ := http.NewRequest("POST", ts.URL+"/api/login", strings.NewReader(`{"token":"secret"}`))
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("login without the CSRF header must fail, got %d", resp.StatusCode)
	}
	req, _ = http.NewRequest("POST", ts.URL+"/api/login", strings.NewReader(`{"token":"secret"}`))
	req.Header.Set(csrfHeader, csrfValue)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK || len(resp.Cookies()) == 0 {
		t.Fatalf("login failed: %d", resp.StatusCode)
	}
	// cookie + state change without the header is refused
	req, _ = http.NewRequest("POST", ts.URL+"/api/reload", nil)
	req.AddCookie(resp.Cookies()[0])
	r2, _ := http.DefaultClient.Do(req)
	if r2.StatusCode != http.StatusForbidden {
		t.Fatalf("cookie POST without CSRF header got %d", r2.StatusCode)
	}
}

func TestConfigMaskedAndSaved(t *testing.T) {
	ts, path, reloads := newTestServer(t)

	_, got := do(t, "GET", ts.URL+"/api/config", "secret", nil)
	masked := got["config"].(string)
	if strings.Contains(masked, "LAr1aNSNF9d0") || !strings.Contains(masked, MaskValue) {
		t.Fatalf("config not masked: %s", masked)
	}

	// an invalid change is rejected and the file is untouched
	bad := strings.Replace(masked, "Address = 10.5.0.2/32", "Address = nonsense", 1)
	if code, _ := do(t, "PUT", ts.URL+"/api/config", "secret", map[string]any{"config": bad}); code != http.StatusBadRequest {
		t.Fatalf("invalid config accepted: %d", code)
	}
	if b, _ := os.ReadFile(path); string(b) != sampleConf {
		t.Fatal("file changed despite validation failure")
	}

	// a valid change keeps secrets, writes a backup and triggers the reload
	good := strings.Replace(masked, "127.0.0.1:25344", "127.0.0.1:1080", 1)
	code, resp := do(t, "PUT", ts.URL+"/api/config", "secret", map[string]any{"config": good, "apply": true})
	if code != http.StatusOK {
		t.Fatalf("save failed: %d %v", code, resp)
	}
	saved, _ := os.ReadFile(path)
	if !strings.Contains(string(saved), "127.0.0.1:1080") || !strings.Contains(string(saved), "hunter2") {
		t.Fatalf("saved config wrong:\n%s", saved)
	}
	if bak, _ := os.ReadFile(path + ".bak"); string(bak) != sampleConf {
		t.Fatal("backup missing or wrong")
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
		t.Fatalf("mode changed to %v", st.Mode().Perm())
	}
	waitFor(t, func() bool { return reloads.Load() == 1 })
}

func TestRestoreBackup(t *testing.T) {
	ts, path, _ := newTestServer(t)
	if code, _ := do(t, "POST", ts.URL+"/api/config/restore", "secret", nil); code != http.StatusNotFound {
		t.Fatalf("restore without backup got %d", code)
	}
	good := strings.Replace(sampleConf, "25344", "1080", 1)
	if code, _ := do(t, "PUT", ts.URL+"/api/config", "secret", map[string]any{"config": good}); code != http.StatusOK {
		t.Fatal("save failed")
	}
	if code, _ := do(t, "POST", ts.URL+"/api/config/restore", "secret", nil); code != http.StatusOK {
		t.Fatal("restore failed")
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), "25344") {
		t.Fatal("backup not restored")
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for i := 0; i < 50; i++ {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}
