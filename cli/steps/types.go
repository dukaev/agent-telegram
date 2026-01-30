// Package steps provides common types for login steps.
package steps

// AuthError is sent when authentication fails.
type AuthError struct {
	Step  string
	Error string
}

// PhoneCodeSent is sent when verification code is successfully sent.
type PhoneCodeSent struct {
	Phone    string
	CodeHash string
}

// TwoFARequired is sent when 2FA password is required.
type TwoFARequired struct {
	Hint string
}
