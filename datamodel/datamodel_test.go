package datamodel

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/localhots/SimulaTR69/rpc"
)

func TestNew(t *testing.T) {
	dm := New(newState())
	assert.NotNil(t, dm)
	assert.Empty(t, dm.values.Changes)
	assert.Empty(t, dm.values.Deleted)
	assert.False(t, dm.values.Bootstrapped)
}

func TestReset(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	dm.Reset()
	assert.Equal(t, tr181, dm.version)
	assert.Empty(t, dm.commandKey)
	assert.Empty(t, dm.events)
	assert.Zero(t, dm.retryAttempts)
	assert.True(t, dm.downUntil.IsZero())
}

func TestVersionKnown(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	assert.Equal(t, "TR-181", dm.Version())
}

func TestVersionUnknown(t *testing.T) {
	dm := New(newState())
	assert.Equal(t, "Unknown", dm.Version())
}

func TestGetAll(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.Description",
			Value: "1.0",
		},
		"Device.Ethernet.Interface.1.DuplexMode": {
			Path:  "Device.Ethernet.Interface.1.DuplexMode",
			Value: "Auto",
		},
	}))
	params, ok := dm.GetAll("Device.DeviceInfo.")
	assert.True(t, ok)
	assert.Len(t, params, 2)
}

func TestGetValue(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	param, ok := dm.GetValue("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, "Residential Gateway", param.Value)
}

func TestGetValues(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.Description",
			Value: "1.0",
		},
		"Device.Ethernet.Interface.1.DuplexMode": {
			Path:  "Device.Ethernet.Interface.1.DuplexMode",
			Value: "Auto",
		},
	}))
	params, ok := dm.GetValues(
		"Device.DeviceInfo.Description",
		"Device.DeviceInfo.HardwareVersion",
	)
	fmt.Println(ok, params)
	assert.True(t, ok)
	assert.Len(t, params, 2)
}

func TestSetValue(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	dm.SetValue("Device.DeviceInfo.Description", "New Description")
	param, ok := dm.GetValue("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, "New Description", param.Value)
}

func TestSetValuesEmpty(t *testing.T) {
	dm := New(newState())
	dm.SetValues([]Parameter{})
	assert.Empty(t, dm.values.Changes)
}

func TestSetValues(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.HardwareVersion",
			Value: "1.0",
		},
	}))
	dm.SetValues([]Parameter{
		{
			Path:  "Device.DeviceInfo.Description",
			Value: "New Description",
		},
		{
			Path:  "Device.DeviceInfo.HardwareVersion",
			Value: "2.0",
		},
	})
	param1, ok1 := dm.GetValue("Device.DeviceInfo.Description")
	param2, ok2 := dm.GetValue("Device.DeviceInfo.HardwareVersion")
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Equal(t, "New Description", param1.Value)
	assert.Equal(t, "2.0", param2.Value)
}

func TestCanSetValueNonWritable(t *testing.T) {
	state := newState()
	param := Parameter{Path: "Device.DeviceInfo.Description", Writable: false}
	dm := New(state.WithDefaults(map[string]Parameter{
		param.Path: param,
	}))
	fault := dm.CanSetValue(param)
	require.NotNil(t, fault)
	assert.Equal(t, rpc.FaultNonWritableParameter, *fault)
}

func TestCanSetValueInvalidParent(t *testing.T) {
	state := newState()
	param := Parameter{Path: "Device.NonExistent.Path"}
	dm := New(state)
	fault := dm.CanSetValue(param)
	require.NotNil(t, fault)
	assert.Equal(t, rpc.FaultInvalidParameterName, *fault)
}

func TestAddObject(t *testing.T) {
	dm := New(newState())
	dm.SetValue("Device.DeviceInfo.Description", "Residential Gateway")
	_, err := dm.AddObject("Device.DeviceInfo.Description")
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("parent is not an object"), err)
}

func TestAddObjectNonExistentParent(t *testing.T) {
	dm := New(newState())
	_, err := dm.AddObject("Device.NonExistent.Parent")
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("parent object doesn't exist"), err)
}

