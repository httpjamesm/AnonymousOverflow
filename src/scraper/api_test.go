package scraper

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGetQuestionContentNotFound(t *testing.T) {
	client := resty.New().SetTransport(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"items":[]}`)),
			Request:    r,
		}, nil
	}))

	_, err := getQuestionContent(client, ViewQuestionInputs{QuestionID: "2120178", Sub: "stackoverflow"})
	if !errors.Is(err, ErrQuestionNotFound) {
		t.Fatalf("expected ErrQuestionNotFound, got %v", err)
	}
}
