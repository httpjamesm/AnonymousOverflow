package utils

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleInput = `<div class="d-flex fd-column fw-nowrap">
	<div class="d-flex fw-nowrap">
		<div class="flex--item wmn0 fl1 lh-lg">
			<div class="flex--item fl1 lh-lg">
					<div>
						<b>This question already has answers here</b>:

					</div>
			</div>
		</div>
	</div>
			<div class="flex--item mb0 mt4">
				<a href="/questions/55083952/is-it-possible-to-populate-a-large-set-at-compile-time" dir="ltr">Is it possible to populate a large set at compile time?</a>
					<span class="question-originals-answer-count">
						(3 answers)
					</span>
			</div>
			<div class="flex--item mb0 mt4">
				<a href="https://stackoverflow.com/questions/27221504/how-can-you-make-a-safe-static-singleton-in-rust" dir="ltr">How can you make a safe static singleton in Rust?</a>
					<span class="question-originals-answer-count">
						(5 answers)
					</span>
			</div>
			<div class="flex--item mb0 mt4">
				<a href="https://security.stackexchange.com/questions/25371/brute-force-an-ssh-login-that-has-only-a-4-letter-password" dir="ltr">Brute-force an SSH-login that has only a 4-letter password</a>
					<span class="question-originals-answer-count">
						(9 answers)
					</span>
			</div>
		<div class="flex--item mb0 mt8">Closed <span title="2020-01-29 14:28:42Z" class="relativetime">4 years ago</span>.</div>
</div>`

func TestReplaceStackOverflowLinks(t *testing.T) {
	replacedLinks := ReplaceStackOverflowLinks(sampleInput)

	fmt.Println(replacedLinks)

	assert.False(t, strings.Contains(replacedLinks, "stackoverflow.com"))
	assert.False(t, strings.Contains(replacedLinks, "stackexchange.com"))
}

var sampleRelativeAnchorURLsInput = `<aside class="s-notice s-notice__info post-notice js-post-notice mb16" role="status">
        <div class="d-flex fd-column fw-nowrap">
            <div class="d-flex fw-nowrap">
                <div class="flex--item wmn0 fl1 lh-lg">
                    <div class="flex--item fl1 lh-lg">
                            <div>
                                <b>This question already has an answer here</b>:
                                
                            </div>
                    </div>
                </div>
            </div>
                    <div class="flex--item mb0 mt4">
                        <a href="/questions/5771/is-it-possible-to-restrict-gnu-gplv3-to-non-commercial-use-only" dir="ltr">Is it possible to restrict GNU GPLv3 to non-commercial use only?</a>
                            <span class="question-originals-answer-count">
                                (1 answer)
                            </span>
                    </div>
                <div class="flex--item mb0 mt8">Closed <span title="2018-09-30 22:47:30Z" class="relativetime">5 years ago</span>.</div>
        </div>
</aside>
<div class="answer-author-parent">
            <div class="answer-author">
                Answered Sep 27, 2018 at 21:21 by
                <a href="https://notopensource.stackexchange.com/users/1212/amon" target="_blank" rel="noopener noreferrer">amon</a>
            </div>
        </div>

		<a href="/exchange/opensource/9999/1111">This shouldn't be re-prefixed</a>
`

func TestConvertRelativeAnchorURLsToAbsolute(t *testing.T) {
	prefix := "/exchange/opensource"
	fixedHTML := ConvertRelativeAnchorURLsToAbsolute(sampleRelativeAnchorURLsInput, prefix)

	log.Println(fixedHTML)

	assert.True(t, strings.Contains(fixedHTML, prefix))
	assert.True(t, strings.Contains(fixedHTML, "https://notopensource.stackexchange.com"))
}
