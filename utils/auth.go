package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken generates a new JWT token.
func GenerateToken(userID string) (string, error) {
	// Set up the token claims
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
			IssuedAt:  time.Now().Unix(),
			Subject:   "user_token",
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	secretKey := []byte(os.Getenv("JWT_SECRET")) // You need to set JWT_SECRET in your environment variables
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// VerifyToken verifies and decodes a JWT token.
func VerifyToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	// Check for errors in parsing the token
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	// Extract claims from the token
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from token")
	}

	return claims, nil
}
