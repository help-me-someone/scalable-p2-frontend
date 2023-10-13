package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
)

// For simplicity, we will just declare a secret here.
// NOTE: For production, please remove this!
var jwtKey = []byte("my_secret_key")

func ValidateJWTTOken(r *http.Request) (*Claims, error) {
	// Obtain the session token.
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Println("The cookie does not exist")
		return nil, err
	}

	// Get the JWT string from the cookie
	tokenString := cookie.Value

	log.Printf("Cookie: %s\n", tokenString)

	claims := &Claims{}

	// Parse the JWT and store it inside claims.
	// Will fail in two cases:
	// - Expired
	// - Signature does not match.
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// We use the same key that was responsible for creating the token.
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Invalid token.")
	}

	return claims, nil
}
