package middleware

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (f *fakeClock) now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.t
}

func (f *fakeClock) advance(d time.Duration) {
	f.mu.Lock()
	f.t = f.t.Add(d)
	f.mu.Unlock()
}

func limitedEngine(l *limiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("home.html").Parse("limited")))
	r.Use(ratelimit(l))
	r.GET("/version", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func request(r *gin.Engine, ip string) int {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	req.RemoteAddr = ip + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRatelimitLimitsAndRecovers(t *testing.T) {
	clock := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	r := limitedEngine(newLimiter(clock.now))

	for i := 1; i <= ratelimitRequests; i++ {
		if code := request(r, "192.0.2.1"); code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i, code)
		}
	}
	// Overshoot well past the limit, as a script or a shared IP would.
	for i := 0; i < 70; i++ {
		if code := request(r, "192.0.2.1"); code != http.StatusTooManyRequests {
			t.Fatalf("over-limit request: status %d, want 429", code)
		}
	}
	if code := request(r, "192.0.2.2"); code != http.StatusOK {
		t.Fatalf("other client limited too: status %d", code)
	}

	clock.advance(ratelimitWindow)
	if code := request(r, "192.0.2.1"); code != http.StatusOK {
		t.Fatalf("client still limited after the window: status %d", code)
	}
}

func TestRatelimitForgetsOldClients(t *testing.T) {
	clock := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	l := newLimiter(clock.now)
	r := limitedEngine(l)

	for i := 0; i < 5000; i++ {
		request(r, fmt.Sprintf("198.51.%d.%d", i/256, i%256))
	}
	l.mu.Lock()
	seen := len(l.counts)
	l.mu.Unlock()
	if seen != 5000 {
		t.Fatalf("limiter tracked %d clients, want 5000", seen)
	}

	clock.advance(ratelimitWindow)
	request(r, "192.0.2.1")

	l.mu.Lock()
	n := len(l.counts)
	l.mu.Unlock()
	if n != 1 {
		t.Fatalf("limiter holds %d clients after the window, want 1", n)
	}
}

func TestRatelimitConcurrent(t *testing.T) {
	r := limitedEngine(newLimiter(time.Now))
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			request(r, "192.0.2.9")
		}()
	}
	wg.Wait()
}
