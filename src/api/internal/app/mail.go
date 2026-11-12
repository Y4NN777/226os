package app

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// newMailer maps the SMTP configuration onto the SMTP adapter, or returns mailer.Disabled when
// there is none, and logs which one. Only the host and the TLS mode are logged: never the
// username, the password or the sender.
func newMailer(cfg config.Config, logger *slog.Logger) mailer.Mailer {
	if cfg.SMTP == nil {
		logger.Info("mail not configured")
		return mailer.Disabled{}
	}
	logger.Info("mail configured", "host", cfg.SMTP.Host, "tls", cfg.SMTP.TLS)
	return mailer.NewSMTP(smtpConfig(cfg.SMTP))
}

func smtpConfig(s *config.SMTP) mailer.SMTPConfig {
	return mailer.SMTPConfig{Host: s.Host, Port: s.Port, Username: s.Username, Password: s.Password, From: s.From, TLS: mailer.Mode(s.TLS)}
}
