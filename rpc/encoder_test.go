package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeInformRequest(t *testing.T) {
	env := NewEnvelope("123")
	vals := []ParameterValueEncoder{
		{Name: "Device.DeviceInfo.HardwareVersion", Value: ValueEncoder{Type: XSD(TypeString), Value: "1.0"}},
		{Name: "Device.DeviceInfo.ProvisioningCode", Value: ValueEncoder{Type: XSD(TypeString), Value: "provisioning.code"}},
		{Name: "Device.DeviceInfo.SoftwareVersion", Value: ValueEncoder{Type: XSD(TypeString), Value: "G3000E-1.2.3"}},
		{Name: "Device.ManagementServer.AliasBasedAddressing", Value: ValueEncoder{Type: XSD(TypeBoolean), Value: "1"}},
		{Name: "Device.ManagementServer.ConnectionRequestURL", Value: ValueEncoder{Type: XSD(TypeString), Value: "http://192.168.1.1:7547/acs"}},
		{Name: "Device.ManagementServer.ParameterKey", Value: ValueEncoder{Type: XSD(TypeString), Value: "n/a"}},
		{Name: "Device.RootDataModelVersion", Value: ValueEncoder{Type: XSD(TypeString), Value: "2.11"}},
		{Name: "Device.X_ACME_WANDetection.PPPUserName", Value: ValueEncoder{Type: XSD(TypeString), Value: "username"}},
		{Name: "Device.X_ACME_WANDetection.WANIPAddress", Value: ValueEncoder{Type: XSD(TypeString), Value: "192.168.1.1"}},
		{Name: "Device.X_ACME_WANDetection.WANMACAddress", Value: ValueEncoder{Type: XSD(TypeString), Value: "de:ca:de:11:22:33"}},
	}
	env.Body.Inform = &InformRequestEncoder{
		DeviceId: DeviceID{
			Manufacturer: "ACME Networks",
			OUI:          "DECADE",
			ProductClass: "G3000E",
			SerialNumber: "G3000E-9799109101",
		},
		Event: EventEncoder{
			ArrayType: ArrayType("cwmp:EventStruct", 1),
			Events: []EventStruct{
				{
					EventCode: EventBootstrap,
				},
			},
		},
		MaxEnvelopes: 1,
		CurrentTime:  "2024-06-10T01:33:00Z",
		RetryCount:   0,
		ParameterList: ParameterListEncoder{
			ArrayType:       ArrayType("cwmp:ParameterValueStruct", len(vals)),
			ParameterValues: vals,
		},
	}
	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(informRequestTestData), string(b))
}

func TestEncodeGetRPCMethodsResponse(t *testing.T) {
	env := NewEnvelope("123")
	methods := []string{
		"GetRPCMethods",
		"SetParameterValues",
		"GetParameterValues",
		"GetParameterNames",
		"SetParameterAttributes",
		"GetParameterAttributes",
		"AddObject",
		"DeleteObject",
		"Download",
		"Reboot",
		"ScheduleInform",
		"FactoryReset",
	}
	env.Body.GetRPCMethodsResponse = &GetRPCMethodsResponseEncoder{
		MethodList: MethodListEncoder{
			ArrayType: ArrayType(XSD(TypeString), len(methods)),
			Methods:   methods,
		},
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(getRPCMethodsResponseTestData), string(b))
}

func TestEncodeSetParameterValuesResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.SetParameterValuesResponse = &SetParameterValuesResponseEncoder{
		Status: 1,
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(setParameterValuesResponseTestData), string(b))
}

func TestEncodeGetParameterValuesResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.GetParameterValuesResponse = &GetParameterValuesResponseEncoder{
		ParameterList: ParameterListEncoder{
			ArrayType: ArrayType("cwmp:ParameterValueStruct", 1),
			ParameterValues: []ParameterValueEncoder{
				{
					Name: "Device.DeviceInfo.VendorConfigFile.1.Version",
					Value: ValueEncoder{
						Type:  XSD(TypeString),
						Value: "0.0",
					},
				},
			},
		},
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(getParameterValuesResponseTestData), string(b))
}

func TestEncodeGetParameterValuesFaultResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.WithFault(FaultInvalidParameterName)

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(getParameterValuesFaultResponseTestData), string(b))
}

func TestEncodeGetParameterNamesResponse(t *testing.T) {
	env := NewEnvelope("123")
	params := []ParameterInfoStruct{
		{Name: "Device.", Writable: false},
		{Name: "Device.RootDataModelVersion", Writable: false},
		{Name: "Device.InterfaceStackNumberOfEntries", Writable: false},
		{Name: "Device.CaptivePortal.", Writable: false},
		{Name: "Device.CaptivePortal.Enable", Writable: true},
	}
	env.Body.GetParameterNamesResponse = &GetParameterNamesResponseEncoder{
		ParameterList: ParameterInfoEncoder{
			ArrayType:  ArrayType("cwmp:ParameterInfoStruct", len(params)),
			Parameters: params,
		},
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(getParameterNamesResponseTestData), string(b))
}

func TestEncodeSetParameterAttributesResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.SetParameterAttributesResponse = &SetParameterAttributesResponseEncoder{}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(setParameterAttributesResponseTestData), string(b))
}

func TestEncodeGetParameterAttributesResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.GetParameterAttributesResponse = &GetParameterAttributesResponseEncoder{
		ParameterList: ParameterAttributeStructEncoder{
			ArrayType: ArrayType("cwmp:ParameterAttributeStruct", 1),
			ParameterAttributes: []ParameterAttributeStruct{
				{
					Name:         "Device.DeviceSummary",
					Notification: 1,
					AccessList: AccessListEncoder{
						ArrayType: ArrayType(XSD(TypeString), 1),
						Values:    []string{"Subscriber"},
					},
				},
			},
		},
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(getParameterAttributesResponseTestData), string(b))
}

func TestEncodeAddObjectResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.AddObjectResponse = &AddObjectResponseEncoder{
		InstanceNumber: 123,
		Status:         1,
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(addObjectResponseTestData), string(b))
}

func TestEncodeDeleteObjectResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.DeleteObjectResponse = &DeleteObjectResponseEncoder{
		Status: 1,
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(deleteObjectResponseTestData), string(b))
}

func TestEncodeRebootResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.RebootResponse = &RebootResponseEncoder{}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(rebootResponseTestData), string(b))
}

func TestEncodeDownloadResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.DownloadResponse = &DownloadResponseEncoder{
		Status:       1,
		StartTime:    "2024-06-10T23:04:00Z",
		CompleteTime: "2024-06-10T23:05:00Z",
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(downloadResponseTestData), string(b))
}

func TestEncodeFactoryResetResponse(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.FactoryResetResponse = &FactoryResetResponseEncoder{}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(factoryResetResponseTestData), string(b))
}

func TestEncodeTransferCompleteSuccessRequest(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.TransferCompleteRequest = &TransferCompleteRequestEncoder{
		CommandKey:   "upgrade",
		Fault:        NoFaultStruct{},
		StartTime:    "2024-06-10T23:04:00Z",
		CompleteTime: "2024-06-10T23:05:00Z",
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(transferCompleteSuccessRequestTestData), string(b))
}

func TestEncodeTransferCompleteFaultRequest(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.TransferCompleteRequest = &TransferCompleteRequestEncoder{
		CommandKey: "upgrade",
		Fault: FaultStruct{
			FaultCode:   FaultDownloadFailureContactFileServer,
			FaultString: "Server Not Found",
		},
		StartTime:    "2024-06-10T23:04:00Z",
		CompleteTime: "2024-06-10T23:05:00Z",
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(transferCompleteFaultRequestTestData), string(b))
}

func TestEncodeAutonomousTransferCompleteRequest(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.AutonomousTransferCompleteRequest = &AutonomousTransferCompleteRequestEncoder{
		AnnounceURL:    "https://acme-networks.com/firmware/downloads",
		TransferURL:    "https://acme-networks.com/firmware/downloads/firmware.bin",
		IsDownload:     true,
		FileType:       FileTypeFirmwareUpgradeImage,
		FileSize:       184258350,
		TargetFileName: "firmware.bin",
		Fault:          NoFaultStruct{},
		StartTime:      "2024-06-10T23:04:00Z",
		CompleteTime:   "2024-06-10T23:05:00Z",
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(autonomousTransferCompleteRequestTestData), string(b))
}

func TestEncodeFault(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.Fault = NewFaultResponse(FaultMethodNotSupported, "Upload method not supported")

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(faultResponseTestData), string(b))
}

func TestEncodeFaultSetParameterValues(t *testing.T) {
	env := NewEnvelope("123")
	env.Body.Fault = NewFaultResponse(FaultInvalidArguments, FaultInvalidArguments.String())
	env.Body.Fault.Detail.Fault.SetParameterValuesFault = []SetParameterValuesFault{
		{
			ParameterName: "InternetGatewayDevice.Time.LocalTimeZone",
			FaultCode:     FaultInvalidParameterValue,
			FaultString:   "Not a valid time zone value",
		},
		{
			ParameterName: "InternetGatewayDevice.Time.LocalTimeZoneName",
			FaultCode:     FaultInvalidParameterValue,
			FaultString:   "String too long",
		},
	}

	b, err := env.EncodePretty()
	require.NoError(t, err)
	assert.Equal(t, string(faultSetParameterValuesResponseTestData), string(b))
}