func TestAddObjectParentNotObject(t *testing.T) {
	dm := New(newState())
	dm.SetValue("Device.DeviceInfo", "Some Value")
	_, err := dm.AddObject("Device.DeviceInfo")
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("parent is not an object"), err)
}

func TestDeleteObject(t *testing.T) {
	dm := New(newState())
	dm.SetValue("Device.DeviceInfo.Description", "Residential Gateway")
	dm.DeleteObject("Device.DeviceInfo.Description")
	_, ok := dm.GetValue("Device.DeviceInfo.Description")
	assert.False(t, ok)
}

func TestDeleteObjectNonExistent(t *testing.T) {
	dm := New(newState())
	dm.DeleteObject("Device.NonExistent.Object")
	_, ok := dm.GetValue("Device.NonExistent.Object")
	assert.False(t, ok)
}

func TestSetParameterAttribute(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	dm.SetParameterAttribute("Device.DeviceInfo.Description", 1, true, []string{"read"}, true)
	param, ok := dm.GetValue("Device.DeviceInfo.Description")
	assert.True(t, ok)
	assert.Equal(t, rpc.AttributeNotification(1), param.Notification)
	assert.Equal(t, []string{"read"}, param.ACL)
}

func TestSetParameterAttributeNonExistent(t *testing.T) {
	dm := New(newState())
	dm.SetParameterAttribute("Device.NonExistent.Path", 1, true, []string{"read"}, true)
	param, ok := dm.GetValue("Device.NonExistent.Path")
	assert.False(t, ok)
	assert.Empty(t, param)
}

func TestPendingEvents(t *testing.T) {
	dm := New(newState())
	dm.AddEvent(rpc.EventPeriodic)
	events := dm.PendingEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, rpc.EventPeriodic, events[0])
}

func TestAddEvent(t *testing.T) {
	dm := New(newState())
	dm.AddEvent(rpc.EventPeriodic)
	assert.Contains(t, dm.events, rpc.EventPeriodic)
}

func TestClearEvents(t *testing.T) {
	dm := New(newState())
	dm.AddEvent(rpc.EventPeriodic)
	dm.ClearEvents()
	assert.Empty(t, dm.events)
}

func TestIsBootstrapped(t *testing.T) {
	state := &State{Bootstrapped: true}
	dm := New(state)
	assert.True(t, dm.IsBootstrapped())
}

func TestSetBootstrapped(t *testing.T) {
	dm := New(newState())
	dm.SetBootstrapped(true)
	assert.True(t, dm.IsBootstrapped())
}

func TestRetryAttempts(t *testing.T) {
	dm := New(newState())
	assert.Zero(t, dm.RetryAttempts())
	dm.IncrRetryAttempts()
	assert.Equal(t, uint32(1), dm.RetryAttempts())
	dm.ResetRetryAttempts()
	assert.Zero(t, dm.RetryAttempts())
}

func TestCommandKey(t *testing.T) {
	dm := New(newState())
	dm.SetCommandKey("TestKey")
	assert.Equal(t, "TestKey", dm.CommandKey())
}

func TestDownUntil(t *testing.T) {
	dm := New(newState())
	now := time.Now()
	dm.SetDownUntil(now)
	assert.Equal(t, now, dm.DownUntil())
}

func TestSetDownUntil(t *testing.T) {
	dm := New(newState())
	future := time.Now().Add(time.Hour)
	dm.SetDownUntil(future)
	assert.Equal(t, future, dm.DownUntil())
}

func TestParameterNamesEmptyPath(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	params := dm.ParameterNames("", true)
	assert.Len(t, params, 0)
}

func TestParameterNamesNoMatch(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
	}))
	params := dm.ParameterNames("Device.Ethernet", true)
	assert.Len(t, params, 0)
}

func TestParameterNamesNextLevel(t *testing.T) {
	state := newState()
	dm := New(state.WithDefaults(map[string]Parameter{
		"Device.DeviceInfo.Description": {
			Path:  "Device.DeviceInfo.Description",
			Value: "Residential Gateway",
		},
		"Device.DeviceInfo.HardwareVersion": {
			Path:  "Device.DeviceInfo.HardwareVersion",
			Value: "1.0",
		},
	}))
	params := dm.ParameterNames("Device.DeviceInfo", true)
	assert.Len(t, params, 2)
}