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

func HighlightSyntaxViaContent(content string) (htmlOut string) {
	content = html.UnescapeString(content)

	fallbackOut := html.EscapeString(content)

	// identify the language
	lexer := lexers.Analyse(content)
	if lexer == nil {
		// unable to identify, so just return the wrapped content
		htmlOut = fallbackOut
		return
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

func StripBlockTags(content string) (result string) {
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
