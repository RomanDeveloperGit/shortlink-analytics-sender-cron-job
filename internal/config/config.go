package config

import (
	"time"

	"github.com/caarlos0/env/v11"

	internalEnv "github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/config/env"
)

type Config struct {
	Env            internalEnv.Env `env:"ENV,required,notEmpty"`
	ServiceName    string          `env:"SERVICE_NAME,required,notEmpty"`
	ServiceVersion string          `env:"SERVICE_VERSION,required,notEmpty"`
	Cron           Cron
	AnalyticsAPI   AnalyticsAPI
	SMTPServer     SMTPServer
}

type Cron struct {
	Schedule string `env:"CRON_SCHEDULE,required,notEmpty"`
	Zone     string `env:"CRON_ZONE,required,notEmpty"`
}

type AnalyticsAPI struct {
	Host           string        `env:"ANALYTICS_API_HOST,required,notEmpty"`
	Port           int           `env:"ANALYTICS_API_PORT,required,notEmpty"`
	KeepAliveTime  time.Duration `env:"ANALYTICS_API_KEEP_ALIVE_TIME,required,notEmpty"`
	ConnectTimeout time.Duration `env:"ANALYTICS_API_CONNECT_TIMEOUT,required,notEmpty"`
}

type SMTPServer struct {
	Host   string `env:"SMTP_HOST,required,notEmpty"`
	Port   int    `env:"SMTP_PORT,required,notEmpty"`
	Sender string `env:"SMTP_SENDER,required,notEmpty"`
	// Username string `env:"SMTP_USERNAME,required,notEmpty"`
	// Password string `env:"SMTP_PASSWORD,required,notEmpty"`
}

func MustLoad() *Config {
	cfg := &Config{}

	// Специально через Must, чтобы вылезла паника в случае некорректного конфига
	cfg = env.Must(cfg, env.Parse(cfg))

	if !internalEnv.IsEnv(string(cfg.Env)) {
		panic("failed to load config: env is not supported")
	}

	return cfg
}
