package scraper

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
)

const questionPageFixture = `<html><body>
<h1 class="fs-headline1"><a href="/questions/1/t">Title</a></h1>
<div class="post-layout">
  <div class="postcell">
    <div class="s-prose"><p>Body</p></div>
    <div class="post-signature owner"><div class="user-info"><span class="relativetime">1h</span>
      <div class="user-details"><a href="/users/1/asker">asker</a></div></div></div>
  </div>
  <div class="js-post-comments-component"><div class="comments"><ul class="comments-list">
    <li class="comment"><div class="comment-score"><span>3</span></div><div class="comment-text"><div class="comment-body">
      <span class="comment-copy">nice</span> <a class="comment-user" href="/users/2/commenter">commenter</a>
      <span class="relativetime-clean">2h</span></div></div></li>
  </ul></div></div>
</div>
<div class="answer" data-answerid="10"><div class="post-layout">
  <div class="votecell"><div class="js-vote-count">5</div></div>
  <div class="answercell"><div class="s-prose"><p>Answer</p></div>
    <div class="post-signature"><div class="user-details"><a href="/users/3/answerer">answerer</a></div>
      <span class="relativetime">3h</span></div></div>
</div></div>
</body></html>`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHtmlScraperUsesSiteHost(t *testing.T) {
	var requested string
	orig := newClient
	newClient = func() *resty.Client {
		return resty.New().SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			requested = r.URL.Host
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"text/html"}},
				Body:       io.NopCloser(strings.NewReader(questionPageFixture)),
				Request:    r,
			}, nil
		}))
	}
	defer func() { newClient = orig }()

	for _, host := range []string{"stackoverflow.com", "math.stackexchange.com", "meta.math.stackexchange.com", "ru.stackoverflow.com"} {
		question, answers, err := HtmlScraper{}.GetQuestion(ViewQuestionInputs{QuestionID: "1", SortValue: "votes", Sub: host})
		if err != nil {
			t.Fatalf("%s: %v", host, err)
		}
		if requested != host {
			t.Errorf("%s: fetched from %q", host, requested)
		}
		want := map[string]string{
			"question author": "https://" + host + "/users/1/asker",
			"comment author":  "https://" + host + "/users/2/commenter",
			"answer author":   "https://" + host + "/users/3/answerer",
		}
		got := map[string]string{"question author": question.AuthorURL}
		if len(question.Comments) == 1 {
			got["comment author"] = question.Comments[0].AuthorURL
		}
		if len(answers) == 1 {
			got["answer author"] = answers[0].AuthorURL
		}
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s: %s URL = %q, want %q", host, k, got[k], v)
			}
		}
	}
}
