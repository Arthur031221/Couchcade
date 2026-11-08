// Package loadtest drives a server with many simulated phones and reports latency.
package loadtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/Arthur031221/Couchcade/internal/content"
	"github.com/Arthur031221/Couchcade/internal/server"
)

type Options struct {
	Phones int
	URL    string // ws://host:port of a running server; empty starts one in process
	Out    io.Writer
}

// Result holds the measurements of one run.
type Result struct {
	Phones, Seated int
	Join           time.Duration
	FanOut         []time.Duration
	Ping           []time.Duration
}

type client struct {
	c     *websocket.Conn
	mu    sync.Mutex
	sent  map[int]time.Time
	rtt   []time.Duration
	ask   time.Time
	ready chan struct{}
	code  string
	tok   string
}

func dial(ctx context.Context, base string) (*client, error) {
	c, _, err := websocket.Dial(ctx, base+"/ws", nil)
	if err != nil {
		return nil, err
	}
	c.SetReadLimit(1 << 20)
	return &client{c: c, sent: map[int]time.Time{}, ready: make(chan struct{})}, nil
}

func (cl *client) send(ctx context.Context, v any) error {
	b, _ := json.Marshal(v)
	return cl.c.Write(ctx, websocket.MessageText, b)
}

func (cl *client) read(ctx context.Context) {
	for {
		_, data, err := cl.c.Read(ctx)
		if err != nil {
			return
		}
		var m struct {
			T    string `json:"t"`
			N    int    `json:"n"`
			Code string `json:"code"`
			G    struct {
				Phase string `json:"phase"`
			} `json:"g"`
		}
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		now := time.Now()
		cl.mu.Lock()
		switch m.T {
		case "tvok", "joined":
			cl.code = m.Code
			select {
			case <-cl.ready:
			default:
				close(cl.ready)
			}
		case "pong":
			if t0, ok := cl.sent[m.N]; ok {
				cl.rtt = append(cl.rtt, now.Sub(t0))
				delete(cl.sent, m.N)
			}
		case "state":
			if m.G.Phase == "ask" && cl.ask.IsZero() {
				cl.ask = now
			}
		}
		cl.mu.Unlock()
	}
}

func wait(ctx context.Context, ch <-chan struct{}, d time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(d):
		return errors.New("timed out")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run executes the scenario and prints a report.
func Run(ctx context.Context, opt Options) error {
	res, err := Measure(ctx, opt)
	if err != nil {
		return err
	}
	out := opt.Out
	if out == nil {
		out = io.Discard
	}
	fmt.Fprintf(out, "phones          %d (%d players, %d audience)\n", res.Phones, res.Seated, res.Phones-res.Seated)
	fmt.Fprintf(out, "join            all connected in %d ms\n", res.Join.Milliseconds())
	fmt.Fprintf(out, "state fan-out   p50 %s  p99 %s  (host starts a game, last phone shows the question)\n", pct(res.FanOut, 50), pct(res.FanOut, 99))
	fmt.Fprintf(out, "ping round trip p50 %s  p99 %s  (%d pings while the round runs)\n", pct(res.Ping, 50), pct(res.Ping, 99), len(res.Ping))
	return nil
}

func pct(ds []time.Duration, p int) string {
	if len(ds) == 0 {
		return "n/a"
	}
	s := append([]time.Duration(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	d := s[min(len(s)-1, len(s)*p/100)]
	return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000)
}

// Percentile returns the p-th percentile of ds in milliseconds.
func (r Result) Percentile(ds []time.Duration, p int) float64 {
	if len(ds) == 0 {
		return 0
	}
	s := append([]time.Duration(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return float64(s[min(len(s)-1, len(s)*p/100)].Microseconds()) / 1000
}

// Measure runs the scenario and returns raw numbers.
func Measure(ctx context.Context, opt Options) (Result, error) {
	var res Result
	base := opt.URL
	if base == "" {
		set, err := content.Builtin()
		if err != nil {
			return res, err
		}
		srv := server.New(server.Options{Content: set, PublicURL: "http://localhost"})
		rctx, stop := context.WithCancel(ctx)
		defer stop()
		go srv.Run(rctx)
		hs := httptest.NewServer(srv.Handler())
		defer hs.Close()
		base = "ws" + strings.TrimPrefix(hs.URL, "http")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	tv, err := dial(ctx, base)
	if err != nil {
		return res, err
	}
	go tv.read(ctx)
	if err := tv.send(ctx, map[string]string{"t": "tv"}); err != nil {
		return res, err
	}
	if err := wait(ctx, tv.ready, 5*time.Second); err != nil {
		return res, fmt.Errorf("tv: %w", err)
	}

	t0 := time.Now()
	phones := make([]*client, opt.Phones)
	var wg sync.WaitGroup
	errs := make(chan error, opt.Phones)
	sem := make(chan struct{}, 25)
	for i := range phones {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			cl, err := dial(ctx, base)
			if err != nil {
				errs <- err
				return
			}
			go cl.read(ctx)
			_ = cl.send(ctx, map[string]string{"t": "join", "code": tv.code, "name": fmt.Sprintf("Phone%d", i)})
			if err := wait(ctx, cl.ready, 10*time.Second); err != nil {
				errs <- fmt.Errorf("phone %d: %w", i, err)
				return
			}
			phones[i] = cl
		}(i)
	}
	wg.Wait()
	close(errs)
	if err := <-errs; err != nil {
		return res, err
	}
	res.Join = time.Since(t0)
	res.Phones, res.Seated = opt.Phones, min(opt.Phones, 8)

	start := time.Now()
	_ = tv.send(ctx, map[string]any{"t": "host", "op": "start", "game": "trivia", "rounds": 1})
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		missing := false
		for _, p := range phones {
			p.mu.Lock()
			if p.ask.IsZero() {
				missing = true
			}
			p.mu.Unlock()
		}
		if !missing {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	for i, p := range phones {
		p.mu.Lock()
		if p.ask.IsZero() {
			p.mu.Unlock()
			return res, fmt.Errorf("phone %d never saw the question", i)
		}
		res.FanOut = append(res.FanOut, p.ask.Sub(start))
		p.mu.Unlock()
	}

	var pw sync.WaitGroup
	for i, p := range phones {
		pw.Add(1)
		go func(i int, p *client) {
			defer pw.Done()
			rng := rand.New(rand.NewSource(int64(i)))
			answerAt := time.Now().Add(time.Duration(rng.Intn(2000)) * time.Millisecond)
			answered := false
			for n := 0; n < 24; n++ {
				p.mu.Lock()
				p.sent[n] = time.Now()
				p.mu.Unlock()
				_ = p.send(ctx, map[string]any{"t": "ping", "n": n})
				if !answered && time.Now().After(answerAt) {
					_ = p.send(ctx, map[string]any{"t": "act", "a": map[string]any{"k": "answer", "i": rng.Intn(4)}})
					answered = true
				}
				time.Sleep(250 * time.Millisecond)
			}
		}(i, p)
	}
	pw.Wait()
	for _, p := range phones {
		p.mu.Lock()
		res.Ping = append(res.Ping, p.rtt...)
		p.mu.Unlock()
		_ = p.c.Close(websocket.StatusNormalClosure, "")
	}
	_ = tv.c.Close(websocket.StatusNormalClosure, "")
	return res, nil
}
