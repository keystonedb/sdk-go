package reflector

import "reflect"

// Deref safely unwraps pointers and interfaces.
// If a nil pointer or nil interface is encountered, it returns an invalid reflect.Value.
func Deref(value reflect.Value) reflect.Value {
	for {
		if !value.IsValid() {
			return value
		}
		k := value.Kind()
		if k == reflect.Pointer || k == reflect.Interface {
			if value.IsNil() {
				return reflect.Value{}
			}
			value = value.Elem()
			continue
		}
		return value
	}
}
