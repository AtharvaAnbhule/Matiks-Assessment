package service

import (

	"fmt"
	"strings"
	"unicode"
)

const (
	MinRating    = 100
	MaxRating    = 5000
	MinUsername  = 3
	MaxUsername  = 50
)

 
func ValidateRating(rating int32) error {
	if rating < MinRating || rating > MaxRating {
		return fmt.Errorf("rating must be between %d and %d, got %d", MinRating, MaxRating, rating)
	}
	return nil
}

 
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if len(username) < MinUsername {
		return fmt.Errorf("username must be at least %d characters", MinUsername)
	}

	if len(username) > MaxUsername {
		return fmt.Errorf("username must not exceed %d characters", MaxUsername)
	}

	 
	for _, ch := range username {
		if !isValidUsernameChar(ch) {
			return fmt.Errorf("username contains invalid character: %c", ch)
		}
	}

	return nil
}

 
func isValidUsernameChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-'
}

 
func SanitizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	return username
}
