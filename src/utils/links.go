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

		// Check if the host is a subdomain
		parts := strings.Split(url.Host, ".")
		if len(parts) > 2 {
			// Prepend the subdomain to the path
			url.Path = "/exchange/" + parts[0] + url.Path
		}

		newUrl = url.Path + url.RawQuery + url.Fragment

		// Replace the href attribute value in the anchor tag
		return strings.Replace(match, hrefMatch[1], newUrl, 1)
	})
}

var relativeAnchorURLRegex = regexp.MustCompile(`href="(/[^"]+)"`)

func ConvertRelativeAnchorURLsToAbsolute(html, prefix string) string {
	if prefix == "" {
		return html
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return relativeAnchorURLRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the URL from the match
		url := strings.TrimPrefix(match, `href="`)
		url = strings.TrimSuffix(url, `"`)

		// If the URL already has the desired prefix, return the match as is
		if strings.HasPrefix(url, prefix) {
			return match
		}

		// Otherwise, prepend the prefix
		return strings.Replace(match, `href="/`, `href="`+prefix, 1)
	})
}
