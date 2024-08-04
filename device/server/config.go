package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

const (
	AuthDigest = "digest"
	AuthNone   = "none"
)

var ErrNoCreds = errors.New("username/password missing")

var Config struct {
	Host          string        `env:"API_HOST"`
	Port          uint16        `env:"API_PORT, default=7547"`
	ACSURL        string        `env:"ACS_URL, required"`
	SerialNumber  string        `env:"SERIAL_NUMBER, required"`
	DataModelPath string        `env:"DATAMODEL_PATH, required"`
	StateFilePath string        `env:"STATE_PATH, default=state.json"`
	UpgradeDelay  time.Duration `env:"UPGRADE_DELAY, default=15s"`
	ACSAuth       string        `env:"ACS_AUTH, default=none"`
	ACSUsername   string        `env:"ACS_USERNAME"`
	ACSPassword   string        `env:"ACS_PASSWORD"`
	LogLevel      string        `env:"LOG_LEVEL, default=info"`
	RebootDelay time.Duration `env:"REBOOT_DELAY, default=5s"`
}

func LoadConfig(ctx context.Context) error {
	log.Info().Msg("Loading configuration")
	err := envconfig.Process(ctx, &Config)
	if err != nil {
		return fmt.Errorf("load env config: %w", err)
	}

	if Config.Host == "" {
		Config.Host, err = getIP()
		if err != nil {
			return fmt.Errorf("get ip address: %w", err)
		}
	}

	if Config.ACSAuth != AuthNone {
		if Config.ACSUsername == "" || Config.ACSPassword == "" {
			return fmt.Errorf("auth %s: %w", Config.ACSAuth, ErrNoCreds)
		}
	}

	logLevel, err := zerolog.ParseLevel(Config.LogLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	log.Logger = log.Logger.Level(logLevel)

	return nil
}

func getIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "0.0.0.0", nil
}
