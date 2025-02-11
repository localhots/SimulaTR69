package datamodel

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDM = `Parameter,Object,Writable,Value,Type
Device,true,true,,
Device.DeviceInfo,true,true,,
Device.DeviceInfo.Description,false,true,Residential Gateway,xsd:string
Device.DeviceInfo.HardwareVersion,false,true,1.0,xsd:string
Device.DeviceInfo.Manufacturer,false,true,ACME Networks,xsd:string
Device.DeviceInfo.ManufacturerOUI,false,true,DECADE,xsd:string
Device.DeviceInfo.ModelName,false,true,G3000E,xsd:string
`

func TestLoadDataModel(t *testing.T) {
	params, err := LoadDataModel(strings.NewReader(testDM))
	require.NoError(t, err)

	require.Len(t, params, 7)
	assert.Equal(t, "Residential Gateway", params["Device.DeviceInfo.Description"].Value)
	assert.Equal(t, "1.0", params["Device.DeviceInfo.HardwareVersion"].Value)
	assert.Equal(t, "ACME Networks", params["Device.DeviceInfo.Manufacturer"].Value)
	assert.Equal(t, "DECADE", params["Device.DeviceInfo.ManufacturerOUI"].Value)
	assert.Equal(t, "G3000E", params["Device.DeviceInfo.ModelName"].Value)
}

func TestLoadingGenerators(t *testing.T) {
	dmsrc := `Parameter,Object,Writable,Value,Type
Device.Foo,false,false,"randomWalk(startValue=50, minValue=0, maxValue=100, step=0) as xsd:int",sim:generator
`
	params, err := LoadDataModel(strings.NewReader(dmsrc))
	require.NoError(t, err)
	dm := New(newState().WithDefaults(params))

	t.Run("GetValue", func(t *testing.T) {
		p, ok := dm.GetValue("Device.Foo")
		assert.True(t, ok)
		assert.Equal(t, "50", p.Encode().Value.Value)
		assert.Equal(t, "xsd:int", p.Encode().Value.Type)
	})
	t.Run("GetValues", func(t *testing.T) {
		pp, ok := dm.GetValues("Device.Foo")
		assert.True(t, ok)
		require.Len(t, pp, 1)
		assert.Equal(t, "50", pp[0].Encode().Value.Value)
		assert.Equal(t, "xsd:int", pp[0].Encode().Value.Type)
	})
	t.Run("GetAll", func(t *testing.T) {
		pp, ok := dm.GetAll("Device.")
		assert.True(t, ok)
		require.Len(t, pp, 1)
		assert.Equal(t, "50", pp[0].Encode().Value.Value)
		assert.Equal(t, "xsd:int", pp[0].Encode().Value.Type)
	})
}
