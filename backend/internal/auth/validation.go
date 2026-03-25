package auth

import "regexp"

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func IsValidEmail(email string) bool {
	return emailPattern.MatchString(email)
}

func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasDigit := false
	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasDigit = true
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		}
	}

	return hasLetter && hasDigit
}
