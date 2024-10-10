package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeParameters(t *testing.T) {
	dm := DataModel{values: newState(map[string]Parameter{
		"Device.DeviceInfo.DeviceCategory": {
			Path:  "Device.DeviceInfo.DeviceCategory",
			Type:  "xsd:string",
			Value: "",
		},
		"Device.DeviceInfo.DeviceImageNumberOfEntries": {
			Path:  "Device.DeviceInfo.DeviceImageNumberOfEntries",
			Type:  "xsd:unsignedInt",
			Value: "",
		},
		"Device.DeviceInfo.FirstUseDate": {
			Path:  "Device.DeviceInfo.FirstUseDate",
			Type:  "dateTime",
			Value: "2023-11-22T04:30:27Z",
		},
		"Device.DeviceInfo.MemoryStatus": {
			Path:   "Device.DeviceInfo.MemoryStatus",
			Object: true,
		},
		"Device.DeviceInfo.MemoryStatus.Free": {
			Path:  "Device.DeviceInfo.MemoryStatus.Free",
			Type:  "unsignedInt",
			Value: "163636",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Enable": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Enable",
			Type:  "xsd:boolean",
			Value: "1",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.LowAlarmValue": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.LowAlarmValue",
			Type:  "xsd:int",
			Value: "-274",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.MaxValue": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.MaxValue",
			Type:  "int",
			Value: "-274",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.PollingInterval": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.PollingInterval",
			Type:  "unsignedInt",
			Value: "-1",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Reset": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Reset",
			Type:  "boolean",
			Value: "",
		},
		"Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Status": {
			Path:  "Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Status",
			Type:  "string",
			Value: "Enabled",
		},
	})}
	dm.NormalizeParameters()

	param, _ := dm.values.get("Device.DeviceInfo.DeviceCategory")
	assert.Equal(t, "xsd:string", param.Type)
	assert.Equal(t, "", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.DeviceImageNumberOfEntries")
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "0", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.FirstUseDate")
	assert.Equal(t, "xsd:dateTime", param.Type)
	assert.Equal(t, "2023-11-22T04:30:27Z", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.MemoryStatus")
	assert.Equal(t, "Device.DeviceInfo.MemoryStatus", param.Path)
	assert.Equal(t, "object", param.Type)
	assert.Equal(t, "", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.MemoryStatus.Free")
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "163636", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Enable")
	assert.Equal(t, "xsd:boolean", param.Type)
	assert.Equal(t, "true", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.LowAlarmValue")
	assert.Equal(t, "xsd:int", param.Type)
	assert.Equal(t, "-274", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.MaxValue")
	assert.Equal(t, "xsd:int", param.Type)
	assert.Equal(t, "-274", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.PollingInterval")
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "0", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Reset")
	assert.Equal(t, "xsd:boolean", param.Type)
	assert.Equal(t, "false", param.Value)

	param, _ = dm.values.get("Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Status")
	assert.Equal(t, "xsd:string", param.Type)
	assert.Equal(t, "Enabled", param.Value)
}
