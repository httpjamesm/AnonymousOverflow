package utils

import (
	"bytes"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var plainFormattedCodeRegex = regexp.MustCompile(`(?s)<pre tabindex="0" class="chroma"><code>(.+?)</code></pre>`)

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

	formatter := formatters.Get("html")
	if formatter == nil {
		formatter = formatters.Fallback
	}

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

	// parse only the <pre><code>...</code></pre> part
	htmlOut = b.String()
	htmlOut = plainFormattedCodeRegex.FindString(htmlOut)

	htmlOut = StripBlockTags(htmlOut)

	// remove <pre tabindex="0" class="chroma">
	htmlOut = strings.Replace(htmlOut, "<pre tabindex=\"0\" class=\"chroma\">", "", -1)

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
