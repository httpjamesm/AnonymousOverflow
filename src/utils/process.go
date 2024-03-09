package utils

// ProcessHTMLBody runs HTML through the various preparation functions.
func ProcessHTMLBody(bodyHTML string) string {
	highlightedBody := HighlightCodeBlocks(bodyHTML)
	imageProxiedBody := ReplaceImgTags(highlightedBody)
	stackOverflowLinksReplacedBody := ReplaceStackOverflowLinks(imageProxiedBody)
	return stackOverflowLinksReplacedBody
}
