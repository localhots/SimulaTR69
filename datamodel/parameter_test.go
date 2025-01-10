package datamodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeParameters(t *testing.T) {
	params := map[string]Parameter{
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
	}
	NormalizeParameters(params)

	param := params["Device.DeviceInfo.DeviceCategory"]
	assert.Equal(t, "xsd:string", param.Type)
	assert.Equal(t, "", param.GetValue())

	param = params["Device.DeviceInfo.DeviceImageNumberOfEntries"]
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "0", param.GetValue())

	param = params["Device.DeviceInfo.FirstUseDate"]
	assert.Equal(t, "xsd:dateTime", param.Type)
	assert.Equal(t, "2023-11-22T04:30:27Z", param.GetValue())

	param = params["Device.DeviceInfo.MemoryStatus"]
	assert.Equal(t, "Device.DeviceInfo.MemoryStatus", param.Path)
	assert.Equal(t, "object", param.Type)
	assert.Equal(t, "", param.GetValue())

	param = params["Device.DeviceInfo.MemoryStatus.Free"]
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "163636", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Enable"]
	assert.Equal(t, "xsd:boolean", param.Type)
	assert.Equal(t, "true", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.LowAlarmValue"]
	assert.Equal(t, "xsd:int", param.Type)
	assert.Equal(t, "-274", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.MaxValue"]
	assert.Equal(t, "xsd:int", param.Type)
	assert.Equal(t, "-274", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.PollingInterval"]
	assert.Equal(t, "xsd:unsignedInt", param.Type)
	assert.Equal(t, "0", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Reset"]
	assert.Equal(t, "xsd:boolean", param.Type)
	assert.Equal(t, "false", param.GetValue())

	param = params["Device.DeviceInfo.TemperatureStatus.TemperatureSensor.1.Status"]
	assert.Equal(t, "xsd:string", param.Type)
	assert.Equal(t, "Enabled", param.GetValue())
}

func TestNormalizeBool(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		val := normalizeBool("foo", "")
		assert.Equal(t, "false", val)
	})
	t.Run("0", func(t *testing.T) {
		val := normalizeBool("foo", "0")
		assert.Equal(t, "false", val)
	})
	t.Run("1", func(t *testing.T) {
		val := normalizeBool("foo", "1")
		assert.Equal(t, "true", val)
	})
	t.Run("false", func(t *testing.T) {
		val := normalizeBool("foo", "false")
		assert.Equal(t, "false", val)
	})
	t.Run("true", func(t *testing.T) {
		val := normalizeBool("foo", "true")
		assert.Equal(t, "true", val)
	})
	t.Run("no", func(t *testing.T) {
		val := normalizeBool("foo", "no")
		assert.Equal(t, "false", val)
	})
	t.Run("off", func(t *testing.T) {
		val := normalizeBool("foo", "off")
		assert.Equal(t, "false", val)
	})
	t.Run("disabled", func(t *testing.T) {
		val := normalizeBool("foo", "disabled")
		assert.Equal(t, "false", val)
	})
	t.Run("yes", func(t *testing.T) {
		val := normalizeBool("foo", "yes")
		assert.Equal(t, "true", val)
	})
	t.Run("on", func(t *testing.T) {
		val := normalizeBool("foo", "on")
		assert.Equal(t, "true", val)
	})
	t.Run("enabled", func(t *testing.T) {
		val := normalizeBool("foo", "enabled")
		assert.Equal(t, "true", val)
	})
	t.Run("invalid", func(t *testing.T) {
		val := normalizeBool("foo", "invalid")
		assert.Equal(t, "false", val)
	})
}

func TestNormalizeInt(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		val := normalizeInt("foo", "")
		assert.Equal(t, "0", val)
	})
	t.Run("123", func(t *testing.T) {
		val := normalizeInt("foo", "123")
		assert.Equal(t, "123", val)
	})
	t.Run("invalid", func(t *testing.T) {
		val := normalizeInt("foo", "invalid")
		assert.Equal(t, "0", val)
	})
}

func TestNormalizeUint(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		val := normalizeUint("foo", "")
		assert.Equal(t, "0", val)
	})
	t.Run("123", func(t *testing.T) {
		val := normalizeUint("foo", "123")
		assert.Equal(t, "123", val)
	})
	t.Run("negative", func(t *testing.T) {
		val := normalizeUint("foo", "-123")
		assert.Equal(t, "0", val)
	})
	t.Run("invalid", func(t *testing.T) {
		val := normalizeUint("foo", "invalid")
		assert.Equal(t, "0", val)
	})
}

func TestGenerator(t *testing.T) {
	fd := strings.NewReader(strings.TrimSpace(`
Parameter,Object,Writable,Value,Type
Device.DeviceInfo.ProcessStatus.CPUUsage,false,true,"perlinNoise(offset=50, alpha=2, beta=2, seed=42, scale=40) as xsd:int",sim:generator
`))
	params, err := LoadDataModel(fd)
	assert.NoError(t, err)
	param := params["Device.DeviceInfo.ProcessStatus.CPUUsage"]
	val := param.GetValue()
	assert.Equal(t, "50", val)
	assert.Equal(t, "xsd:int", param.Type)
}
