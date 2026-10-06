package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSolveSharesConcurrentResult(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		failed bool
	}{
		{"success", http.StatusOK, `{"status":"ok","solution":{"userAgent":"test-agent","cookies":[{"name":"cf_clearance","value":"test-cookie"}]}}`, false},
		{"solver error", http.StatusOK, `{"status":"error","message":"challenge failed"}`, true},
		{"HTTP error", http.StatusInternalServerError, `{"status":"error","message":"solver unavailable"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls int32
			entered := make(chan struct{})
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v1" {
					t.Errorf("unexpected solver request: %s %s", r.Method, r.URL.Path)
				}
				if atomic.AddInt32(&calls, 1) == 1 {
					close(entered)
					<-release
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			defer unblock()
			t.Setenv("FLARESOLVERR_URL", server.URL)

			host := t.Name()
			stale := &clearance{userAgent: "old-agent"}
			clearancesMu.Lock()
			clearances[host] = stale
			clearancesMu.Unlock()
			defer func() {
				clearancesMu.Lock()
				delete(clearances, host)
				clearancesMu.Unlock()
			}()

			type result struct {
				cl  *clearance
				err error
			}
			const requests = 8
			results := make(chan result, requests)
			start := make(chan struct{})
			for i := 0; i < requests; i++ {
				go func() {
					<-start
					cl, err := solve(host, "https://images.example/image.png", stale)
					results <- result{cl, err}
				}()
			}
			close(start)
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("solver request did not start")
			}
			// Callers must wait while the shared solver request is in progress.
			select {
			case <-results:
				t.Fatal("solve returned before the solver responded")
			case <-time.After(100 * time.Millisecond):
			}
			unblock()

			var first result
			for i := 0; i < requests; i++ {
				select {
				case got := <-results:
					if (got.err != nil) != tc.failed {
						t.Fatalf("unexpected solve result: clearance=%v error=%v", got.cl, got.err)
					}
					if i == 0 {
						first = got
					} else if got.cl != first.cl || got.err != first.err {
						t.Fatal("concurrent callers did not share the same result")
					}
				case <-time.After(5 * time.Second):
					t.Fatal("waiting caller did not finish")
				}
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("solver received %d requests, want 1", got)
			}
			if !tc.failed && (first.cl.userAgent != "test-agent" || len(first.cl.cookies) != 1 || first.cl.cookies[0].Value != "test-cookie") {
				t.Fatalf("unexpected clearance: %+v", first.cl)
			}

			cl, err := solve(host, "https://images.example/another.png", stale)
			if tc.failed {
				if err == nil || atomic.LoadInt32(&calls) != 2 {
					t.Fatal("a later request should retry after a failed solve")
				}
			} else if err != nil || cl != first.cl || atomic.LoadInt32(&calls) != 1 {
				t.Fatal("a later request with stale clearance should reuse the refreshed clearance")
			}
		})
	}
}
