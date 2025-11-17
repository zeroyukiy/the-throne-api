package internal

import "github.com/golang-jwt/jwt/v4"

type JwtCustomClaims struct {
	Name   string `json:"name"`
	Admin  bool   `json:"admin"`
	Avatar string `json:"avatar"`
	jwt.RegisteredClaims
}
