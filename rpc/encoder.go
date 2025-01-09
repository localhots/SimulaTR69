// nolint:revive
package rpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type EnvelopeEncoder struct {
	XMLName      xml.Name `xml:"soapenv:Envelope"`
	XMLSpaceEnv  string   `xml:"xmlns:soapenv,attr"`
	XMLSpaceEnc  string   `xml:"xmlns:soapenc,attr"`
	XMLSpaceXSD  string   `xml:"xmlns:xsd,attr"`
	XMLSpaceXSI  string   `xml:"xmlns:xsi,attr"`
	XMLSpaceCWMP string   `xml:"xmlns:cwmp,attr"`

	Header HeaderEncoder `xml:"soapenv:Header"`
	Body   BodyEncoder   `xml:"soapenv:Body"`
}

type HeaderEncoder struct {
	ID IDEncoder `xml:"cwmp:ID"`
}

type IDEncoder struct {
	MustUnderstand int    `xml:"soapenv:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type BodyEncoder struct {
	Inform                            *InformRequestEncoder                     `xml:"cwmp:Inform,omitempty"`
	GetRPCMethodsResponse             *GetRPCMethodsResponseEncoder             `xml:"cwmp:GetRPCMethodsResponse,omitempty"`
	SetParameterValuesResponse        *SetParameterValuesResponseEncoder        `xml:"cwmp:SetParameterValuesResponse,omitempty"`
	GetParameterValuesResponse        *GetParameterValuesResponseEncoder        `xml:"cwmp:GetParameterValuesResponse,omitempty"`
	GetParameterNamesResponse         *GetParameterNamesResponseEncoder         `xml:"cwmp:GetParameterNamesResponse,omitempty"`
	SetParameterAttributesResponse    *SetParameterAttributesResponseEncoder    `xml:"cwmp:SetParameterAttributesResponse,omitempty"`
	GetParameterAttributesResponse    *GetParameterAttributesResponseEncoder    `xml:"cwmp:GetParameterAttributesResponse,omitempty"`
	AddObjectResponse                 *AddObjectResponseEncoder                 `xml:"cwmp:AddObjectResponse,omitempty"`
	DeleteObjectResponse              *DeleteObjectResponseEncoder              `xml:"cwmp:DeleteObjectResponse,omitempty"`
	RebootResponse                    *RebootResponseEncoder                    `xml:"cwmp:RebootResponse,omitempty"`
	DownloadResponse                  *DownloadResponseEncoder                  `xml:"cwmp:DownloadResponse,omitempty"`
	FactoryResetResponse              *FactoryResetResponseEncoder              `xml:"cwmp:FactoryResetResponse,omitempty"`
	TransferCompleteRequest           *TransferCompleteRequestEncoder           `xml:"cwmp:TransferComplete,omitempty"`
	AutonomousTransferCompleteRequest *AutonomousTransferCompleteRequestEncoder `xml:"cwmp:AutonomousTransferComplete,omitempty"`
	Fault                             *FaultEncoder                             `xml:"soapenv:Fault,omitempty"`
}

//
// Payloads
//

// nolint:stylecheck
type InformRequestEncoder struct {
	DeviceId      DeviceID
	Event         EventEncoder
	MaxEnvelopes  int
	CurrentTime   string
	RetryCount    int
	ParameterList ParameterListEncoder
}

type SetParameterValuesResponseEncoder struct {
	Status int
}

type GetParameterValuesResponseEncoder struct {
	ParameterList ParameterListEncoder
}

type GetParameterNamesResponseEncoder struct {
	ParameterList ParameterInfoEncoder
}

type AddObjectResponseEncoder struct {
	InstanceNumber int
	Status         int
}

type DeleteObjectResponseEncoder struct {
	Status int
}

type ParameterInfoEncoder struct {
	ArrayType  string                `xml:"soapenc:arrayType,attr"`
	Parameters []ParameterInfoStruct `xml:"ParameterInfoStruct"`
}

type GetRPCMethodsResponseEncoder struct {
	MethodList MethodListEncoder
}

type SetParameterAttributesResponseEncoder struct{}

type GetParameterAttributesResponseEncoder struct {
	ParameterList ParameterAttributeStructEncoder
}

type RebootResponseEncoder struct{}

type DownloadResponseEncoder struct {
	Status       int
	StartTime    string
	CompleteTime string
}

type FactoryResetResponseEncoder struct{}

type TransferCompleteRequestEncoder struct {
	CommandKey   string
	Fault        any `xml:"FaultStruct,omitempty"`
	StartTime    string
	CompleteTime string
}

type AutonomousTransferCompleteRequestEncoder struct {
	AnnounceURL    string
	TransferURL    string
	IsDownload     bool
	FileType       string
	FileSize       uint
	TargetFileName string
	Fault          any `xml:"FaultStruct,omitempty"`
	StartTime      string
	CompleteTime   string
}

type FaultEncoder struct {
	FaultCode   string             `xml:"faultcode"`
	FaultString string             `xml:"faultstring"`
	Detail      FaultDetailEncoder `xml:"detail"`
}

type FaultDetailEncoder struct {
	Fault FaultStruct `xml:"cwmp:Fault"`
}

//
// Embedded structs
//

type ParameterAttributeStructEncoder struct {
	ArrayType           string                     `xml:"soapenc:arrayType,attr"`
	ParameterAttributes []ParameterAttributeStruct `xml:"ParameterAttributeStruct"`
}

type ParameterAttributeStruct struct {
	Name         string
	Notification AttributeNotification
	AccessList   AccessListEncoder
}

type AccessListEncoder struct {
	ArrayType string   `xml:"soapenc:arrayType,attr"`
	Values    []string `xml:"string"`
}

type MethodListEncoder struct {
	ArrayType string   `xml:"soapenc:arrayType,attr"`
	Methods   []string `xml:"string"`
}

type EventEncoder struct {
	ArrayType string        `xml:"soapenc:arrayType,attr"`
	Events    []EventStruct `xml:"EventStruct"`
}

type ParameterListEncoder struct {
	ArrayType       string                  `xml:"soapenc:arrayType,attr"`
	ParameterValues []ParameterValueEncoder `xml:"ParameterValueStruct"`
}

type ParameterValueEncoder struct {
	Name  string
	Value ValueEncoder
}

type ValueEncoder struct {
	Type  string `xml:"xsi:type,attr"`
	Value string `xml:",chardata"`
}

type ParameterInfoStruct struct {
	Name     string
	Writable bool
}

type FaultStruct struct {
	FaultCode               FaultCode
	FaultString             string
	SetParameterValuesFault []SetParameterValuesFault
}

type SetParameterValuesFault struct {
	ParameterName string
	FaultCode     FaultCode
	FaultString   string
}

type NoFaultStruct struct {
	Status int `xml:",chardata"`
}

func NewFaultResponse(code FaultCode, str string) *FaultEncoder {
	return &FaultEncoder{
		FaultCode:   "Client",
		FaultString: "CWMP fault",
		Detail: FaultDetailEncoder{
			Fault: FaultStruct{
				FaultCode:   code,
				FaultString: str,
			},
		},
	}
}

func NewEnvelope(id string) *EnvelopeEncoder {
	return &EnvelopeEncoder{
		XMLSpaceEnv:  NSEnv,
		XMLSpaceEnc:  NSEnc,
		XMLSpaceXSD:  NSXSD,
		XMLSpaceXSI:  NSXSI,
		XMLSpaceCWMP: NSCWMP,
		Header: HeaderEncoder{
			ID: IDEncoder{
				MustUnderstand: 1,
				Value:          id,
			},
		},
	}
}

// nolint:gocyclo
func (ee *EnvelopeEncoder) Method() string {
	switch {
	case ee == nil:
		return "Empty"
	case ee.Body.Inform != nil:
		return "Inform"
	case ee.Body.GetRPCMethodsResponse != nil:
		return "GetRPCMethodsResponse"
	case ee.Body.SetParameterValuesResponse != nil:
		return "SetParameterValuesResponse"
	case ee.Body.GetParameterValuesResponse != nil:
		return "GetParameterValuesResponse"
	case ee.Body.GetParameterNamesResponse != nil:
		return "GetParameterNamesResponse"
	case ee.Body.SetParameterAttributesResponse != nil:
		return "SetParameterAttributesResponse"
	case ee.Body.GetParameterAttributesResponse != nil:
		return "GetParameterAttributesResponse"
	case ee.Body.AddObjectResponse != nil:
		return "AddObjectResponse"
	case ee.Body.DeleteObjectResponse != nil:
		return "DeleteObjectResponse"
	case ee.Body.RebootResponse != nil:
		return "RebootResponse"
	case ee.Body.DownloadResponse != nil:
		return "DownloadResponse"
	case ee.Body.FactoryResetResponse != nil:
		return "FactoryResetResponse"
	case ee.Body.TransferCompleteRequest != nil:
		return "TransferCompleteRequest"
	case ee.Body.AutonomousTransferCompleteRequest != nil:
		return "AutonomousTransferCompleteRequest"
	case ee.Body.Fault != nil:
		return "Fault"
	default:
		return "None"
	}
}

func (ee *EnvelopeEncoder) WithFault(fault FaultCode) *EnvelopeEncoder {
	ee.Body.Fault = NewFaultResponse(fault, fault.String())
	return ee
}

func (ee *EnvelopeEncoder) WithFaultMsg(fault FaultCode, msg string) *EnvelopeEncoder {
	ee.Body.Fault = NewFaultResponse(fault, msg)
	return ee
}

func (ee EnvelopeEncoder) Encode() ([]byte, error) {
	return ee.encode(false)
}

func (ee EnvelopeEncoder) EncodePretty() ([]byte, error) {
	return ee.encode(true)
}

func (ee EnvelopeEncoder) encode(pretty bool) ([]byte, error) {
	buf := bytes.Buffer{}
	if _, err := buf.WriteString(xml.Header); err != nil {
		return nil, fmt.Errorf("write xml header: %w", err)
	}
	enc := xml.NewEncoder(&buf)
	if pretty {
		enc.Indent("", "    ")
	}
	if err := enc.Encode(ee); err != nil {
		return nil, fmt.Errorf("encode envelope: %w", err)
	}
	if pretty {
		if _, err := buf.WriteRune('\n'); err != nil {
			return nil, fmt.Errorf("write trailing newline: %w", err)
		}
	}
	return buf.Bytes(), nil
}
