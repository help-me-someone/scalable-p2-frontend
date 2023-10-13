package main

import (
	"github.com/golang-jwt/jwt/v5"
)

type Response struct {
	Username      string `json:"username"`
	Authenticated bool   `json:"authenticated"`
}

// Create a struct which will be encoded to a JWT.
// Embedded registered claims. This provides us with fields like "Expire time", etc...
type Claims struct {
	Username string `json:"username" form:"username"`
	jwt.RegisteredClaims
}
