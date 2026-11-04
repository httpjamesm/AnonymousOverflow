package utils

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func rewrittenHref(t *testing.T, href string) (string, *goquery.Document) {
	t.Helper()
	out := ReplaceStackOverflowLinks(`<p><a href="` + href + `" rel="nofollow noreferrer">link</a></p>`)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	return doc.Find("a").AttrOr("href", ""), doc
}

func TestReplaceStackOverflowLinksKeepsHrefEscaped(t *testing.T) {
	// A link in a post whose path carries percent-encoded markup.
	href, doc := rewrittenHref(t, `https://stackoverflow.com/questions/1/x%22%3E%3Cimg%20src=%22https://tracker.example/p.gif%22%3E`)
	if n := doc.Find("img").Length(); n != 0 {
		t.Fatalf("rewritten link injected %d <img> element(s)", n)
	}
	if want := `/questions/1/x%22%3E%3Cimg%20src=%22https://tracker.example/p.gif%22%3E`; href != want {
		t.Fatalf("href = %q, want %q", href, want)
	}
	if doc.Find("a").AttrOr("rel", "") != "nofollow noreferrer" {
		t.Fatal("attributes after href were lost")
	}

	href, _ = rewrittenHref(t, `https://stackoverflow.com/questions/1/x#a%22b`)
	if href != `/questions/1/x#a%22b` {
		t.Fatalf("fragment href = %q", href)
	}
}

func TestReplaceStackOverflowLinksKeepsQueryAndFragment(t *testing.T) {
	cases := map[string]string{
		"https://stackoverflow.com/questions/123/slug/456#456":       "/questions/123/slug/456#456",
		"https://stackoverflow.com/questions/123?rq=1":               "/questions/123?rq=1",
		"https://stackoverflow.com/questions/123#comment9_123":       "/questions/123#comment9_123",
		"https://stackoverflow.com/questions/1?a=1&amp;b=2":          "/questions/1?a=1&b=2",
		"https://math.stackexchange.com/questions/1/x?a=1#c":         "/exchange/math/questions/1/x?a=1#c",
		"/questions/55083952/is-it-possible-to-populate-a-large-set": "/questions/55083952/is-it-possible-to-populate-a-large-set",
	}
	for in, want := range cases {
		if got, _ := rewrittenHref(t, in); got != want {
			t.Errorf("%s -> %q, want %q", in, got, want)
		}
	}
}
