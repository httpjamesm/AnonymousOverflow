package routes

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func exchangeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetHTMLTemplate(template.Must(template.New("home.html").Parse(`{{ .errorMessage }}`)))
	r.Use(func(c *gin.Context) { c.Set("disable_images", false) })
	r.GET("/exchange/:sub/questions/:id/:title", ViewQuestion)
	r.GET("/exchange/:sub/q/:id", RedirectShortenedOverflowURL)
	return r
}

func TestExchangeRoutesRejectNonExchangeHosts(t *testing.T) {
	// Stands in for an internal service; counts incoming connections.
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	var connections int32
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			atomic.AddInt32(&connections, 1)
			conn.Close()
		}
	}()
	port := l.Addr().(*net.TCPAddr).Port

	paths := []string{
		// %23 cuts off the ".stackexchange.com" the HTML scraper appended.
		fmt.Sprintf("/exchange/%%5B::ffff:7f00:1%%5D:%d%%23/questions/1/x", port),
		// The shortener used any dotted :sub verbatim as the host.
		fmt.Sprintf("/exchange/127.0.0.1:%d/q/1", port),
	}
	r := exchangeRouter()
	for _, path := range paths {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s: status %d, want 400", path, w.Code)
		}
	}
	time.Sleep(50 * time.Millisecond)
	if n := atomic.LoadInt32(&connections); n != 0 {
		t.Fatalf("exchange routes connected to an internal address %d time(s)", n)
	}
}

func TestParseParametersResolvesSiteHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := map[string]string{
		"":                            "stackoverflow.com",
		"math":                        "math.stackexchange.com",
		"meta.math.stackexchange.com": "meta.math.stackexchange.com",
		"ru.stackoverflow.com":        "ru.stackoverflow.com",
		"askubuntu.com":               "askubuntu.com",
	}
	for sub, want := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}, {Key: "sub", Value: sub}}
		params, err := parseAndValidateParameters(c)
		if err != nil || params.Sub != want {
			t.Errorf("sub %q: got %q (err %v), want %q", sub, params.Sub, err, want)
		}
	}
}
