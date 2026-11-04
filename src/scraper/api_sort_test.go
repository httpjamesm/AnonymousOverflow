package scraper

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
)

type apiSortStub func(*http.Request) (*http.Response, error)

func (f apiSortStub) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// fakeSortAPI answers like api.stackexchange.com does: an unsupported sort
// value is a 400 bad_parameter.
func fakeSortAPI() *resty.Client {
	return resty.New().SetTransport(apiSortStub(func(r *http.Request) (*http.Response, error) {
		allowed := map[string]bool{"creation": true, "votes": true}
		if strings.HasSuffix(r.URL.Path, "/answers") {
			allowed["activity"] = true
		}
		q := r.URL.Query()
		status, body := http.StatusOK, `{"items":[]}`
		order := q.Get("order")
		if !allowed[q.Get("sort")] || (order != "" && order != "asc" && order != "desc") {
			status, body = http.StatusBadRequest, `{"error_id":400,"error_message":"sort","error_name":"bad_parameter"}`
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	}))
}

func TestApiScraperSortOptions(t *testing.T) {
	client := fakeSortAPI()
	// Every option in the question page's sort menu, plus junk.
	for _, sortValue := range []string{"votes", "trending", "newest", "oldest", "bogus&site=x"} {
		params := ViewQuestionInputs{QuestionID: "11227809", Sub: "stackoverflow", SortValue: sortValue}
		if _, err := getAnswers(client, params); err != nil {
			t.Errorf("sort_by=%s: answers: %v", sortValue, err)
		}
		if _, err := getComments(client, params, []string{"11227809"}, "questions"); err != nil {
			t.Errorf("sort_by=%s: question comments: %v", sortValue, err)
		}
		if _, err := getComments(client, params, []string{"11227831"}, "answers"); err != nil {
			t.Errorf("sort_by=%s: answer comments: %v", sortValue, err)
		}
	}
}
