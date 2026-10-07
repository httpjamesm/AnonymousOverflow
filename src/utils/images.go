package utils

import (
	"anonymousoverflow/src/types"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var imgTagRegex = regexp.MustCompile(`<img[^>]*\s+src\s*=\s*"(.*?)"[^>]*>`)

func ReplaceImgTags(inHtml string) string {
	imgTags := imgTagRegex.FindAllString(inHtml, -1)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, imgTag := range imgTags {
		wg.Add(1)
		go func(imgTag string) {
			defer wg.Done()
			srcRegex := regexp.MustCompile(`src\s*=\s*"(.*?)"`)
			src := srcRegex.FindStringSubmatch(imgTag)[1]

			authToken, _ := generateImageProxyAuth(src)

			mu.Lock()
			defer mu.Unlock()
			inHtml = strings.Replace(inHtml, imgTag, fmt.Sprintf(`<img src="%s/proxy?auth=%s">`, os.Getenv("APP_URL"), authToken), 1)
		}(imgTag)
	}

	wg.Wait()

	return inHtml
}

func generateImageProxyAuth(url string) (string, error) {
	// generate a jwt with types.ImageProxyClaims
	now := time.Now()
	claims := types.ImageProxyClaims{
		Action:   "imageProxy",
		ImageURL: url,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// sign the token
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_SECRET")))
	if err != nil {
		return "", err
	}

	return ss, nil
}
