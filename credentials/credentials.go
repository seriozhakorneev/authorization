package credentials

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

// Credentials модель данных необходимых для авторизации
type Credentials struct {
	LoginURL        string `yaml:"login_url"`
	RuCaptchaApiKey string `yaml:"ru-captcha-api-key"`
	Login           string `yaml:"login"`
	Password        string `yaml:"password"`
}

// NewCredentials возвращает необходиммые данные
// авторизации снятые с конфига
func NewCredentials() (*Credentials, error) {
	cfg := &Credentials{}

	err := cleanenv.ReadConfig("./credentials/credentials.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("cleanenv.ReadConfig: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("cleanenv.ReadEnv: %w", err)
	}

	return cfg, nil
}
