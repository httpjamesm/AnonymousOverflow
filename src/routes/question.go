package routes

import (
	"anonymousoverflow/src/utils"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var codeBlockRegex = regexp.MustCompile(`(?s)<pre><code>(.+?)<\/code><\/pre>`)
var questionCodeBlockRegex = regexp.MustCompile(`(?s)<pre class=".+"><code( class=".+")?>(.+?)</code></pre>`)

func ViewQuestion(c *gin.Context) {
	client := resty.New()

	questionId := c.Param("id")
	questionTitle := c.Param("title")

	soLink := fmt.Sprintf("https://stackoverflow.com/questions/%s/%s", questionId, questionTitle)

	resp, err := client.R().Get(soLink)
	if err != nil {
		panic(err)
	}

	respBody := resp.String()

	respBodyReader := strings.NewReader(respBody)

	doc, err := goquery.NewDocumentFromReader(respBodyReader)
	if err != nil {
		panic(err)
	}

	questionTextParent := doc.Find("h1.fs-headline1")

	questionText := questionTextParent.Children().First().Text()

	questionBodyParent := doc.Find("div.s-prose")

	questionBodyParentHTML, err := questionBodyParent.Html()
	if err != nil {
		panic(err)
	}

	// parse any code blocks and highlight them
	answerCodeBlocks := questionCodeBlockRegex.FindAllString(questionBodyParentHTML, -1)
	for _, codeBlock := range answerCodeBlocks {
		codeBlock = utils.StripBlockTags(codeBlock)

		// syntax highlight
		highlightedCodeBlock := utils.HighlightSyntaxViaContent(codeBlock)

		// replace the code block with the highlighted code block
		questionBodyParentHTML = strings.Replace(questionBodyParentHTML, codeBlock, highlightedCodeBlock, 1)
	}

	questionCard := doc.Find("div.postcell")

	questionMetadata := questionCard.Find("div.user-info")
	questionTimestamp := ""
	questionMetadata.Find("span.relativetime").Each(func(i int, s *goquery.Selection) {
		// get the second
		if i == 0 {
			if s.Text() != "" {
				// if it's not been edited, it means it's the first
				questionTimestamp = s.Text()
				return
			}
		}

		// otherwise it's the second element
		if i == 1 {
			questionTimestamp = s.Text()
			return
		}
	})

	userDetails := questionMetadata.Find("div.user-details")

	questionAuthor := ""
	questionAuthorURL := ""

	userDetails.Find("a").Each(func(i int, s *goquery.Selection) {
		// get the second
		if i == 0 {
			if s.Text() != "" {
				// if it's not been edited, it means it's the first
				questionAuthor = s.Text()
				questionAuthorURL, _ = s.Attr("href")
				return
			}
		}

		// otherwise it's the second element
		if i == 1 {
			questionAuthor = s.Text()
			questionAuthorURL, _ = s.Attr("href")
			return
		}
	})

	answers := []template.HTML{}

	doc.Find("div.answer").Each(func(i int, s *goquery.Selection) {
		postLayout := s.Find("div.post-layout")
		voteCell := postLayout.Find("div.votecell")
		answerCell := postLayout.Find("div.answercell")
		answerBody := answerCell.Find("div.s-prose")
		answerBodyHTML, _ := answerBody.Html()

		voteCount := voteCell.Find("div.js-vote-count").Text()

		if s.HasClass("accepted-answer") {
			// add <div class="answer-meta accepted">Accepted Answer</div> to the top of the answer
			answerBodyHTML = fmt.Sprintf(`<div class="answer-meta accepted">Accepted Answer - %s Upvotes</div>`, voteCount) + answerBodyHTML
		} else {
			// add <div class="answer-meta">%s Upvotes</div> to the top of the answer
			answerBodyHTML = fmt.Sprintf(`<div class="answer-meta">%s Upvotes</div>`, voteCount) + answerBodyHTML
		}

		answerFooter := s.Find("div.mt24")

		answerAuthorURL := ""
		answerAuthorName := ""
		answerTimestamp := ""

		answerFooter.Find("div.post-signature").Each(func(i int, s *goquery.Selection) {
			answerAuthorDetails := s.Find("div.user-details")

			if answerAuthorDetails.Length() == 0 {
				return
			}

			if answerAuthorDetails.Length() > 1 {
				if i == 0 {
					return
				}
			}

			answerAuthor := answerAuthorDetails.Find("a").First()

			answerAuthorURL = answerAuthor.AttrOr("href", "")
			answerAuthorName = answerAuthor.Text()
			answerTimestamp = s.Find("span.relativetime").Text()
		})

		// append <div class="answer-author">Answered %s by %s</div> to the bottom of the answer
		answerBodyHTML += fmt.Sprintf(`<div class="answer-author-parent"><div class="answer-author">Answered %s by <a href="https://stackoverflow.com/%s" target="_blank" rel="noopener noreferrer">%s</a></div></div>`, answerTimestamp, answerAuthorURL, answerAuthorName)

		// parse any code blocks and highlight them
		answerCodeBlocks := codeBlockRegex.FindAllString(answerBodyHTML, -1)
		for _, codeBlock := range answerCodeBlocks {
			codeBlock = utils.StripBlockTags(codeBlock)

			// syntax highlight
			highlightedCodeBlock := utils.HighlightSyntaxViaContent(codeBlock)

			// replace the code block with the highlighted code block
			answerBodyHTML = strings.Replace(answerBodyHTML, codeBlock, highlightedCodeBlock, 1)
		}

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
			answerBodyHTML += fmt.Sprintf(`<details class="comments"><summary>Show <b>%d comments</b></summary><div class="comments-parent">%s</div></details>`, len(comments), strings.Join(comments, ""))
		}

		answers = append(answers, template.HTML(answerBodyHTML))
	})

	imagePolicy := "'self' https:"

	if c.MustGet("disable_images").(bool) {
		imagePolicy = "'self'"
	}

	c.HTML(200, "question.html", gin.H{
		"title":         questionText,
		"body":          template.HTML(questionBodyParentHTML),
		"timestamp":     questionTimestamp,
		"author":        questionAuthor,
		"authorURL":     questionAuthorURL,
		"answers":       answers,
		"imagePolicy":   imagePolicy,
		"shortenedBody": questionBodyParent.Text()[0:50],
		"theme":         c.MustGet("theme").(string),
		"currentUrl":    fmt.Sprintf("%s/questions/%s/%s", os.Getenv("APP_URL"), questionId, questionTitle),
	})

}
