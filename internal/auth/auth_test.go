package auth

import (
	"fmt"
	"net/http"
    "testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHash(t *testing.T) {
    password := "Gladys"
    hashed, err := HashPassword(password)
	
	if err != nil {
		t.Errorf(`Error was generated from HashPassword(): %v`, err)
	}

    got, err := CheckPasswordHash(password, hashed)
	want := true
	
    if !got {
		t.Errorf("CheckPasswordHash() = %t; want %t", got, want)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestCheckJWT(t *testing.T) {
	user := uuid.New()
	secret := "Kamehameha!"
	validToken, _ := MakeJWT(user, secret, 5*time.Minute)
	
	tests := []struct {
		name          string
		tokenString   string
		tokenSecret   string
		wantedUserID  uuid.UUID
		wantErr       bool
	}{
		{
			name:          "Valid JWT",
			tokenString:   validToken,
			tokenSecret:   secret,
			wantedUserID:  user,
			wantErr:       false,
		},
		{
			name:          "Invalid JWT",
			tokenString:   "bad_jwt",
			tokenSecret:   secret,
			wantedUserID:  uuid.Nil,
			wantErr:       true,
		},
		{
			name:          "Bad Secret",
			tokenString:   validToken,
			tokenSecret:   "something_bad",
			wantedUserID:  uuid.Nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if userID != tt.wantedUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", userID, tt.wantedUserID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenStr := "some-token-string"
	good_headers := http.Header{}
	good_headers.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
	bad_headers := http.Header{}
	bad_headers.Set("Content-Type", "application/json")
	
	tests := []struct {
		name          string
		headers       http.Header
		wantedToken   string
		wantErr       bool
	}{
		{
			name:          "Valid Authorization Header",
			headers:       http.Header{
				"Authorization": []string{"Bearer good_token"},
			},
			wantedToken:   "good_token",
			wantErr:       false,
		},
		{
			name:          "Missing Authorization Header",
			headers:       http.Header{},
			wantedToken:   "",
			wantErr:       true,
		},
		{
			name:          "Malformed Authorization Header",
			headers:       http.Header{
				"Authorization": []string{"InvalidBearer valid_token"},
			},
			wantedToken:   "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotToken != tt.wantedToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantedToken)
			}
		})
	}
}
