package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateUrl(t *testing.T) {
	assert := assert.New(t)

	// Test with a Valid StackOverflow URL
	assert.Equal("/questions/example-question", translateUrl("https://stackoverflow.com/questions/example-question"), "StackOverflow URL should not be modified")

	// Test with Complex Subdomain
	assert.Equal("/exchange/meta.math.stackexchange.com/q/example-question", translateUrl("https://meta.math.stackexchange.com/q/example-question"), "Complex StackExchange subdomain should be used as full exchange")

	// Test with Non-StackExchange Domain
	assert.Equal("/exchange/example.com/questions/example-question", translateUrl("https://example.com/questions/example-question"), "Non-StackExchange domain should be detected as exchange")

	// Test with Invalid URL
	assert.Equal("", translateUrl("This is not a URL"), "Invalid URL should return an empty string")

	// Test with Empty String
	assert.Equal("", translateUrl(""), "Empty string should return an empty string")

	// Test with Missing Path
	assert.Equal("", translateUrl("https://stackoverflow.com"), "URL without path should return an empty string")

	// Test with Valid URL but Root Domain for StackExchange
	assert.Equal("", translateUrl("https://stackexchange.com"), "Root StackExchange domain without subdomain should return an empty string")
}
