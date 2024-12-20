package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/simulator"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	})
	if err := simulator.LoadConfig(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	cfg := simulator.Config

	log.Info().Str("file", cfg.DataModelPath).Msg("Loading datamodel")
	defaults, err := datamodel.LoadDataModelFile(cfg.DataModelPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load datamodel")
	}
	if cfg.NormalizeParameters {
		datamodel.NormalizeParameters(defaults)
	}

	log.Info().Int("sim_number", cfg.SimNumber).Msg("Starting simulators")

	var wg sync.WaitGroup

	for i := 0; i < cfg.SimNumber; i++ {
		wg.Add(1)

		log.Info().Str("file", cfg.StateFilePath).Msg("Loading state")
		state, err := datamodel.LoadState(cfg.StateFilePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load state")
		}

		dm := datamodel.New(state.WithDefaults(defaults))

		if cfg.SerialNumber != "" {
			dm.SetSerialNumber(cfg.SerialNumber + "-" + strconv.Itoa(i))
		}

		id := dm.DeviceID()
		log.Info().
			Str("manufacturer", id.Manufacturer).
			Str("oui", id.OUI).
			Str("product_class", id.ProductClass).
			Str("serial_number", id.SerialNumber).
			Msg("Simulating device")

		srv := simulator.New(dm, simulator.Config.ConnectionRequestPort+uint16(i))
		go func() {
			// defer srv.Stop(ctx)

			// FIXME: something's off with error checking here
			// nolint:errorlint
			if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start server")
			}

			<-ctx.Done()

			log.Info().Str("id", id.SerialNumber).Msg("Stopping simulated device")
			srv.Stop(ctx)
			wg.Done()

			// TODO: should we stop the server ? srv.Stop(ctx)
		}()

	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	wg.Wait()

	log.Info().Msg("All simulators stopped, shutting down")
}
