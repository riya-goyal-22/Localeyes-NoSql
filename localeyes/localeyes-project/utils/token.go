package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
	"time"
)

var ExtractClaimsFunc = ExtractClaims
var GenerateTokenFunc = GenerateToken
var ValidateTokenFunc = ValidateToken
var ValidateAdminTokenFunc = ValidateAdminToken

func GenerateToken(username string, uid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
		"id":  uid,
	})
	signedToken, err := token.SignedString([]byte(os.Getenv("Secret")))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateToken(bearerToken string) bool {
	token := strings.TrimPrefix(bearerToken, "Bearer ")
	token = strings.TrimSpace(token)
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("Secret")), nil
	})
	if err != nil {
		return false
	}
	return parsedToken.Valid
}

func ValidateAdminToken(bearerToken string) bool {
	token := strings.TrimPrefix(bearerToken, "Bearer ")
	token = strings.TrimSpace(token)
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("Secret")), nil
	})
	if err != nil {
		return false
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		if claims["sub"].(string) == os.Getenv("AdminUsername") {
			return true
		}
	}
	return false
}

func ExtractClaims(bearerToken string) (jwt.MapClaims, error) {
	token := strings.TrimPrefix(bearerToken, "Bearer ")
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("Secret")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
