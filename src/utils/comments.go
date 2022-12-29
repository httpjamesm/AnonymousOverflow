package utils

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func FindAndReturnComments(inHtml string, postLayout *goquery.Selection) (outHtml string) {
	outHtml = inHtml

	comments := []string{}

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

		comment := fmt.Sprintf(`<div class="comment-parent"><div class="comment"><div class="comment-body">%s</div><div class="comment-author">Commented %s by <a href="https://stackoverflow.com%s" target="_blank" rel="noopener noreferrer">%s</a>.</div></div></div>`, commentCopy, commentTimestamp, commentAuthorURL, commentAuthor.Text())

		comments = append(comments, comment)

	})

	if len(comments) > 0 {
		outHtml = inHtml + fmt.Sprintf(`<details class="comments"><summary>Show <b>%d comments</b></summary><div class="comments-parent">%s</div></details>`, len(comments), strings.Join(comments, ""))
	}

	return

}
