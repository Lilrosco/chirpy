package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

const tokenAccessType = "chirpy-access"

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    tokenAccessType,
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	
	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)

	if err != nil {
		log.Printf("Error parsing token: %v\n", err)
		return uuid.Nil, err
	}

	userID, err := parsedToken.Claims.GetSubject()

	if err != nil {
		log.Printf("Error getting subject from token: %v\n", err)
		return uuid.Nil, err
	}

	issuer, err := parsedToken.Claims.GetIssuer()

	if err != nil {
		log.Printf("Error getting issuer from token: %v\n", err)
		return uuid.Nil, err
	}

	if issuer != tokenAccessType {
		log.Println("Invalid issuer detected!")
		return uuid.Nil, fmt.Errorf("Invalid issuer!")
	}

	parsedUUID, err := uuid.Parse(userID)

	if err != nil {
		log.Printf("Error parsing UUID: %v", err)
		return uuid.Nil, err
	}
		
	return parsedUUID, nil
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		log.Printf("Could not hash password: %v\n", err)
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		log.Printf("Error in call to compare password and hash: %v\n", err)
		return false, err
	}

	return match, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	if authHeader == "" {
		return "", fmt.Errorf("Authorization header missing in request")
	}

	splitStr := strings.Split(authHeader, " ")

	if len(splitStr) < 2 || splitStr[0] != "Bearer" {
		return "", fmt.Errorf("Malformed Authorization Header")
	}

	return splitStr[1], nil
}
