package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleDownload(envID string, r *rpc.DownloadRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "Download").Msg("Received message")
	r.Debug()
	resp := rpc.NewEnvelope(envID)

	req, err := http.NewRequest(http.MethodGet, r.URL, nil)
	if err != nil {
		return resp.WithFault(rpc.FaultInternalError)
	}
	if r.Username != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.WithFault(rpc.FaultInternalError)
	}
	if hresp.Body == nil {
		return resp.WithFault(rpc.FaultInternalError)
	}
	defer hresp.Body.Close()
	b, err := io.ReadAll(hresp.Body)
	if err != nil {
		return resp.WithFault(rpc.FaultInternalError)
	}
	var status int
	if r.FileType == rpc.FileTypeFirmwareUpgradeImage {
		status = rpc.DownloadNotCompleted
		var ver struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(b, &ver); err != nil {
			return resp.WithFault(rpc.FaultInternalError)
		}
		if ver.Version != "" {
			s.dm.SetFirmwareVersion(ver.Version)
			// schedule message
		} else {
			return resp.WithFault(rpc.FaultInternalError)
		}
	} else {
		status = rpc.DownloadCompleted
	}

	resp.Body.DownloadResponse = &rpc.DownloadResponseEncoder{
		Status:       status,
		StartTime:    time.Now().Format(time.RFC3339),
		CompleteTime: time.Now().Format(time.RFC3339),
	}
	s.dm.SetCommandKey(r.CommandKey)
	return resp
}
