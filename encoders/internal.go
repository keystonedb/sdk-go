package encoders

import "reflect"

func deRef(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Pointer {
		return value.Elem()
	}
	return value
}
