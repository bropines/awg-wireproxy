package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/amnezia-vpn/amneziawg-go/v3/device"
	wireproxy "github.com/bropines/awg-wireproxy"
	"github.com/bropines/awg-wireproxy/admin"
)

type adminOptions struct {
	configPath string
	addr       string
	token      string
	silent     bool
	infoAddr   string
}

// captureLogs mirrors everything the daemon prints into buf while keeping the
// original stderr output.
func captureLogs(buf *admin.LogBuffer) {
	orig := os.Stderr
	out := io.MultiWriter(orig, buf)

	log.SetOutput(out)
	wireproxy.SetLogOutput(out)

	// The amneziawg-go logger writes to os.Stdout at creation time.
	r, w, err := os.Pipe()
	if err != nil {
		os.Stdout = orig
		return
	}
	os.Stdout = w
	go func() { _, _ = io.Copy(out, r) }()
}

// runAdmin runs the daemon together with the web panel. Unlike the plain mode
// it keeps running when the configuration is broken, so the panel can be used
// to repair it, and it does not apply the landlock/pledge sandbox because the
// process has to write the config file and restart itself.
func runAdmin(ctx context.Context, opts adminOptions) {
	if abs, err := filepath.Abs(opts.configPath); err == nil {
		opts.configPath = abs
	}
	logs := admin.NewLogBuffer(2000)
	captureLogs(logs)

	if opts.token == "" {
		opts.token = admin.GenerateToken()
		fmt.Fprintf(os.Stderr, "admin: no token given, generated one for this run: %s\n", opts.token)
	}

	var (
		tun      atomic.Pointer[wireproxy.VirtualTun]
		startErr atomic.Value // string
		listen   atomic.Value // []string
	)
	startErr.Store("")
	listen.Store([]string{})

	logLevel := device.LogLevelVerbose
	if opts.silent {
		logLevel = device.LogLevelSilent
	}

	if _, err := os.Stat(opts.configPath); errors.Is(err, os.ErrNotExist) {
		msg := fmt.Sprintf("config file %s does not exist yet; create it in the panel", opts.configPath)
		startErr.Store(msg)
		log.Print(msg)
	} else if conf, err := wireproxy.ParseConfig(opts.configPath); err != nil {
		startErr.Store("invalid configuration: " + err.Error())
		log.Printf("invalid configuration: %v", err)
	} else if vt, err := wireproxy.StartWireguard(conf, logLevel); err != nil {
		startErr.Store("cannot start the tunnel: " + err.Error())
		log.Printf("cannot start the tunnel: %v", err)
	} else {
		tun.Store(vt)
		var desc []string
		for _, spawner := range conf.Routines {
			desc = append(desc, wireproxy.DescribeRoutine(spawner))
			go spawner.SpawnRoutine(vt)
		}
		listen.Store(desc)
		vt.StartPingIPs()
		if opts.infoAddr != "" {
			go func() {
				if err := http.ListenAndServe(opts.infoAddr, vt); err != nil {
					log.Printf("info server: %v", err)
				}
			}()
		}
	}

	srv := admin.New(admin.Options{
		ConfigPath: opts.configPath,
		Token:      opts.token,
		Version:    version,
		Logs:       logs,
		Tun:        tun.Load,
		StartError: func() string { return startErr.Load().(string) },
		Listeners:  func() []string { return listen.Load().([]string) },
		Reload:     reload,
	})

	ln := listenRetry(opts.addr)
	log.Printf("admin panel listening on http://%s", ln.Addr())
	if host, _, err := net.SplitHostPort(opts.addr); err == nil {
		if ip := net.ParseIP(host); host == "" || (ip != nil && !ip.IsLoopback()) {
			log.Print("admin: WARNING the panel is reachable from the network over plain HTTP; put it behind a TLS reverse proxy")
		}
	}
	go func() {
		hs := &http.Server{Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
		if err := hs.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("admin server: %v", err)
		}
	}()

	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for range hup {
			log.Print("SIGHUP received, reloading")
			reload()
		}
	}()

	<-ctx.Done()
}

func reload() {
	log.Print("restarting to apply the configuration")
	if err := reexec(); err != nil {
		log.Printf("restart failed: %v", err)
	}
}

// listenRetry keeps trying for a few seconds: after a restart on platforms
// without exec the old process may still hold the port for a moment.
func listenRetry(addr string) net.Listener {
	var lastErr error
	for i := 0; i < 20; i++ {
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			return ln
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	log.Fatalf("admin: cannot listen on %s: %v", addr, lastErr)
	return nil
}
