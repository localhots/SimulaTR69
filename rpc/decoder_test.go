package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeGetRPCMethodsRequest(t *testing.T) {
	env, err := Decode(getRPCMethodsRequestTestData)
	require.NoError(t, err)
	assert.NotNil(t, env.Body.GetRPCMethods)
}

func TestDecodeSetParameterValuesRequest(t *testing.T) {
	env, err := Decode(setParameterValuesRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.SetParameterValues)

	assert.Equal(t, "n/a", env.Body.SetParameterValues.ParameterKey)
	pl := env.Body.SetParameterValues.ParameterList
	require.Len(t, pl.ParameterValues, 2)
	require.Equal(t, ArrayType("cwmp:ParameterValueStruct", len(pl.ParameterValues)), pl.ArrayType)

	v1 := pl.ParameterValues[0]
	assert.Equal(t, "Device.ManagementServer.ConnectionRequestUsername", v1.Name)
	assert.Equal(t, TypeXSDString, v1.Value.Type)
	assert.Equal(t, "G3000E-9799109101", v1.Value.Value)
	v2 := pl.ParameterValues[1]
	assert.Equal(t, "Device.ManagementServer.ConnectionRequestPassword", v2.Name)
	assert.Equal(t, TypeXSDString, v2.Value.Type)
	assert.Equal(t, "secret", v2.Value.Value)
}

func TestDecodeGetParameterValuesRequest(t *testing.T) {
	env, err := Decode(getParameterValuesRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.GetParameterValues)

	pn := env.Body.GetParameterValues.ParameterNames
	require.Len(t, pn.Names, 1)
	require.Equal(t, ArrayType(TypeXSDString, len(pn.Names)), pn.ArrayType)
	assert.Equal(t, "Device.DeviceSummary.", pn.Names[0])
}

func TestDecodeGetParameterNamesRequest(t *testing.T) {
	env, err := Decode(getParameterNamesRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.GetParameterNames)

	pn := env.Body.GetParameterNames
	assert.Equal(t, "Device.", pn.ParameterPath)
	assert.Equal(t, false, pn.NextLevel)
}

func TestDecodeSetParameterAttributesRequest(t *testing.T) {
	env, err := Decode(setParameterAttributesRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.SetParameterAttributes)
	pl := env.Body.SetParameterAttributes.ParameterList
	assert.Equal(t, ArrayType("cwmp:SetParameterAttributesStruct", 1), pl.ArrayType)
	require.Len(t, pl.ParameterAttributes, 1)

	pa := pl.ParameterAttributes[0]
	assert.Equal(t, "Device.DeviceSummary", pa.Name)
	assert.Equal(t, true, pa.NotificationChange)
	assert.Equal(t, AttributeNotificationPassive, pa.Notification)
	assert.Equal(t, true, pa.AccessListChange)
	assert.Equal(t, ArrayType(TypeXSDString, 1), pa.AccessList.ArrayType)
	require.Len(t, pa.AccessList.Values, 1)
	assert.Equal(t, "Subscriber", pa.AccessList.Values[0])
}

func TestDecodeGetParameterAttributesRequest(t *testing.T) {
	env, err := Decode(getParameterAttributesRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.GetParameterAttributes)

	pn := env.Body.GetParameterAttributes.ParameterNames
	assert.Equal(t, ArrayType(TypeXSDString, 1), pn.ArrayType)
	require.Len(t, pn.Names, 1)
	assert.Equal(t, "Device.DeviceInfo.VendorConfigFile.1.Version", pn.Names[0])
}

func TestDecodeAddObjectRequest(t *testing.T) {
	env, err := Decode(addObjectTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.AddObject)

	assert.Equal(t, "Device.NAT.PortMapping.", env.Body.AddObject.ObjectName)
	assert.Equal(t, "123", env.Body.AddObject.ParameterKey)
}

func TestDecodeDeleteObjectRequest(t *testing.T) {
	env, err := Decode(deleteObjectRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.DeleteObject)

	assert.Equal(t, "Device.NAT.PortMapping.", env.Body.DeleteObject.ObjectName)
	assert.Equal(t, "123", env.Body.DeleteObject.ParameterKey)
}

func TestDecodeRebootRequest(t *testing.T) {
	env, err := Decode(rebootRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.Reboot)

	assert.Equal(t, "example", env.Body.Reboot.CommandKey)
}

func TestDecodeDownloadRequest(t *testing.T) {
	env, err := Decode(downloadRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.Download)

	dl := env.Body.Download
	assert.Equal(t, "FirmwareUpgrade", dl.CommandKey)
	assert.Equal(t, FileTypeFirmwareUpgradeImage, dl.FileType)
	assert.Equal(t, "https://acme-networks.com/firmware/downloads/firmware.bin", dl.URL)
	assert.Equal(t, "cpe", dl.Username)
	assert.Equal(t, "secret", dl.Password)
	assert.Equal(t, 184258350, dl.FileSize)
	assert.Equal(t, "firmware.bin", dl.TargetFileName)
	assert.Equal(t, 0, dl.DelaySeconds)
	assert.Equal(t, "http://success", dl.SuccessURL)
	assert.Equal(t, "http://failure", dl.FailureURL)
}

func TestDecodeFactoryResetRequest(t *testing.T) {
	env, err := Decode(factoryResetRequestTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.FactoryReset)
}

func TestDecodeInformResponse(t *testing.T) {
	env, err := Decode(informResponseTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.InformResponse)
	assert.Equal(t, MaxEnvelopes, env.Body.InformResponse.MaxEnvelopes)
}

func TestDecodeTransferCompleteResponse(t *testing.T) {
	env, err := Decode(transferCompleteResponseTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.TransferCompleteResponse)
}

func TestDecodeAutonomousTransferCompleteResponse(t *testing.T) {
	env, err := Decode(autonomousTransferCompleteResponseTestData)
	require.NoError(t, err)
	require.NotNil(t, env.Body.AutonomousTransferCompleteResponse)
}
