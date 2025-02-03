package constant

// User Roles
const (
	Host   = "HOST"
	Seeker = "SEEKER"
)

// Verification Token Types
const (
	EmailVerification = "EMAIL_VERIFICATION"
	PasswordReset     = "PASSWORD_RESET"
)

// Log messages
const (
	FailedMarshal = "failed to parse request body"
)

const (
	EmailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	PhoneRegex = `^(?:(?:\+91)|91)?[0-9]{10}$`
)
