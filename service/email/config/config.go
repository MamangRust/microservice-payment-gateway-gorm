package config

// Config holds the configuration for the email service.
type Config struct {
	KafkaBrokers []string
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
}
