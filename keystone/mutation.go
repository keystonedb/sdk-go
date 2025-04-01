package keystone

import (
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

func observeMutation(v interface{}, m *proto.MutateResponse) {
	if v == nil {
		return
	}

	val := reflector.Deref(reflect.ValueOf(v))
	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous {
			continue
		}
		currentVal := val.FieldByIndex(field.Index)

		if reflect.PointerTo(currentVal.Type()).Implements(mutationObserverType) {
			vp := reflect.New(currentVal.Type())
			vp.Elem().Set(currentVal)
			x := vp.Interface()
			if mo, ok := x.(MutationObserver); ok {
				mo.ObserveMutation(m)
				currentVal.Set(reflector.Deref(vp))
				continue
			}
		}

		if currentVal.Type().Implements(mutationObserverType) {
			currentVal.Interface().(MutationObserver).ObserveMutation(m)
		} else if field.Type.Kind() == reflect.Struct && !currentVal.IsZero() && currentVal.CanInterface() {
			observeMutation(currentVal.Interface(), m)
		}
	}
}
