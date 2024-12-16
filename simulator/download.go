package simulator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleDownload(envID string, r *rpc.DownloadRequest) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "Download").Msg("Received message")
	r.Debug()
	resp := rpc.NewEnvelope(envID)

	resp.Body.DownloadResponse = &rpc.DownloadResponseEncoder{
		Status:       rpc.DownloadNotCompleted,
		StartTime:    time.Now().Format(time.RFC3339),
		CompleteTime: time.Now().Format(time.RFC3339),
	}
	s.dm.SetCommandKey(r.CommandKey)
	go s.asyncDownload(r)

	return resp
}

func (s *Simulator) asyncDownload(r *rpc.DownloadRequest) {
	tcr := rpc.TransferCompleteRequestEncoder{
		CommandKey: s.dm.CommandKey(),
		StartTime:  time.Now().UTC().Format(time.RFC3339),
		Fault:      &rpc.FaultStruct{},
	}
	err := s.upgradeFirmware(r)
	tcr.CompleteTime = time.Now().UTC().Format(time.RFC3339)
	if err != nil {
		tcr.Fault = &rpc.FaultStruct{
			FaultCode:   rpc.FaultInternalError,
			FaultString: err.Error(),
		}
	}

	s.transferComplete <- tcr
}

func (s *Simulator) upgradeFirmware(r *rpc.DownloadRequest) error {
	req, err := http.NewRequest(http.MethodGet, r.URL, nil)
	if err != nil {
		return fmt.Errorf("create new request: %w", err)
	}
	if r.Username != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	log.Debug().Str("url", r.URL).Msg("Downloading file")
	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	if hresp.Body == nil {
		return errors.New("empty download")
	}
	defer hresp.Body.Close()
	b, err := io.ReadAll(hresp.Body)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if r.FileType != rpc.FileTypeFirmwareUpgradeImage {
		return nil
	}

	log.Debug().Msg("Parsing firmware file")
	var ver struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &ver); err != nil {
		return fmt.Errorf("parse firmware upgrade file: %w", err)
	}
	if ver.Version == "" {
		return errors.New("incompatible firmware")
	}

	log.Info().Str("version", ver.Version).Msg("Upgrading firmware")
	s.dm.SetFirmwareVersion(ver.Version)
	return nil
}
