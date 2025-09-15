// Package main initializes and runs the TR-069 device simulator, handling
// configuration loading, datamodel setup, and server lifecycle management.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/localhots/blip"
	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/simulator"
)

func main() {
	ctx := context.Background()
	log.Info("Loading configuration")
	if err := simulator.LoadConfig(ctx); err != nil {
		log.Fatal("Failed to load config", log.Cause(err))
	}
	cfg := simulator.Config

	logcfg := blip.DefaultConfig()
	logcfg.Level = blipLevel(cfg.LogLevel)
	log.Setup(logcfg)

	log.Info("Loading datamodel", log.F{"path": cfg.DataModelPath})
	defaults, err := datamodel.LoadDataModelFile(cfg.DataModelPath)
	if err != nil {
		log.Fatal("Failed to load datamodel", log.Cause(err))
	}
	if cfg.NormalizeParameters {
		datamodel.NormalizeParameters(defaults)
	}

	log.Info("Loading state", log.F{"file": cfg.StateFilePath})
	state, err := datamodel.LoadState(cfg.StateFilePath)
	if err != nil {
		log.Fatal("Failed to load state", log.Cause(err))
	}
	dm := datamodel.New(state.WithDefaults(defaults))

	if cfg.SerialNumber != "" {
		dm.SetSerialNumber(cfg.SerialNumber)
	}

	id := dm.DeviceID()
	log.Info("Simulating device", log.F{
		"manufacturer":  id.Manufacturer,
		"oui":           id.OUI,
		"product_class": id.ProductClass,
		"serial_number": id.SerialNumber,
	})

	srv := simulator.New(dm)
	go func() {
		if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server", log.Cause(err))
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Info("Stopping server...")
	if err := srv.Stop(ctx); err != nil {
		log.Fatal("Failed to stop server", log.Cause(err))
	}
}

func blipLevel(level string) blip.Level {
	switch level {
	case "trace":
		return blip.LevelTrace
	case "debug":
		return blip.LevelDebug
	case "info":
		return blip.LevelInfo
	case "warn":
		return blip.LevelWarn
	case "error":
		return blip.LevelError
	case "panic":
		return blip.LevelPanic
	case "fatal":
		return blip.LevelFatal
	default:
		return blip.LevelInfo
	}
}
