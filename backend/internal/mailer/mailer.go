package mailer

// Sender sends transactional emails (e.g. verification).
type Sender interface {
	// SendVerificationEmail sends an email with a verification link to the given address.
	SendVerificationEmail(to, verificationLink string) error
}
