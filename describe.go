package wireproxy

import (
	"fmt"
	"io"
)

// SetLogOutput redirects wireproxy's own error log (the amneziawg-go device
// logger is configured separately through os.Stdout).
func SetLogOutput(w io.Writer) {
	errorLogger.SetOutput(w)
}

// DescribeRoutine returns a short human readable summary of a configured
// proxy or tunnel, for status pages.
func DescribeRoutine(r RoutineSpawner) string {
	switch c := r.(type) {
	case *Socks5Config:
		s := "SOCKS5 " + c.BindAddress
		if c.Username != "" {
			s += " (auth)"
		}
		if len(c.TunnelDomains) > 0 {
			s += fmt.Sprintf(" [%d tunnel domain rule(s)]", len(c.TunnelDomains))
		}
		return s
	case *HTTPConfig:
		s := "HTTP " + c.BindAddress
		if c.CertFile != "" {
			s = "HTTPS " + c.BindAddress
		}
		if c.Username != "" {
			s += " (auth)"
		}
		if len(c.TunnelDomains) > 0 {
			s += fmt.Sprintf(" [%d tunnel domain rule(s)]", len(c.TunnelDomains))
		}
		return s
	case *SNIConfig:
		return "SNI proxy " + c.BindAddress
	case *TCPClientTunnelConfig:
		return fmt.Sprintf("TCP client tunnel %s -> %s", c.BindAddress, c.Target)
	case *TCPServerTunnelConfig:
		return fmt.Sprintf("TCP server tunnel :%d -> %s", c.ListenPort, c.Target)
	case *UDPProxyTunnelConfig:
		return fmt.Sprintf("UDP tunnel %s -> %s", c.BindAddress, c.Target)
	case *STDIOTunnelConfig:
		return "STDIO tunnel -> " + c.Target
	default:
		return fmt.Sprintf("%T", r)
	}
}
