package types

import "html/template"

type FilteredQuestion struct {
	Title         string
	Body          template.HTML
	Timestamp     string
	AuthorName    string
	AuthorURL     string
	ShortenedBody string
	Comments      []FilteredComment
	Tags          []string
}
