package wireproxy

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
)

// AmneziaWG 3.x extensions of the [Interface] section. Each entry maps the
// config key to the UAPI key understood by amneziawg-go/v3.
type awgV3Kind int

const (
	awgV3Range awgV3Kind = iota // "a" or "a-b" (uint32)
	awgV3Key                    // 32 bytes, hex or base64 encoded
	awgV3Bool
)

type awgV3Param struct {
	confKey string
	uapiKey string
	kind    awgV3Kind
}

var awgV3Params = []awgV3Param{
	{"HeaderProtectionKey", "header_protection_key", awgV3Key},
	{"ContentPaddingAddition", "content_padding_addition", awgV3Range},
	{"RekeyAfterTime", "rekey_after_time", awgV3Range},
	{"RekeyTimeout", "rekey_timeout", awgV3Range},
	{"RejectAfterTime", "reject_after_time", awgV3Range},
	{"KeepaliveTimeout", "keepalive_timeout", awgV3Range},
	{"MaxHandshakeAttempts", "max_handshake_attempts", awgV3Range},
	{"RandomTrailers", "random_trailers", awgV3Bool},
	{"DisableCookies", "disable_cookies", awgV3Bool},
}

// parseAWGV3Params reads the AWG 3.x keys of a section and returns them as
// normalized UAPI values keyed by UAPI key.
func parseAWGV3Params(section *ini.Section) (map[string]string, error) {
	var out map[string]string
	for _, p := range awgV3Params {
		key, err := section.GetKey(p.confKey)
		if err != nil {
			continue
		}
		raw := strings.TrimSpace(key.String())
		var value string
		switch p.kind {
		case awgV3Range:
			value, err = normalizeUintRange(raw)
		case awgV3Key:
			value, err = normalizeHeaderProtectionKey(raw)
		case awgV3Bool:
			var b bool
			b, err = strconv.ParseBool(raw)
			value = strconv.FormatBool(b)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", p.confKey, err)
		}
		if out == nil {
			out = make(map[string]string)
		}
		out[p.uapiKey] = value
	}
	return out, nil
}

// normalizeUintRange validates "a" or "a-b" (a <= b, uint32) and returns it.
func normalizeUintRange(s string) (string, error) {
	lo, hi, err := parseMagicHeaderInterval(s)
	if err != nil {
		return "", err
	}
	return formatMagicHeaderInterval(lo, hi), nil
}

// normalizeHeaderProtectionKey accepts a 32-byte key in hex or base64 (as
// printed by `awg genkey`) and returns it hex encoded for the UAPI.
func normalizeHeaderProtectionKey(s string) (string, error) {
	if len(s) == 64 {
		if _, err := hex.DecodeString(s); err == nil {
			return strings.ToLower(s), nil
		}
	}
	decoded, err := encodeBase64ToHex(s)
	if err != nil {
		return "", errors.New("expected a 32-byte key in hex or base64")
	}
	return decoded, nil
}
