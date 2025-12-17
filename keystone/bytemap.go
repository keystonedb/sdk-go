package keystone

import (
	"encoding/json"
	"reflect"

	"github.com/keystonedb/sdk-go/keystone/reflector"
)

func ToByteMap(from any) map[string][]byte {
	final := make(map[string][]byte)

	val := reflector.Deref(reflect.ValueOf(from))
	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}

		if jsnVal, err := json.Marshal(val.FieldByIndex(field.Index).Interface()); err == nil {
			final[field.Name] = jsnVal
		}
	}

	return final
}

func FromByteMap(from map[string][]byte, to any) {
	val := reflector.Deref(reflect.ValueOf(to))
	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}

		if jsnVal, ok := from[field.Name]; ok {
			_ = json.Unmarshal(jsnVal, val.FieldByIndex(field.Index).Addr().Interface())
		}
	}
}
