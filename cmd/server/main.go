package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/simulator"
)

func main() {
	ctx := context.Background()
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	})

	log.Info().Msg("Loading configuration")
	if err := simulator.LoadConfig(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	cfg := simulator.Config

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse log level")
	}
	log.Logger = log.Logger.Level(logLevel)

	log.Info().Str("file", cfg.DataModelPath).Msg("Loading datamodel")
	defaults, err := datamodel.LoadDataModelFile(cfg.DataModelPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load datamodel")
	}
	if cfg.NormalizeParameters {
		datamodel.NormalizeParameters(defaults)
	}

	log.Info().Str("file", cfg.StateFilePath).Msg("Loading state")
	state, err := datamodel.LoadState(cfg.StateFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load state")
	}
	dm := datamodel.New(state.WithDefaults(defaults))

	if cfg.SerialNumber != "" {
		dm.SetSerialNumber(cfg.SerialNumber)
	}

	id := dm.DeviceID()
	log.Info().
		Str("manufacturer", id.Manufacturer).
		Str("oui", id.OUI).
		Str("product_class", id.ProductClass).
		Str("serial_number", id.SerialNumber).
		Msg("Simulating device")

	srv := simulator.New(dm)
	go func() {
		if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Info().Msg("Stopping server...")
	if err := srv.Stop(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to stop server")
	}
}
