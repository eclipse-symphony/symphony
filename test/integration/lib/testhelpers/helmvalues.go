package testhelpers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type (
	HelmValues map[string]interface{}
)

func (h HelmValues) String() string {
	flatMap := make(map[string]interface{}, len(h))
	flatten(reflect.ValueOf(h), "", flatMap)

	items := make([]string, 0, len(flatMap))
	for key, value := range flatMap {
		switch v := value.(type) {
		case string:
			value = fmt.Sprintf(`"%s"`, strings.ReplaceAll(v, `"`, `\"`))
		}
		items = append(items, fmt.Sprintf("--set %s=%v", key, value))
	}
	return strings.Join(items, " ")
}

func flatten(inputValue reflect.Value, parentKey string, flatMap map[string]interface{}) {
	if inputValue.Kind() == reflect.Interface || inputValue.Kind() == reflect.Ptr {
		inputValue = inputValue.Elem()
	}
	inputType := inputValue.Type()

	switch inputType.Kind() {
	case reflect.Map:
		for _, key := range inputValue.MapKeys() {
			flatten(inputValue.MapIndex(key), parentKey+"."+key.String(), flatMap)
		}
	case reflect.Slice:
		for i := 0; i < inputValue.Len(); i++ {
			flatten(inputValue.Index(i), parentKey+"["+strconv.Itoa(i)+"]", flatMap)
		}
	default:
		parentKey = strings.TrimPrefix(parentKey, ".")
		inputValue.Type()
		flatMap[parentKey] = inputValue.Interface()
	}
}
