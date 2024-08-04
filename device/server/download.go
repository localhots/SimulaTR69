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
	log.Debug().Str("url", r.URL).Msg("Downloading file")
	hresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.WithFaultMsg(rpc.FaultInternalError, err.Error())
	}
	if hresp.Body == nil {
		return resp.WithFaultMsg(rpc.FaultInternalError, "firmware file is empty")
	}
	defer hresp.Body.Close()
	b, err := io.ReadAll(hresp.Body)
	if err != nil {
		return resp.WithFaultMsg(rpc.FaultInternalError, err.Error())
	}
	var status int
	if r.FileType == rpc.FileTypeFirmwareUpgradeImage {
		log.Debug().Msg("Parsing firmware file")
		status = rpc.DownloadNotCompleted
		var ver struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(b, &ver); err != nil {
			return resp.WithFault(rpc.FaultInternalError)
		}
		if ver.Version != "" {
			log.Info().Str("version", ver.Version).Msg("Upgrading firmware")
			s.dm.SetFirmwareVersion(ver.Version)
			s.dm.AddEvent(rpc.EventTransferComplete)
			s.dm.AddEvent(rpc.EventBoot)
			status = rpc.DownloadNotCompleted
			s.dm.NotifyParams = append(s.dm.NotifyParams, "DeviceInfo.SoftwareVersion")
			// Stop informs for the upgrade delay duration
			s.dm.SetPeriodicInformTime(time.Now().Add(Config.UpgradeDelay))
			s.ResetInformTimer()
		} else {
			return resp.WithFaultMsg(rpc.FaultInternalError, "incompatible firmware")
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
