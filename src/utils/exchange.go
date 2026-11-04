package utils

import (
	"regexp"
	"strings"
)

var exchangeRoots = []string{
	"stackoverflow.com",
	"stackexchange.com",
	"superuser.com",
	"serverfault.com",
	"askubuntu.com",
	"mathoverflow.net",
	"stackapps.com",
}

var hostLabelRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func ExchangeHost(sub string) (host string, ok bool) {
	if sub == "" {
		return "stackoverflow.com", true
	}
	host = strings.TrimPrefix(strings.ToLower(sub), "www.")
	for _, label := range strings.Split(host, ".") {
		if !hostLabelRegex.MatchString(label) {
			return "", false
		}
	}
	if !strings.Contains(host, ".") {
		return host + ".stackexchange.com", true
	}
	for _, root := range exchangeRoots {
		if host == root || strings.HasSuffix(host, "."+root) {
			return host, true
		}
	}
	return "", false
}
