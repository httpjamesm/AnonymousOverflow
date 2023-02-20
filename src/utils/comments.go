package utils

import (
	"anonymousoverflow/src/types"
	"fmt"
	"html/template"

	"github.com/PuerkitoBio/goquery"
)

func FindAndReturnComments(inHtml, domain string, postLayout *goquery.Selection) (comments []types.FilteredComment) {

	commentsComponent := postLayout.Find("div.js-post-comments-component")

	commentsList := commentsComponent.Find("div.comments")
	commentsList2 := commentsList.Find("ul.comments-list")

	allComments := commentsList2.Find("li.comment")

	allComments.Each(func(i int, s *goquery.Selection) {
		commentScoreParent := s.Find("div.comment-score")

		// find the first span within commentScoreParent
		commentScore := commentScoreParent.Find("span").First().Text()
		if commentScore == "" {
			commentScore = "0"
		}

		commentText := s.Find("div.comment-text")

		commentBody := commentText.Find("div.comment-body")

		commentCopy, err := commentBody.Find("span.comment-copy").Html()
		if err != nil {
			return
		}

		commentAuthorURL := fmt.Sprintf("https://%s", domain)

		commentAuthor := commentBody.Find("span.comment-user")
		if commentAuthor.Length() == 0 {
			commentAuthor = commentBody.Find("a.comment-user")
			if commentAuthor.Length() == 0 {
				return
			}

			commentAuthorURL += commentAuthor.AttrOr("href", "")
		}

		commentTimestamp := commentBody.Find("span.relativetime-clean").Text()

		newFilteredComment := types.FilteredComment{
			Text:       template.HTML(commentCopy),
			Timestamp:  commentTimestamp,
			AuthorName: commentAuthor.Text(),
			AuthorURL:  commentAuthorURL,
			Upvotes:    commentScore,
		}

		comments = append(comments, newFilteredComment)

	})

	return

}
