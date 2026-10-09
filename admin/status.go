package admin

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
)

// PeerStatus is the runtime state of a peer.
type PeerStatus struct {
	PublicKey     string   `json:"public_key"`
	Endpoint      string   `json:"endpoint,omitempty"`
	AllowedIPs    []string `json:"allowed_ips"`
	LastHandshake int64    `json:"last_handshake"` // unix seconds, 0 = never
	RxBytes       uint64   `json:"rx_bytes"`
	TxBytes       uint64   `json:"tx_bytes"`
	Keepalive     string   `json:"keepalive,omitempty"`
}

// ParsePeers extracts peer information from a UAPI "get" dump. Secret values
// (private and preshared keys) are never read.
func ParsePeers(ipc string) []PeerStatus {
	peers := []PeerStatus{}
	var cur *PeerStatus
	for _, line := range strings.Split(ipc, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if k == "public_key" {
			peers = append(peers, PeerStatus{PublicKey: hexToBase64(v), AllowedIPs: []string{}})
			cur = &peers[len(peers)-1]
			continue
		}
		if cur == nil {
			continue
		}
		switch k {
		case "endpoint":
			cur.Endpoint = v
		case "allowed_ip":
			cur.AllowedIPs = append(cur.AllowedIPs, v)
		case "last_handshake_time_sec":
			cur.LastHandshake, _ = strconv.ParseInt(v, 10, 64)
		case "rx_bytes":
			cur.RxBytes, _ = strconv.ParseUint(v, 10, 64)
		case "tx_bytes":
			cur.TxBytes, _ = strconv.ParseUint(v, 10, 64)
		case "persistent_keepalive_interval":
			if v != "0" {
				cur.Keepalive = v
			}
		}
	}
	return peers
}

func hexToBase64(s string) string {
	b, err := hex.DecodeString(s)
	if err != nil {
		return s
	}
	return base64.StdEncoding.EncodeToString(b)
}
