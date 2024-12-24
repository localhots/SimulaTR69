// nolint:revive
package rpc

import (
	"encoding/xml"
	"fmt"

	"github.com/rs/zerolog/log"
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

func (r SetParameterValuesRequest) Debug() {
	for _, v := range r.ParameterList.ParameterValues {
		log.Debug().
			Str("name", v.Name).
			Str("type", v.Value.Type).
			Str("value", v.Value.Value).
			Msg("SetParameterValues")
	}
}

type GetParameterValuesRequest struct {
	ParameterNames ParameterNames
}

func (r GetParameterValuesRequest) Debug() {
	for _, name := range r.ParameterNames.Names {
		log.Debug().Str("name", name).Msg("GetParameterValues")
	}
}

type GetParameterNamesRequest struct {
	ParameterPath string
	NextLevel     bool
}

func (r GetParameterNamesRequest) Debug() {
	log.Debug().
		Str("name", r.ParameterPath).
		Bool("next_level", r.NextLevel).
		Msg("GetParameterNames")
}

type SetParameterAttributesRequest struct {
	ParameterList struct {
		ArrayType           string                         `xml:"arrayType,attr"`
		ParameterAttributes []SetParameterAttributesStruct `xml:"SetParameterAttributesStruct"`
	}
}

func (r SetParameterAttributesRequest) Debug() {
	for _, attr := range r.ParameterList.ParameterAttributes {
		log.Debug().
			Str("name", attr.Name).
			Int("notification", int(attr.Notification)).
			Bool("notification_change", attr.NotificationChange).
			Strs("access_list", attr.AccessList.Values).
			Bool("access_list_change", attr.AccessListChange).
			Msg("SetParameterAttributes")
	}
}

type GetParameterAttributesRequest struct {
	ParameterNames ParameterNames
}

func (r GetParameterAttributesRequest) Debug() {
	for _, path := range r.ParameterNames.Names {
		log.Debug().Str("name", path).Msg("GetParameterAttributes")
	}
}

type AddObjectRequest struct {
	ObjectName   string
	ParameterKey string
}

func (r AddObjectRequest) Debug() {
	log.Debug().Str("path", r.ObjectName).Msg("AddObjectRequest")
}

type DeleteObjectRequest struct {
	ObjectName   string
	ParameterKey string
}

func (r DeleteObjectRequest) Debug() {
	log.Debug().Str("path", r.ObjectName).Msg("DeleteObjectRequest")
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

func (r DownloadRequest) Debug() {
	log.Debug().
		Str("file_type", r.FileType).
		Str("url", r.URL).
		Int("file_size", r.FileSize).
		Msg("DownloadRequest")
}

type UploadRequest struct {
	CommandKey string
}

func (r UploadRequest) Debug() {
	log.Debug().Str("command_key", r.CommandKey).Msg("UploadRequest")
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
