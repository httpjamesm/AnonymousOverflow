package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
