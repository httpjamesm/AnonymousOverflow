package utils

import (
	"encoding/hex"
	"net/url"
	"strings"
)

// decodeEmail decodes an obfuscated email string.
func DecodeEmail(encoded string) (string, error) {
	// remove the prefix if present
	prefix := "/cdn-cgi/l/email-protection#"
	encoded = strings.TrimPrefix(encoded, prefix)

	// get the key for XOR operation
	key, err := hex.DecodeString(encoded[:2])
	if err != nil {
		return "", err
	}

	// decode the rest of the string using the key
	var decodedChars []byte
	// start a loop from the third character of the encoded string (index 2), as the first two characters are used as the key for XOR operation.
	// the loop increments by 2 in each iteration because the encoded string is in hexadecimal format, and each hexadecimal character represents 4 bits.
	// so, two hexadecimal characters represent one byte (8 bits), which is the size of a character in ASCII.
	for i := 2; i < len(encoded); i += 2 {
		// get the current pair of hexadecimal characters from the encoded string.
		encodedPair := encoded[i : i+2]

		// decode the hexadecimal pair to get the original byte. this is done by converting the hexadecimal pair to a byte using the DecodeString function from the encoding/hex package.
		// if the hexadecimal pair is not a valid hexadecimal number, DecodeString will return an error.
		pairBytes, err := hex.DecodeString(encodedPair)
		if err != nil {
			// if there is an error in decoding, return the error and an empty string.
			return "", err
		}

		// perform an XOR operation on the decoded byte with the key. this is the reverse operation of the original encoding process.
		// the XOR operation is a bitwise operation that returns 0 if the two bits are the same, and 1 otherwise.
		// by performing the XOR operation with the same key used in the encoding process, we can retrieve the original byte.
		decodedChar := pairBytes[0] ^ key[0]

		// append the decoded character to the decodedChars slice. this slice will hold all the decoded characters of the encoded string.
		decodedChars = append(decodedChars, decodedChar)
	}
	// convert the decoded bytes to a string
	decoded := string(decodedChars)

	// unescape any URL-encoded characters
	unescaped, err := url.QueryUnescape(decoded)
	if err != nil {
		return "", err
	}

	return unescaped, nil
}
