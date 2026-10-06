package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/singleflight"
)

func TestSolveSharesConcurrentResult(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		failed bool
	}{
		{"success", http.StatusOK, `{"status":"ok","solution":{"url":"https://images.example/image.png","userAgent":"test-agent","cookies":[{"name":"cf_clearance","value":"test-cookie","secure":true}]}}`, false},
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
			firstResult := make(chan result, 1)
			var worker sync.WaitGroup
			worker.Add(1)
			defer func() {
				unblock()
				worker.Wait()
			}()
			go func() {
				defer worker.Done()
				cl, err := solve(host, "https://images.example/image.png", stale)
				firstResult <- result{cl, err}
			}()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("solver request did not start")
			}

			// DoChan registers each waiter before returning, without scheduler delays.
			var waiters []<-chan singleflight.Result
			for i := 0; i < 7; i++ {
				waiters = append(waiters, solves.DoChan(host, func() (interface{}, error) {
					return nil, fmt.Errorf("waiting request started another solve")
				}))
			}
			unblock()

			var first result
			select {
			case first = <-firstResult:
			case <-time.After(5 * time.Second):
				t.Fatal("solve did not finish")
			}
			if (first.err != nil) != tc.failed {
				t.Fatalf("unexpected solve result: clearance=%v error=%v", first.cl, first.err)
			}
			for _, waiter := range waiters {
				select {
				case got := <-waiter:
					if !got.Shared || got.Err != first.err || (got.Err == nil && got.Val != first.cl) {
						t.Fatal("concurrent callers did not share the same result")
					}
				case <-time.After(5 * time.Second):
					t.Fatal("waiting caller did not finish")
				}
			}
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("solver received %d requests, want 1", got)
			}
			if !tc.failed && (first.cl.userAgent != "test-agent" || len(first.cl.cookies) != 1 || first.cl.cookies[0].Value != "test-cookie" || !first.cl.cookies[0].Secure) {
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

func TestGetWithClearanceAcrossHosts(t *testing.T) {
	var generation int32 = 1
	var solverCalls int32
	image := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/image.png" {
			t.Errorf("unexpected image path: %s", r.URL.Path)
		}
		expectedCookie := fmt.Sprintf("cf_clearance=v%d", atomic.LoadInt32(&generation))
		if r.Header.Get("Cookie") != expectedCookie || r.UserAgent() != "test-agent" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		fmt.Fprint(w, "image bytes")
	}))
	defer image.Close()
	finalURL := strings.Replace(image.URL, "127.0.0.1", "localhost", 1) + "/images/image.png"
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie := r.Header.Get("Cookie"); cookie != "" {
			t.Errorf("clearance leaked to the original host: %s", cookie)
		}
		http.Redirect(w, r, finalURL, http.StatusFound)
	}))
	defer origin.Close()

	solver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&solverCalls, 1)
		var request struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.URL != finalURL {
			t.Errorf("solver URL = %q, want %q", request.URL, finalURL)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","solution":{"url":%q,"userAgent":"test-agent","cookies":[
			{"name":"cf_clearance","value":"v%d","domain":"localhost","path":"/images","secure":false},
			{"name":"other_domain","value":"hidden","domain":"example.invalid","path":"/"},
			{"name":"other_path","value":"hidden","domain":"localhost","path":"/private"}
		]}}`, finalURL, atomic.LoadInt32(&generation))
	}))
	defer solver.Close()
	t.Setenv("FLARESOLVERR_URL", solver.URL)
	defer func() {
		clearancesMu.Lock()
		delete(clearances, "localhost")
		clearancesMu.Unlock()
	}()

	for _, tc := range []struct {
		name       string
		generation int32
		calls      int32
	}{
		{"initial solve", 1, 1},
		{"cached clearance", 1, 1},
		{"expired clearance", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			atomic.StoreInt32(&generation, tc.generation)
			res, err := GetWithClearance(origin.URL + "/old.png")
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode() != http.StatusOK || string(res.Body()) != "image bytes" {
				t.Fatalf("image status=%d body=%q", res.StatusCode(), res.Body())
			}
			if calls := atomic.LoadInt32(&solverCalls); calls != tc.calls {
				t.Fatalf("solver calls=%d, want %d", calls, tc.calls)
			}
		})
	}

	clearancesMu.Lock()
	cl := clearances["localhost"]
	clearancesMu.Unlock()
	res, err := getWith(origin.URL+"/old.png", cl)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() != http.StatusOK {
		t.Fatalf("cookie jar did not apply clearance after redirect: %d", res.StatusCode())
	}
}
