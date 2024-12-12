package simulator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

// Config is a global configuration store.
// nolint:gochecknoglobals
var Config struct {
	// LogLevel controls how verbose the levels are. Supported values: trace,
	// debug, info, warn, error, fatal, panic.
	LogLevel string `env:"LOG_LEVEL, default=info"`

	// ConnReqEnableHTTP enables an HTTP server that can accept connection
	// requests.
	ConnReqEnableHTTP bool `env:"CR_HTTP, default=true"`

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
	StateFilePath string `env:"STATE_PATH"`

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

	// InformInterval allows to override inform interval in the datamodel.
	InformInterval time.Duration `env:"INFORM_INTERVAL"`

	// NormalizeParameters when set to true will attempt to normalize datamodel
	// parameter types and values in order to bring them closer to the spec.
	NormalizeParameters bool `env:"NORMALIZE_PARAMETERS, default=false"`

	// RebootDelay defines how long the simulator should wait and drop incoming
	// connection requests to pretend that it reboots.
	RebootDelay time.Duration `env:"REBOOT_DELAY, default=5s"`

	// UpgradeDelay defines how long the simulator should wait and drop incoming
	// connection requests to pretend that software upgrades take time.
	UpgradeDelay time.Duration `env:"UPGRADE_DELAY, default=15s"`

	// ConnectionTimeout defines how long it can take to establish a TCP
	// connection with the ACS.
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT, default=5s"`

	// RequestTimeout defines how long request processing could take.
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT, default=5s"`
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
