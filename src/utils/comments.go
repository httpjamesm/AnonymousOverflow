package utils

import (
	"anonymousoverflow/src/types"

	"github.com/PuerkitoBio/goquery"
)

func FindAndReturnComments(inHtml string, postLayout *goquery.Selection) (comments []types.FilteredComment) {

	commentsComponent := postLayout.Find("div.js-post-comments-component")

	commentsList := commentsComponent.Find("div.comments")
	commentsList2 := commentsList.Find("ul.comments-list")

	allComments := commentsList2.Find("li.comment")

	allComments.Each(func(i int, s *goquery.Selection) {
		commentText := s.Find("div.comment-text")

		commentBody := commentText.Find("div.comment-body")

		commentCopy, err := commentBody.Find("span.comment-copy").Html()
		if err != nil {
			return
		}

		commentAuthorURL := ""

		commentAuthor := commentBody.Find("span.comment-user")
		if commentAuthor.Length() == 0 {
			commentAuthor = commentBody.Find("a.comment-user")
			if commentAuthor.Length() == 0 {
				return
			}

			commentAuthorURL = commentAuthor.AttrOr("href", "")
		}

		commentTimestamp := commentBody.Find("span.relativetime-clean").Text()

		newFilteredComment := types.FilteredComment{
			Text:       commentCopy,
			Timestamp:  commentTimestamp,
			AuthorName: commentAuthor.Text(),
			AuthorURL:  commentAuthorURL,
		}

		comments = append(comments, newFilteredComment)

	})

	return

}
