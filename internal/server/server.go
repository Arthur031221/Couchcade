// Package server exposes the engine over HTTP and WebSockets.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/engine"
	"github.com/Arthur031221/Couchcade/internal/games"
	"github.com/Arthur031221/Couchcade/web"
)

// Options configure a Server.
type Options struct {
	Content    *content.Set
	PublicURL  string // address phones use, for example http://192.168.1.20:8080
	MaxPlayers int
	Version    string
	// OriginPatterns lists extra hosts allowed to open WebSockets, for reverse proxies.
	OriginPatterns []string
	// TimeScale runs game clocks faster than real time. Tests use it, players never need to.
	TimeScale float64
}

type Server struct {
	opt Options
	mgr *engine.Manager
	mux *http.ServeMux
}

func New(opt Options) *Server {
	s := &Server{opt: opt, mux: http.NewServeMux()}
	s.mgr = engine.NewManager(engine.Config{
		Content: opt.Content, Games: games.New, Catalog: games.Catalog, MaxPlayers: opt.MaxPlayers,
		JoinURL: func(code string) string { return strings.TrimRight(opt.PublicURL, "/") + "/play?c=" + code },
		Clock:   scaledClock(opt.TimeScale),
	})
	s.routes()
	return s
}

func (s *Server) Manager() *engine.Manager { return s.mgr }

func (s *Server) Handler() http.Handler { return secure(s.mux) }

// Run ticks the rooms until ctx ends.
func (s *Server) Run(ctx context.Context) { s.mgr.Run(ctx) }

func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:; frame-ancestors 'self'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	files := http.FileServerFS(web.FS)
	page := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			b, err := fs.ReadFile(web.FS, name)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(b)
		}
	}
	s.mux.HandleFunc("GET /{$}", page("index.html"))
	s.mux.HandleFunc("GET /play", page("play.html"))
	s.mux.Handle("GET /logo.svg", files)
	s.mux.Handle("GET /css/", files)
	s.mux.Handle("GET /js/", files)
	s.mux.Handle("GET /locales/", files)
	s.mux.HandleFunc("GET /ws", s.serveWS)
	s.mux.HandleFunc("GET /api/qr", s.serveQR)
	s.mux.HandleFunc("GET /api/info", s.serveInfo)
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "ok rooms=%d\n", s.mgr.RoomCount())
	})
}

func (s *Server) serveInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"version": s.opt.Version, "join": s.opt.PublicURL, "languages": content.Languages})
}

type wsConn struct {
	c      *websocket.Conn
	out    chan []byte
	cancel context.CancelFunc
}

func (w *wsConn) Send(b []byte) {
	cp := append([]byte(nil), b...)
	select {
	case w.out <- cp:
	default:
		// A phone that cannot keep up is dropped; it reconnects with its token.
		w.cancel()
	}
}

func (w *wsConn) Close() {
	go func() {
		_ = w.c.Close(websocket.StatusPolicyViolation, "closed by room")
		w.cancel()
	}()
}

func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.opt.OriginPatterns})
	if err != nil {
		return
	}
	c.SetReadLimit(256 << 10)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	conn := &wsConn{c: c, out: make(chan []byte, 64), cancel: cancel}
	sess := engine.NewSession(conn)
	defer s.mgr.Close(sess)
	defer c.CloseNow()

	go func() {
		tick := time.NewTicker(25 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-conn.out:
				wctx, wcancel := context.WithTimeout(ctx, 10*time.Second)
				err := c.Write(wctx, websocket.MessageText, b)
				wcancel()
				if err != nil {
					cancel()
					return
				}
			case <-tick.C:
				pctx, pcancel := context.WithTimeout(ctx, 10*time.Second)
				err := c.Ping(pctx)
				pcancel()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return
		}
		s.mgr.Handle(sess, data)
	}
}

// LANAddress returns the first private IPv4 address of this machine.
func LANAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil && n.IP.IsPrivate() {
			return n.IP.String()
		}
	}
	return ""
}

func scaledClock(scale float64) func() time.Time {
	if scale <= 1 {
		return time.Now
	}
	t0 := time.Now()
	return func() time.Time {
		return t0.Add(time.Duration(float64(time.Since(t0)) * scale))
	}
}
