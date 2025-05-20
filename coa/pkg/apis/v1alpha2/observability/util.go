/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package observability

import (
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
)

func convertMapToAttributes(m map[string]any) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(m))
	i := 0
	for k, v := range m {
		attrs[i] = convertKey(k, v)
		i++
	}

	return attrs
}

func convertKey(k string, v any) attribute.KeyValue {
	r := reflect.TypeOf(v)
	if r != nil && r.Kind() == reflect.Func {
		res := reflect.ValueOf(v).Call(nil)[0]
		return convertKey(k, res.Interface())
	}

	switch val := v.(type) {
	case string:
		return attribute.String(k, val)
	case int:
		return attribute.Int(k, val)
	case int8:
		return attribute.Int(k, int(val))
	case int16:
		return attribute.Int(k, int(val))
	case int32:
		return attribute.Int(k, int(val))
	case int64:
		return attribute.Int64(k, val)
	case float32:
		return attribute.Float64(k, float64(val))
	case float64:
		return attribute.Float64(k, val)
	case bool:
		return attribute.Bool(k, val)
	case []string:
		return attribute.StringSlice(k, val)
	case []bool:
		return attribute.BoolSlice(k, val)
	case []float64:
		return attribute.Float64Slice(k, val)
	case []int64:
		return attribute.Int64Slice(k, val)
	case []int:
		return attribute.IntSlice(k, val)
	default:
		return attribute.String(k, fmt.Sprintf("%v", val))
	}
}
