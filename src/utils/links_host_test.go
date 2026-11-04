package utils

import (
	"regexp"
	"testing"
)

var hrefRegexForTest = regexp.MustCompile(`href="([^"]*)"`)

func TestReplaceStackOverflowLinksKeepsSite(t *testing.T) {
	cases := map[string]string{
		"/questions/1/x": "/questions/1/x",
		"https://stackoverflow.com/questions/1/x":      "/questions/1/x",
		"https://www.stackoverflow.com/questions/1/x":  "/questions/1/x",
		"https://math.stackexchange.com/questions/1/x": "/exchange/math/questions/1/x",
		"https://meta.stackoverflow.com/questions/1/x": "/exchange/meta.stackoverflow.com/questions/1/x",
		"https://ru.stackoverflow.com/questions/1/x":   "/exchange/ru.stackoverflow.com/questions/1/x",
		"https://meta.math.stackexchange.com/q/1":      "/exchange/meta.math.stackexchange.com/q/1",
	}
	for in, want := range cases {
		out := ReplaceStackOverflowLinks(`<a href="` + in + `">l</a>`)
		m := hrefRegexForTest.FindStringSubmatch(out)
		if m == nil || m[1] != want {
			t.Errorf("%s -> %q, want %q", in, out, want)
		}
	}
}
