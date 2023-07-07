package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStringWithOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "HIJ", val)
}
func TestReadStringWithNoOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "DEF", val)
}
func TestReadStringOverrideOnly(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "CDE", "")
	assert.Equal(t, "HIJ", val)
}

func TestReadStringMissWithDefault(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "DEF", "HE")
	assert.Equal(t, "HE", val)
}
func TestReadStringEmptyOverride(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "",
	}, "DEF", "")
	assert.Equal(t, "", val)
}

func TestFormatObjectEmpty(t *testing.T) {
	obj := new(interface{})
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "null", string(val))
}
func TestFormatObjectEmptyDict(t *testing.T) {
	obj := map[string]interface{}{}
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(val))
}
func TestFormatObjectDictJson(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"bar\"", string(val))
}
func TestFormatObjectDictYaml(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar\n", string(val))
}
