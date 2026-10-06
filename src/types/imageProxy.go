package types

import "github.com/golang-jwt/jwt/v5"

type ImageProxyClaims struct {
	Action string `json:"action"`

	ImageURL string `json:"image_url"`

	Iss int64 `json:"iss"`
	Exp int64 `json:"exp"`

	jwt.RegisteredClaims
}
