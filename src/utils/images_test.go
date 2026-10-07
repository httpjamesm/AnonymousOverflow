package utils

import (
	"anonymousoverflow/src/types"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestImageProxyAuth(t *testing.T) {
	const secret = "image-proxy-test-secret"
	const imageURL = "https://example.com/image.png"
	t.Setenv("JWT_SIGNING_SECRET", secret)

	authorization, err := generateImageProxyAuth(imageURL)
	if err != nil {
		t.Fatal(err)
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}
	claims := &types.ImageProxyClaims{}
	token, err := jwt.ParseWithClaims(authorization, claims, keyFunc)
	if err != nil || !token.Valid {
		t.Fatalf("fresh token rejected: %v", err)
	}
	if claims.Action != "imageProxy" || claims.ImageURL != imageURL {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("missing issue time or expiry")
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt.Time) != time.Minute {
		t.Fatal("token lifetime is not one minute")
	}

	originalTimeFunc := jwt.TimeFunc
	t.Cleanup(func() { jwt.TimeFunc = originalTimeFunc })
	jwt.TimeFunc = func() time.Time { return claims.ExpiresAt.Add(time.Second) }

	token, err = jwt.ParseWithClaims(authorization, &types.ImageProxyClaims{}, keyFunc)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected expiry error from ParseWithClaims, got %v", err)
	}
	if token != nil && token.Valid {
		t.Fatal("expired token is valid")
	}
}
