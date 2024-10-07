package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
