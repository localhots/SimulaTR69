//nolint:revive
package rpc

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/localhots/blip"
	"github.com/localhots/blip/noctx/log"
)

type EnvelopeDecoder struct {
	XMLName xml.Name `xml:"Envelope"`
	Header  HeaderDecoder
	Body    BodyDecoder
}

type HeaderDecoder struct {
	ID IDDecoder
}

type IDDecoder struct {
	MustUnderstand int    `xml:"mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type BodyDecoder struct {
	GetRPCMethods          *EmptyPayload
	SetParameterValues     *SetParameterValuesRequest
	GetParameterValues     *GetParameterValuesRequest
	GetParameterNames      *GetParameterNamesRequest
	SetParameterAttributes *SetParameterAttributesRequest
	GetParameterAttributes *GetParameterAttributesRequest
	AddObject              *AddObjectRequest
	DeleteObject           *DeleteObjectRequest
	Reboot                 *RebootRequest
	Download               *DownloadRequest
	Upload                 *UploadRequest
	FactoryReset           *EmptyPayload
	GetQueuedTransfers     *EmptyPayload
	GetAllQueuedTransfers  *EmptyPayload
	ScheduleInform         *ScheduleInformRequest
	SetVouchers            *SetVouchersRequest
	GetOptions             *GetOptionsRequest

	InformResponse                     *InformResponse
	TransferCompleteResponse           *EmptyPayload
	AutonomousTransferCompleteResponse *EmptyPayload
	Fault                              *FaultPayload
}

//
// Request payloads
//

type SetParameterValuesRequest struct {
	ParameterList struct {
		ArrayType       string                  `xml:"arrayType,attr"`
		ParameterValues []ParameterValueDecoder `xml:"ParameterValueStruct"`
	}
	ParameterKey string
}

func (r SetParameterValuesRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "SetParameterValues"})
	for _, v := range r.ParameterList.ParameterValues {
		logger.Debug(ctx, "SetParameterValues", log.F{
			"name":  v.Name,
			"type":  v.Value.Type,
			"value": v.Value.Value,
		})
	}
}

type GetParameterValuesRequest struct {
	ParameterNames ParameterNames
}

func (r GetParameterValuesRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "GetParameterValues"})
	for _, name := range r.ParameterNames.Names {
		logger.Debug(ctx, "GetParameterValues", log.F{"name": name})
	}
}

type GetParameterNamesRequest struct {
	ParameterPath string
	NextLevel     bool
}

func (r GetParameterNamesRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "GetParameterNames"})
	logger.Debug(ctx, "GetParameterNames", log.F{
		"name":       r.ParameterPath,
		"next_level": r.NextLevel,
	})
}

type SetParameterAttributesRequest struct {
	ParameterList struct {
		ArrayType           string                         `xml:"arrayType,attr"`
		ParameterAttributes []SetParameterAttributesStruct `xml:"SetParameterAttributesStruct"`
	}
}

func (r SetParameterAttributesRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "SetParameterAttributes"})
	for _, attr := range r.ParameterList.ParameterAttributes {
		logger.Debug(ctx, "SetParameterAttributes", log.F{
			"name":                attr.Name,
			"notification":        int(attr.Notification),
			"notification_change": attr.NotificationChange,
			"access_list":         attr.AccessList.Values,
			"access_list_change":  attr.AccessListChange,
		})
	}
}

type GetParameterAttributesRequest struct {
	ParameterNames ParameterNames
}

func (r GetParameterAttributesRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "GetParameterAttributes"})
	for _, path := range r.ParameterNames.Names {
		logger.Debug(ctx, "GetParameterAttributes", log.F{"name": path})
	}
}

type AddObjectRequest struct {
	ObjectName   string
	ParameterKey string
}

func (r AddObjectRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "AddObject"})
	logger.Debug(ctx, "AddObjectRequest", log.F{"path": r.ObjectName})
}

type DeleteObjectRequest struct {
	ObjectName   string
	ParameterKey string
}

func (r DeleteObjectRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "DeleteObject"})
	logger.Debug(ctx, "DeleteObjectRequest", log.F{"path": r.ObjectName})
}

type RebootRequest struct {
	CommandKey string
}

type DownloadRequest struct {
	CommandKey     string
	FileType       string
	URL            string
	Username       string
	Password       string
	FileSize       int
	TargetFileName string
	DelaySeconds   int
	SuccessURL     string
	FailureURL     string
}

func (r DownloadRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "Download"})
	logger.Debug(ctx, "DownloadRequest", log.F{
		"file_type": r.FileType,
		"url":       r.URL,
		"file_size": r.FileSize,
	})
}

type UploadRequest struct {
	CommandKey string
}

func (r UploadRequest) Debug(ctx context.Context, logger *blip.Logger) {
	logger.Info(ctx, "Received message", log.F{"method": "Upload"})
	logger.Debug(ctx, "UploadRequest", log.F{"command_key": r.CommandKey})
}

type ScheduleInformRequest struct {
	DelaySeconds int64
	CommandKey   string
}

type SetVouchersRequest struct {
	VoucherList struct {
		ArrayType string   `xml:"arrayType,attr"`
		Values    []string `xml:"base64"`
	}
}

type GetOptionsRequest struct {
	OptionName string
}

type EmptyPayload struct{}

type InformResponse struct {
	MaxEnvelopes int
}

type FaultPayload struct {
	FaultCode   string             `xml:"faultcode"`
	FaultString string             `xml:"faultstring"`
	Detail      FaultDetailPayload `xml:"detail"`
}

type FaultDetailPayload struct {
	Fault FaultStruct
}

//
// Embedded structs
//

type ParameterNames struct {
	ArrayType string   `xml:"arrayType,attr"`
	Names     []string `xml:"string"`
}

type ParameterValueDecoder struct {
	Name  string
	Value struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	}
}

type SetParameterAttributesStruct struct {
	Name               string
	NotificationChange bool
	Notification       AttributeNotification
	AccessListChange   bool
	AccessList         struct {
		ArrayType string   `xml:"arrayType,attr"`
		Values    []string `xml:"string"`
	}
}

// Decode attempts to decode given payload into a SOAP envelope.
func Decode(b []byte) (*EnvelopeDecoder, error) {
	var env EnvelopeDecoder
	err := xml.Unmarshal(b, &env)
	if err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	return &env, nil
}

//nolint:gocyclo
func (env EnvelopeDecoder) Method() string {
	switch {
	case env.Body.GetRPCMethods != nil:
		return "GetRPCMethods"
	case env.Body.SetParameterValues != nil:
		return "SetParameterValues"
	case env.Body.GetParameterValues != nil:
		return "GetParameterValues"
	case env.Body.GetParameterNames != nil:
		return "GetParameterNames"
	case env.Body.SetParameterAttributes != nil:
		return "SetParameterAttributes"
	case env.Body.GetParameterAttributes != nil:
		return "GetParameterAttributes"
	case env.Body.AddObject != nil:
		return "AddObject"
	case env.Body.DeleteObject != nil:
		return "DeleteObject"
	case env.Body.Reboot != nil:
		return "Reboot"
	case env.Body.Download != nil:
		return "Download"
	case env.Body.Upload != nil:
		return "Upload"
	case env.Body.FactoryReset != nil:
		return "FactoryReset"
	case env.Body.GetQueuedTransfers != nil:
		return "GetQueuedTransfers"
	case env.Body.GetAllQueuedTransfers != nil:
		return "GetAllQueuedTransfers"
	case env.Body.ScheduleInform != nil:
		return "ScheduleInform"
	case env.Body.SetVouchers != nil:
		return "SetVouchers"
	case env.Body.GetOptions != nil:
		return "GetOptions"
	case env.Body.Fault != nil:
		return "Fault"
	case env.Body.TransferCompleteResponse != nil:
		return "TransferCompleteResponse"
	default:
		return "Unknown"
	}
}
