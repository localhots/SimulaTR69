package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTypeDef(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		td, err := parseTypeDef("string")
		require.NoError(t, err)
		require.NotNil(t, td)
		assert.Equal(t, "string", td.name)
		assert.Nil(t, td.min)
		assert.Nil(t, td.max)
	})
	t.Run("xsd:string", func(t *testing.T) {
		td, err := parseTypeDef("xsd:string")
		require.NoError(t, err)
		require.NotNil(t, td)
		assert.Equal(t, "string", td.name)
		assert.Nil(t, td.min)
		assert.Nil(t, td.max)
	})
	t.Run("int(10)", func(t *testing.T) {
		td, err := parseTypeDef("int(10)")
		require.NoError(t, err)
		require.NotNil(t, td)
		assert.Equal(t, "int", td.name)
		assert.Nil(t, td.min)
		require.NotNil(t, td.max)
		assert.Equal(t, 10, *td.max)
	})
	t.Run("int(10:50)", func(t *testing.T) {
		td, err := parseTypeDef("int(10:50)")
		require.NoError(t, err)
		require.NotNil(t, td)
		assert.Equal(t, "int", td.name)
		require.NotNil(t, td.min)
		assert.Equal(t, 10, *td.min)
		require.NotNil(t, td.max)
		assert.Equal(t, 50, *td.max)
	})
	t.Run("whitespaced", func(t *testing.T) {
		td, err := parseTypeDef("   string 	 ")
		require.NoError(t, err)
		require.NotNil(t, td)
		assert.Equal(t, "string", td.name)
		assert.Nil(t, td.min)
		assert.Nil(t, td.max)
	})
	t.Run("invalid+", func(t *testing.T) {
		td, err := parseTypeDef("invalid+")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid type definition")
		require.Nil(t, td)
	})
	t.Run("invalid(:50)", func(t *testing.T) {
		td, err := parseTypeDef("invalid(:50)")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid type definition")
		require.Nil(t, td)
	})
	t.Run("invalid(x:50)", func(t *testing.T) {
		td, err := parseTypeDef("invalid(x:50)")
		require.Error(t, err)
		assert.EqualError(t, err, "invalid type definition")
		require.Nil(t, td)
	})
	t.Run("int(0:9999999999999999999999)", func(t *testing.T) {
		td, err := parseTypeDef("int(0:9999999999999999999999)")
		require.Error(t, err)
		assert.EqualError(t, err, `parse type max: strconv.Atoi: parsing "9999999999999999999999": value out of range`)
		require.Nil(t, td)
	})
}
