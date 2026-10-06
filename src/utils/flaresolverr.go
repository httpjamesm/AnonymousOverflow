package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
)

// clearance is a solved Cloudflare challenge. cf_clearance is bound to the
// hostname and to the user agent that solved it, so both are kept per host.
type clearance struct {
	userAgent string
	cookies   []*http.Cookie
}

type flaresolverrResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		UserAgent string `json:"userAgent"`
		Cookies   []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"cookies"`
	} `json:"solution"`
}

var (
	clearancesMu sync.Mutex
	clearances   = map[string]*clearance{}

	// Concurrent requests for a host share the solve result, including errors.
	solves singleflight.Group

	// solveMu limits FlareSolverr to one solve at a time.
	solveMu sync.Mutex
)

const solveTimeout = 60 * time.Second

// GetWithClearance fetches target, solving a Cloudflare challenge through
// FlareSolverr (FLARESOLVERR_URL) when the host answers with 403.
func GetWithClearance(target string) (*resty.Response, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	host := u.Hostname()

	clearancesMu.Lock()
	cl := clearances[host]
	clearancesMu.Unlock()

	res, err := getWith(target, cl)
	if err != nil || res.StatusCode() != http.StatusForbidden || os.Getenv("FLARESOLVERR_URL") == "" {
		return res, err
	}

	cl, err = solve(host, target, cl)
	if err != nil {
		return nil, err
	}

	return getWith(target, cl)
}

func getWith(target string, cl *clearance) (*resty.Response, error) {
	req := resty.New().R()
	if cl != nil {
		req.SetHeader("User-Agent", cl.userAgent).SetCookies(cl.cookies)
	}
	return req.Get(target)
}

func solve(host, target string, stale *clearance) (*clearance, error) {
	result, err, _ := solves.Do(host, func() (interface{}, error) {
		solveMu.Lock()
		defer solveMu.Unlock()

		clearancesMu.Lock()
		current := clearances[host]
		clearancesMu.Unlock()
		// Another request refreshed the clearance while this one waited.
		if current != stale {
			return current, nil
		}

		endpoint, err := url.JoinPath(os.Getenv("FLARESOLVERR_URL"), "v1")
		if err != nil {
			return nil, err
		}

		var fsRes flaresolverrResponse
		// maxTimeout only bounds the solve inside FlareSolverr; the HTTP timeout
		// keeps a hung FlareSolverr from holding solveMu forever.
		res, err := resty.New().SetTimeout(solveTimeout + 30*time.Second).R().
			SetBody(map[string]any{
				"cmd":               "request.get",
				"url":               target,
				"maxTimeout":        solveTimeout.Milliseconds(),
				"returnOnlyCookies": true,
			}).
			SetResult(&fsRes).
			SetError(&fsRes).
			Post(endpoint)
		if err != nil {
			return nil, fmt.Errorf("flaresolverr request failed: %w", err)
		}
		if res.StatusCode() != http.StatusOK || fsRes.Status != "ok" {
			return nil, fmt.Errorf("flaresolverr failed: %d %s", res.StatusCode(), fsRes.Message)
		}

		cl := &clearance{userAgent: fsRes.Solution.UserAgent}
		for _, c := range fsRes.Solution.Cookies {
			cl.cookies = append(cl.cookies, &http.Cookie{Name: c.Name, Value: c.Value})
		}

		clearancesMu.Lock()
		clearances[host] = cl
		clearancesMu.Unlock()

		return cl, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*clearance), nil
}
