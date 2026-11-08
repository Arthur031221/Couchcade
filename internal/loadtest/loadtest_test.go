package loadtest

import (
	"context"
	"testing"
)

func TestHundredPhonesStayResponsive(t *testing.T) {
	if testing.Short() {
		t.Skip("takes about eight seconds")
	}
	res, err := Measure(context.Background(), Options{Phones: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FanOut) != 100 {
		t.Fatalf("only %d phones saw the question", len(res.FanOut))
	}
	if p99 := res.Percentile(res.Ping, 99); p99 > 500 {
		t.Fatalf("ping p99 %.1f ms", p99)
	}
	if p99 := res.Percentile(res.FanOut, 99); p99 > 1000 {
		t.Fatalf("fan-out p99 %.1f ms", p99)
	}
}
