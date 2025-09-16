package simulator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleDownload(ctx context.Context, envID string, r *rpc.DownloadRequest) *rpc.EnvelopeEncoder {
	resp := rpc.NewEnvelope(envID)
	resp.Body.DownloadResponse = &rpc.DownloadResponseEncoder{
		Status:       rpc.DownloadNotCompleted,
		StartTime:    time.Now().Format(time.RFC3339),
		CompleteTime: time.Now().Format(time.RFC3339),
	}
	s.dm.SetCommandKey(r.CommandKey)

	s.tasks <- func() taskFn {
		tcr := rpc.TransferCompleteRequestEncoder{
			CommandKey: s.dm.CommandKey(),
			StartTime:  time.Now().UTC().Format(time.RFC3339),
			Fault:      &rpc.FaultStruct{},
		}
		err := s.upgradeFirmware(ctx, r)
		tcr.CompleteTime = time.Now().UTC().Format(time.RFC3339)
		if err != nil {
			tcr.Fault = &rpc.FaultStruct{
				FaultCode:   rpc.FaultInternalError,
				FaultString: err.Error(),
			}
		}

		s.pendingRequests <- func(env *rpc.EnvelopeEncoder) {
			env.Body.TransferCompleteRequest = &tcr
		}
		s.pendingEvents <- rpc.EventTransferComplete

		return func() taskFn {
			s.logger.Debug(ctx, "Simulating firmware upgrade", log.F{"delay": Config.UpgradeDelay})
			s.pretendOfflineFor(Config.UpgradeDelay)
			s.logger.Debug(ctx, "Starting up")
			s.pendingEvents <- rpc.EventBoot
			return nil
		}
	}

	return resp
}

func (s *Simulator) upgradeFirmware(ctx context.Context, r *rpc.DownloadRequest) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL, nil)
	if err != nil {
		return fmt.Errorf("create new request: %w", err)
	}
	if r.Username != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	s.logger.Debug(ctx, "Downloading file", log.F{"url": r.URL})
	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	if hresp.Body == nil {
		return errors.New("empty download")
	}
	defer func() {
		if err := hresp.Body.Close(); err != nil {
			s.logger.Error(ctx, "Failed to close response body", log.Cause(err))
		}
	}()
	b, err := io.ReadAll(hresp.Body)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if r.FileType != rpc.FileTypeFirmwareUpgradeImage {
		return nil
	}

	s.logger.Debug(ctx, "Parsing firmware file")
	var ver struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &ver); err != nil {
		return fmt.Errorf("parse firmware upgrade file: %w", err)
	}
	if ver.Version == "" {
		return errors.New("incompatible firmware")
	}

	s.logger.Info(ctx, "Upgrading firmware", log.F{"version": ver.Version})
	s.dm.SetFirmwareVersion(ver.Version)
	return nil
}
