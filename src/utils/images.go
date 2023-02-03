package utils

import (
	"anonymousoverflow/src/types"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var imgTagRegex = regexp.MustCompile(`<img[^>]*\s+src\s*=\s*"(.*?)"[^>]*>`)

func ReplaceImgTags(inHtml string) string {
	// find all img tags
	imgTags := imgTagRegex.FindAllString(inHtml, -1)

	for _, imgTag := range imgTags {
		// parse the src="" attribute
		srcRegex := regexp.MustCompile(`src\s*=\s*"(.*?)"`)
		src := srcRegex.FindStringSubmatch(imgTag)[1]

		authToken, _ := generateImageProxyAuth(src)

		// replace the img tag with a proxied url
		inHtml = strings.Replace(inHtml, imgTag, fmt.Sprintf(`<img src="%s/proxy?auth=%s">`, os.Getenv("APP_URL"), authToken), 1)
	}

	return inHtml
}

func generateImageProxyAuth(url string) (string, error) {
	// generate a jwt with types.ImageProxyClaims
	claims := types.ImageProxyClaims{
		Action:   "imageProxy",
		ImageURL: url,
		Iss:      time.Now().Unix(),
		Exp:      time.Now().Add(time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// sign the token
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_SECRET")))
	if err != nil {
		return "", err
	}

	return ss, nil
}
