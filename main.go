package main

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"

	"github.com/go-resty/resty/v2"
)

func main() {

	client := resty.New()

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	r.GET("/questions/:id/:title", func(c *gin.Context) {
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

		c.HTML(200, "question.html", gin.H{
			"title":     questionText,
			"body":      template.HTML(questionBodyParentHTML),
			"timestamp": questionTimestamp,
			"author":    questionAuthor,
			"authorURL": questionAuthorURL,
		})

	})
	r.Run("localhost:8080")
}
