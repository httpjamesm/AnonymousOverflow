package utils

import (
	"bytes"
	"html"
	"io"
	"regexp"
	"strings"

	html_formatter "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// highlightSyntaxViaContent uses Chroma to lex code content and apply the appropriate tokenizer engine.
// If it can't find one, it defaults to JavaScript syntax highlighting.
func highlightSyntaxViaContent(content string, lang string) (htmlOut string) {
	content = html.UnescapeString(content)

	fallbackOut := html.EscapeString(content)

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Get(".js")
	}

	style := styles.Get("xcode")
	if style == nil {
		style = styles.Fallback
	}

	formatter := html_formatter.New(html_formatter.PreventSurroundingPre(true), html_formatter.WithClasses(true))

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		htmlOut = fallbackOut
		return
	}

	b := bytes.NewBufferString("")
	w := io.Writer(b)

	err = formatter.Format(w, style, iterator)
	if err != nil {
		htmlOut = fallbackOut
		return
	}

	htmlOut = b.String()

	return
}

var preClassRegex = regexp.MustCompile(`(?s)<pre class=".+">`)

// stripBlockTags takes an extracted code block from HTML and strips it of its pre and code tags.
// What's returned is just the code.
func stripBlockTags(content string) (result string) {
	// strip all "<code>" tags
	content = strings.Replace(content, "<code>", "", -1)
	content = strings.Replace(content, "</code>", "", -1)
	// and the <pre>
	content = strings.Replace(content, "<pre>", "", -1)
	content = strings.Replace(content, "</pre>", "", -1)

	content = preClassRegex.ReplaceAllString(content, "")

	result = content

	return
}

var codeBlockRegex = regexp.MustCompile(`(?s)<pre.*?lang-(.*?)[\s"'].*?><code>(.*?)<\/code><\/pre>`)

// HighlightCodeBlocks uses both highlightSyntaxViaContent stripCodeBlocks and returns the newly highlighted code HTML.
func HighlightCodeBlocks(html string) string {
	// Replace each code block with the highlighted version
	highlightedHTML := codeBlockRegex.ReplaceAllStringFunc(html, func(codeBlock string) string {
		// Extract the code content from the code block
		matches := codeBlockRegex.FindStringSubmatch(codeBlock)
		lang, codeContent := matches[1], matches[2]

		codeContent = stripBlockTags(codeContent)

		// Highlight the code content
		highlightedCode := highlightSyntaxViaContent(codeContent, lang)

		// Replace the original code block with the highlighted version
		highlightedCodeBlock := "<pre>" + highlightedCode + "</pre>"

		return highlightedCodeBlock
	})

	return highlightedHTML
}
