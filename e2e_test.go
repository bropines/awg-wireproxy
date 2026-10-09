package wireproxy

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amnezia-vpn/amneziawg-go/v3/device"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/net/proxy"
)

func genKeyPair(t *testing.T) (priv, pub string) {
	t.Helper()
	var k [32]byte
	if _, err := rand.Read(k[:]); err != nil {
		t.Fatal(err)
	}
	k[0] &= 248
	k[31] = (k[31] & 127) | 64
	p, err := curve25519.X25519(k[:], curve25519.Basepoint)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(k[:]), base64.StdEncoding.EncodeToString(p)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).Port
}

func startFromConfig(t *testing.T, text string) *VirtualTun {
	t.Helper()
	path := filepath.Join(t.TempDir(), "wireproxy.conf")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	conf, err := ParseConfig(path)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	vt, err := StartWireguard(conf, device.LogLevelSilent)
	if err != nil {
		t.Fatalf("StartWireguard: %v", err)
	}
	// The devices are intentionally left running until the test binary exits:
	// the tunnel routines treat a closed device as a fatal error.
	for _, r := range conf.Routines {
		go r.SpawnRoutine(vt)
	}
	return vt
}

// awgTunnelRoundTrip brings up an AmneziaWG server and client on loopback and
// fetches a page through client SOCKS5 -> tunnel -> server tunnel -> HTTP.
func awgTunnelRoundTrip(t *testing.T, awgParams string) {
	t.Helper()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "through-the-tunnel")
	}))
	t.Cleanup(backend.Close)

	sPriv, sPub := genKeyPair(t)
	cPriv, cPub := genKeyPair(t)
	wgPort := freeUDPPort(t)
	tunnelPort := 8000
	socksPort := freePort(t)

	startFromConfig(t, fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.77.0.1/32
ListenPort = %d
%s
[Peer]
PublicKey = %s
AllowedIPs = 10.77.0.2/32

[TCPServerTunnel]
ListenPort = %d
Target = %s
`, sPriv, wgPort, awgParams, cPub, tunnelPort, backend.Listener.Addr()))

	startFromConfig(t, fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.77.0.2/32
%s
[Peer]
PublicKey = %s
AllowedIPs = 10.77.0.1/32
Endpoint = 127.0.0.1:%d
PersistentKeepalive = 1-2

[Socks5]
BindAddress = 127.0.0.1:%d
`, cPriv, awgParams, sPub, wgPort, socksPort))

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), nil, proxy.Direct)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{Dial: dialer.Dial, DisableKeepAlives: true},
	}

	var lastErr error
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(fmt.Sprintf("http://10.77.0.1:%d/", tunnelPort))
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if string(body) == "through-the-tunnel" {
				return
			}
			lastErr = fmt.Errorf("unexpected body %q", body)
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("no traffic through the tunnel: %v", lastErr)
}

func TestTunnelPlainWireGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	awgTunnelRoundTrip(t, "")
}

func TestTunnelAmneziaWG1(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	awgTunnelRoundTrip(t, "Jc = 4\nJmin = 20\nJmax = 80\nS1 = 20\nS2 = 31\nH1 = 1000\nH2 = 2000\nH3 = 3000\nH4 = 4000\n")
}

func TestTunnelAmneziaWG3(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	awgTunnelRoundTrip(t, fmt.Sprintf(`Jc = 4
Jmin = 20
Jmax = 80
S1 = 20
S2 = 31
S3 = 42
S4 = 53
H1 = 1000-2000
H2 = 3000-4000
H3 = 5000-6000
H4 = 7000-8000
I1 = <b 0xc0ffee><r 16><t>
HeaderProtectionKey = %s
ContentPaddingAddition = 4-16
RekeyAfterTime = 100-120
RandomTrailers = true
`, base64.StdEncoding.EncodeToString(key)))
}
