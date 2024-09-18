package reflector

import "reflect"

func Deref(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	return value
}
