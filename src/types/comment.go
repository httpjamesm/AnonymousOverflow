package types

import "html/template"

type FilteredComment struct {
	Text       template.HTML
	Timestamp  string
	AuthorName string
	AuthorURL  string
	Upvotes    string
}
