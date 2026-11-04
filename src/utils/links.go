package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// stackOverflowLinkQualifierRegex matches all anchor elements that meet the following conditions:
// * must be an anchor element
// * the anchor element must have a pathname beginning with /q or /questions
// * if there is a host, it must be stackoverflow.com or a subdomain
var stackOverflowLinkQualifierRegex = regexp.MustCompile(`<a\s[^>]*href="(?:https?://(?:www\.)?(?:\w+\.)*(?:stackoverflow|stackexchange)\.com)?/(?:q|questions)/[^"]*"[^>]*>.*?</a>`)

func ReplaceStackOverflowLinks(html string) string {
	return stackOverflowLinkQualifierRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the href attribute value from the anchor tag
		hrefRegex := regexp.MustCompile(`href="([^"]*)"`)
		hrefMatch := hrefRegex.FindStringSubmatch(match)
		if len(hrefMatch) < 2 {
			return match
		}
		href := hrefMatch[1]

		// Parse the URL
		url, err := url.Parse(href)
		if err != nil {
			return match
		}

		newUrl := url.String()

		host := strings.TrimPrefix(strings.ToLower(url.Host), "www.")
		if strings.HasSuffix(host, ".stackexchange.com") {
			if sub := strings.TrimSuffix(host, ".stackexchange.com"); !strings.Contains(sub, ".") {
				host = sub
			}
		}
		if host != "" && host != "stackoverflow.com" {
			url.Path = "/exchange/" + host + url.Path
		}

		newUrl = url.Path + url.RawQuery + url.Fragment

		// Replace the href attribute value in the anchor tag
		return strings.Replace(match, hrefMatch[1], newUrl, 1)
	})
}