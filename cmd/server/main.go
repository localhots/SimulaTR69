package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/server"
	"github.com/localhots/SimulaTR69/server/metrics"
)

func main() {
	ctx := context.Background()
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	})
	if err := server.LoadConfig(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	cfg := server.Config

	log.Info().Str("file", cfg.DataModelPath).Msg("Loading datamodel")
	defaults, err := datamodel.LoadDataModel(cfg.DataModelPath)
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

	srv := server.New(dm, metrics.NewNoop())
	go func() {
		// FIXME: something's off with error checking here
		// nolint:errorlint
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	srv.Inform(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info().Msg("Stopping server...")
	if err := srv.Stop(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to stop server")
	}
}
