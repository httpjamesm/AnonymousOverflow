package utils

import (
	"github.com/PuerkitoBio/goquery"
)

func GetPostTags(postLayout *goquery.Selection) []string {
	var tags []string
	postLayout.Find("a.post-tag").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})
	return tags
}
