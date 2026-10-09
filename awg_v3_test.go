package wireproxy

import (
	"strings"
	"testing"
)

const awgV3Conf = `
[Interface]
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Address = 10.0.0.2/32
Jc = 5
Jmin = 10
Jmax = 50
S1 = 20
S2 = 30
S3 = 40
S4 = 50
H1 = 100-200
H2 = 300-400
H3 = 500-600
H4 = 700-800
HeaderProtectionKey = %s
ContentPaddingAddition = 4-16
RekeyAfterTime = 100-120
MaxHandshakeAttempts = 5
RandomTrailers = true
DisableCookies = false

[Peer]
PublicKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Endpoint = 1.2.3.4:51820
PersistentKeepalive = 20-30
`

func loadAWGV3(t *testing.T, key string) (*DeviceSetting, error) {
	t.Helper()
	cfg, err := loadIniConfig(strings.Replace(awgV3Conf, "%s", key, 1))
	if err != nil {
		t.Fatal(err)
	}
	var dev DeviceConfig
	if err := ParseInterface(cfg, &dev); err != nil {
		return nil, err
	}
	if err := ParsePeers(cfg, &dev.Peers); err != nil {
		return nil, err
	}
	return CreateIPCRequest(&dev)
}

func TestAWGV3Params(t *testing.T) {
	key := strings.Repeat("ab", 32)
	s, err := loadAWGV3(t, key)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"header_protection_key=" + key,
		"content_padding_addition=4-16",
		"rekey_after_time=100-120",
		"max_handshake_attempts=5",
		"random_trailers=true",
		"disable_cookies=false",
		"persistent_keepalive_interval=20-30",
	} {
		if !strings.Contains(s.IpcRequest, want) {
			t.Errorf("IPC request missing %q:\n%s", want, s.IpcRequest)
		}
	}
}

func TestAWGV3Invalid(t *testing.T) {
	if _, err := loadAWGV3(t, "nothex"); err == nil {
		t.Error("expected error for bad HeaderProtectionKey")
	}
	cfg, _ := loadIniConfig("[Interface]\nRekeyTimeout = 9-3\n")
	if _, err := parseAWGV3Params(cfg.Section("Interface")); err == nil {
		t.Error("expected error for inverted range")
	}
}
