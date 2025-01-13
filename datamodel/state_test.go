package datamodel

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	state := newState()
	assert.NotNil(t, state)
	assert.False(t, state.Bootstrapped)
	assert.Empty(t, state.Changes)
	assert.Empty(t, state.Deleted)
	assert.Empty(t, state.defaults)
}

func TestStateWithDefaults(t *testing.T) {
	state := newState()
	defaults := map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}
	state = state.WithDefaults(defaults)
	assert.Equal(t, defaults, state.defaults)
}

func TestStateGet(t *testing.T) {
	state := newState()
	defaults := map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}
	state = state.WithDefaults(defaults)

	p, ok := state.get("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, "Residential Gateway", p.Value)

	state.save(Parameter{
		Path:  "Device.DeviceInfo.HardwareVersion",
		Value: "1.0",
	})
	p, ok = state.get("Device.DeviceInfo.HardwareVersion")
	assert.True(t, ok)
	assert.Equal(t, "1.0", p.Value)

	state.delete("Device.DeviceInfo.Description")
	p, ok = state.get("Device.DeviceInfo.Description")
	assert.False(t, ok)
}

func TestStateGetNonExistent(t *testing.T) {
	state := newState()
	_, ok := state.get("nonexistent")
	assert.False(t, ok)
}

func TestStateForEach(t *testing.T) {
	state := newState()
	defaults := map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.HardwareVersion",
			Value: "1.0",
		},
	}
	state = state.WithDefaults(defaults)
	state.save(Parameter{
		Path:  "Device.DeviceInfo.UpTime",
		Value: "300",
	})

	var params []Parameter
	state.forEach(func(p Parameter) bool {
		params = append(params, p)
		return true
	})

	assert.Len(t, params, 3)
}

func TestStateForEachStopEarly(t *testing.T) {
	state := newState()
	defaults := map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.HardwareVersion",
			Value: "1.0",
		},
	}
	state = state.WithDefaults(defaults)
	state.save(Parameter{
		Path:  "Device.DeviceInfo.UpTime",
		Value: "300",
	})

	var params []Parameter
	state.forEach(func(p Parameter) bool {
		params = append(params, p)
		return len(params) < 2
	})

	assert.Len(t, params, 2)
}

func TestStateSave(t *testing.T) {
	state := newState()
	param := Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	}
	state.save(param)

	p, ok := state.get("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, "Residential Gateway", p.Value)
}

func TestStateSaveOverwrite(t *testing.T) {
	state := newState()
	param := Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	}
	state.save(param)
	param.Value = "new_value"
	state.save(param)

	p, ok := state.get("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, "new_value", p.Value)
}

func TestStateDelete(t *testing.T) {
	state := newState()
	state.save(Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	})
	state.delete("Device.DeviceInfo.Description")

	_, ok := state.get("Device.DeviceInfo.Description")
	assert.False(t, ok)
	assert.Contains(t, slices.Collect(maps.Keys(state.Deleted)), "Device.DeviceInfo.Description")
}

func TestStateDeleteNonExistent(t *testing.T) {
	state := newState()
	state.delete("nonexistent")

	_, ok := state.get("nonexistent")
	assert.False(t, ok)
	assert.NotContains(t, slices.Collect(maps.Keys(state.Deleted)), "nonexistent")
}

func TestStateDeletePrefix(t *testing.T) {
	state := newState()
	state.save(Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	})
	state.save(Parameter{
		Path:  "Device.DeviceInfo.HardwareVersion",
		Value: "1.0",
	})
	state.deletePrefix("Device.DeviceInfo.")

	_, ok := state.get("Device.DeviceInfo.Description")
	assert.False(t, ok)
	_, ok = state.get("Device.DeviceInfo.HardwareVersion")
	assert.False(t, ok)
	assert.Contains(t, state.Deleted, "Device.DeviceInfo.Description")
	assert.Contains(t, state.Deleted, "Device.DeviceInfo.HardwareVersion")
}

func TestStateDeletePrefixPartialMatch(t *testing.T) {
	state := newState()
	state.save(Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	})
	state.save(Parameter{
		Path:  "Device.Ethernet.Interface.1.DuplexMode",
		Value: "Auto",
	})
	state.deletePrefix("Device.DeviceInfo.")

	_, ok := state.get("Device.DeviceInfo.Description")
	assert.False(t, ok)
	_, ok = state.get("Device.Ethernet.Interface.1.DuplexMode")
	assert.True(t, ok)
	assert.Contains(t, state.Deleted, "Device.DeviceInfo.Description")
	assert.NotContains(t, state.Deleted, "Device.Ethernet.Interface.1.DuplexMode")
}

func TestStateReset(t *testing.T) {
	state := newState()
	state.save(Parameter{
		Path:  "Device.DeviceInfo.Description",
		Value: "Residential Gateway",
	})
	state.reset()

	assert.False(t, state.Bootstrapped)
	assert.Empty(t, state.Changes)
	assert.Empty(t, state.Deleted)
}
