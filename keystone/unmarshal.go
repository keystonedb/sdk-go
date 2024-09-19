package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

func Unmarshal(data map[Property]*proto.Value, v interface{}) error {
	if v == nil || data == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("you must pass a pointer to Unmarshal")
	}
	val = reflector.Deref(val)

	prefixed := make(map[string]map[Property]*proto.Value)
	for k, reVal := range data {
		if k.prefix != "" {
			prefix := k.prefix
			if _, has := prefixed[prefix]; !has {
				prefixed[prefix] = make(map[Property]*proto.Value)
			}
			k.prefix = ""
			prefixed[prefix][k] = reVal
		}
	}

	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}

		currentProp := NewProperty(field.Name)
		toHydrate, hasVal := data[currentProp]
		subData, hasSubData := prefixed[currentProp.Name()]
		if !hasVal && !hasSubData {
			continue
		}

		currentVal := val.FieldByIndex(field.Index)
		ref := GetReflector(field.Type, currentVal)
		if ref != nil {
			if err := ref.SetValue(toHydrate, currentVal); err != nil {
				return err
			}
		} else if field.Type.Kind() == reflect.Struct && len(subData) > 0 {
			if err := Unmarshal(subData, currentVal.Addr().Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}
