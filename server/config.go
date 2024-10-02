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

// Config is a global configuration store.
var Config struct {
	// LogLevel controls how verbose the levels are. Supported values: trace,
	// debug, info, warn, error, fatal, panic.
	LogLevel string `env:"LOG_LEVEL, default=info"`

	// Host is the host name or IP address used by the simulator to accept
	// connection requests. If no value is provided it will be automatically
	// resolved.
	Host string `env:"API_HOST"`

	// Port defines the port number used by the simulator to accept connection
	// requests.
	Port uint16 `env:"API_PORT, default=7547"`

	// SerialNumber will overwrite the DeviceInfo.SerialNumber datamodel
	// parameter value.
	SerialNumber string `env:"SERIAL_NUMBER, required"`

	// DataModelPath must point to a datamodel file in CSV format.
	DataModelPath string `env:"DATAMODEL_PATH, required"`

	// StateFilePath points to the state file. If the file doesn't exist it will
	// be created and will maintain all changes made to the datamodel. Missing
	// state file will trigger a BOOTSTRAP inform event.
	StateFilePath string `env:"STATE_PATH, default=state.json"`

	// ACSURL is the URL for the ACS.
	ACSURL string `env:"ACS_URL, required"`

	// ACSAuth configures authentication scheme for the ACS. It defaults to
	// "none". Supported values: digest, none
	ACSAuth string `env:"ACS_AUTH, default=none"`

	// ACSUsername is used to authenticate requests to the ACS.
	ACSUsername string `env:"ACS_USERNAME"`

	// ACSPassword is used to authenticate requests to the ACS.
	ACSPassword string `env:"ACS_PASSWORD"`

	// ACSVerifyTLS when set to false ignores certificate errors when connecting
	// to the ACS.
	ACSVerifyTLS bool `env:"ACS_VERIFY_TLS, default=false"`

	// RebootDelay defines how long the simulator should wait and drop incoming
	// connection requests to pretend that it reboots.
	RebootDelay time.Duration `env:"REBOOT_DELAY, default=5s"`

	// UpgradeDelay defines how long the simulator should wait and drop incoming
	// connection requests to pretend that software upgrades take time.
	UpgradeDelay time.Duration `env:"UPGRADE_DELAY, default=15s"`
}

// ErrNoCreds is returned when ACS authentication is configured for digest
// access authentication but no credentials are provided.
var ErrNoCreds = errors.New("username/password missing")

const (
	// AuthDigest an identifier for HTTP digest access authentication.
	AuthDigest = "digest"
	// AuthNone an identifier for no HTTP authentication.
	AuthNone = "none"
)

// LoadConfig attempts to load configuration from environment variables.
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
