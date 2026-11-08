// Command couchcade serves party games for one TV and everyone's phones.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/loadtest"
	"github.com/Arthur031221/Couchcade/internal/server"
)

var version = "dev"

const usage = `Couchcade: party games for one TV and everyone's phones.

Usage:
  couchcade [serve] [-addr :8080] [-public-url URL] [-packs DIR] [-max-players 8]
  couchcade packs validate FILE...
  couchcade packs stats [-packs DIR]
  couchcade loadtest [-phones 100] [-url ws://host:port]
  couchcade version
`

func main() {
	args := os.Args[1:]
	cmd := "serve"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}
	var err error
	switch cmd {
	case "serve":
		err = serve(args)
	case "packs":
		err = packs(args)
	case "loadtest":
		err = runLoadtest(args)
	case "version":
		fmt.Println("couchcade", version)
	case "help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "couchcade:", err)
		os.Exit(1)
	}
}

func loadContent(dir string) (*content.Set, error) {
	set, err := content.Builtin()
	if err != nil {
		return nil, err
	}
	if dir != "" {
		if err := set.LoadDir(dir); err != nil {
			return nil, err
		}
	}
	return set, nil
}

func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	pub := fs.String("public-url", "", "address phones use to reach this server (default: LAN address)")
	dir := fs.String("packs", "", "folder of extra content packs (*.json)")
	maxP := fs.Int("max-players", 8, "seats per room, everyone else joins as audience")
	origins := fs.String("allow-origin", "", "comma separated hosts allowed to open WebSockets behind a proxy")
	scale := fs.Float64("time-scale", 1, "run game clocks faster than real time (for tests)")
	_ = fs.Parse(args)

	set, err := loadContent(*dir)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	public := *pub
	if public == "" {
		host := server.LANAddress()
		if host == "" {
			host = "localhost"
		}
		public = fmt.Sprintf("http://%s:%d", host, port)
	}
	var patterns []string
	for _, o := range strings.Split(*origins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			patterns = append(patterns, o)
		}
	}
	srv := server.New(server.Options{Content: set, PublicURL: public, MaxPlayers: *maxP, Version: version, OriginPatterns: patterns, TimeScale: *scale})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go srv.Run(ctx)

	hs := &http.Server{Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = hs.Shutdown(shutdown)
	}()
	fmt.Printf("Couchcade %s\n  Open this on the TV:  %s\n  Phones join at:       %s/play\n", version, public, public)
	if err := hs.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func packs(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: couchcade packs validate FILE... | stats")
	}
	switch args[0] {
	case "validate":
		bad := 0
		for _, f := range args[1:] {
			data, err := os.ReadFile(f)
			if err == nil {
				_, err = content.Parse(data)
			}
			if err != nil {
				fmt.Printf("FAIL %s: %v\n", f, err)
				bad++
				continue
			}
			fmt.Printf("ok   %s\n", f)
		}
		if bad > 0 {
			return fmt.Errorf("%d pack(s) failed", bad)
		}
		return nil
	case "stats":
		fs := flag.NewFlagSet("stats", flag.ExitOnError)
		dir := fs.String("packs", "", "folder of extra content packs")
		_ = fs.Parse(args[1:])
		set, err := loadContent(*dir)
		if err != nil {
			return err
		}
		counts := set.Counts()
		fmt.Printf("%-6s %8s %7s %7s %5s %5s\n", "lang", "squiggle", "fibber", "trivia", "echo", "mole")
		for _, l := range content.Languages {
			c := counts[l]
			fmt.Printf("%-6s %8d %7d %7d %5d %5d\n", l, c["squiggle"], c["fibber"], c["trivia"], c["echo"], c["mole"])
		}
		return nil
	}
	return fmt.Errorf("unknown packs command %q", args[0])
}

func runLoadtest(args []string) error {
	fs := flag.NewFlagSet("loadtest", flag.ExitOnError)
	phones := fs.Int("phones", 100, "simulated phones")
	url := fs.String("url", "", "WebSocket base of a running server, for example ws://192.168.1.20:8080 (default: start one in process)")
	_ = fs.Parse(args)
	return loadtest.Run(context.Background(), loadtest.Options{Phones: *phones, URL: *url, Out: os.Stdout})
}
