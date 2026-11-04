package utils

import "testing"

func TestExchangeHost(t *testing.T) {
	valid := map[string]string{
		"":                            "stackoverflow.com",
		"math":                        "math.stackexchange.com",
		"Math":                        "math.stackexchange.com",
		"askubuntu.com":               "askubuntu.com",
		"mathoverflow.net":            "mathoverflow.net",
		"meta.math.stackexchange.com": "meta.math.stackexchange.com",
		"ru.stackoverflow.com":        "ru.stackoverflow.com",
		"meta.stackoverflow.com":      "meta.stackoverflow.com",
		"www.stackoverflow.com":       "stackoverflow.com",
	}
	for sub, want := range valid {
		if got, ok := ExchangeHost(sub); !ok || got != want {
			t.Errorf("ExchangeHost(%q) = %q, %v; want %q", sub, got, ok, want)
		}
	}

	invalid := []string{
		"[::ffff:7f00:1]:8443#",
		"127.0.0.1",
		"169.254.169.254",
		"localhost:8080",
		"internal.example",
		"stackoverflow.com:8443",
		"stackoverflow.com.attacker.example",
		"attacker.example#.stackoverflow.com",
		"user@stackoverflow.com",
		"notstackoverflow.com",
		"math?",
		"math#",
		"-math",
		"math..stackexchange.com",
		"stackoverflow.com.",
	}
	for _, sub := range invalid {
		if got, ok := ExchangeHost(sub); ok {
			t.Errorf("ExchangeHost(%q) accepted as %q", sub, got)
		}
	}
}
