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

		c.HTML(200, "question.html", gin.H{
			"title": questionText,
			"body":  template.HTML(questionBodyParentHTML),
		})

	})
	r.Run("localhost:8080")
}
