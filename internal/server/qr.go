package server

import (
	"fmt"
	"net/http"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

// serveQR renders the join address of a room as an SVG QR code.
func (s *Server) serveQR(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(r.URL.Query().Get("c"))
	if len(code) != 4 || strings.Trim(code, "ABCDEFGHJKLMNPQRSTUVWXYZ") != "" {
		http.Error(w, "bad code", http.StatusBadRequest)
		return
	}
	svg, err := QRSVG(strings.TrimRight(s.opt.PublicURL, "/") + "/play?c=" + code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(svg))
}

// QRSVG returns a scannable black on white SVG with a two module quiet zone.
func QRSVG(text string) (string, error) {
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return "", err
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	n, pad := len(bm), 2
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" shape-rendering="crispEdges"><rect width="100%%" height="100%%" fill="#fff"/><path fill="#000" d="`, n+2*pad, n+2*pad)
	for y, row := range bm {
		for x := 0; x < len(row); {
			if !row[x] {
				x++
				continue
			}
			start := x
			for x < len(row) && row[x] {
				x++
			}
			fmt.Fprintf(&b, "M%d %dh%dv1h-%dz", start+pad, y+pad, x-start, x-start)
		}
	}
	b.WriteString(`"/></svg>`)
	return b.String(), nil
}
