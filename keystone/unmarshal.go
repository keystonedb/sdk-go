package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
	"sort"
)

var ErrMustPassPointer = errors.New("you must pass a pointer")
var ErrMustPointerSlice = errors.New("you must pass a pointer to a []Struct")
var ErrMustMapStringStruct = errors.New("you must pass a map[string]Struct")
var ErrNilMapGiven = errors.New("you must pass a non-nil map")
var ErrEntityIDMissing = errors.New("entity ID missing")
var ErrMustBeMapOrSlice = errors.New("you must unmarshal onto a map or slice")

func Unmarshal(from *proto.EntityResponse, v interface{}) error {
	if v == nil || from == nil {
		return nil
	}
	conv := entityConverter{protoResponse: from}
	data := conv.Properties()

	if e, ok := v.(Entity); ok && from.Entity != nil {
		e.SetKeystoneID(ID(from.Entity.GetEntityId()))
	}

	if e, ok := v.(RelationshipProvider); ok && from.Relationships != nil {
		e.SetRelationships(from.Relationships)
	}

	if len(from.Objects) > 0 {
		if e, ok := v.(ObjectProvider); ok {
			for _, obj := range from.Objects {
				e.addObject(obj)
			}
		}
	}

	if watchable, ok := v.(WatchedEntity); ok && watchable.HasWatcher() {
		watchable.Watcher().AppendKnownValues(data)
	} else if entity, settable := v.(SettableWatchedEntity); settable && !watchable.HasWatcher() {
		if w, err := NewDefaultsWatcher(v); err == nil {
			w.AppendKnownValues(data)
			entity.SetWatcher(w)
		}
	}

	return UnmarshalProperties(data, v)
}

func forTypeFromResponse(elementType reflect.Type, response *proto.EntityResponse) (reflect.Value, error) {
	pointer := elementType.Kind() == reflect.Ptr
	if pointer {
		elementType = elementType.Elem()
	}
	dstEle := reflect.New(elementType)
	ifa := dstEle.Interface()
	if err := Unmarshal(response, ifa); err != nil {
		return reflect.Value{}, err
	}
	val := reflect.ValueOf(ifa)
	if pointer {
		val = reflect.ValueOf(&ifa)
	}
	if val.Kind() == reflect.Ptr {
		val = reflect.ValueOf(val.Elem().Interface())
	}
	return val, nil
}

func New[T any](entity *proto.EntityResponse) (T, error) {
	var dst T
	err := Unmarshal(entity, &dst)
	return dst, err
}

func AsSlice[T any](entities ...*proto.EntityResponse) ([]T, error) {
	var dst []T
	err := UnmarshalToSlice(&dst, entities...)
	return dst, err
}

func UnmarshalToSlice(dstPtr any, entities ...*proto.EntityResponse) error {
	if len(entities) == 0 {
		return nil
	}
	dstT := reflect.TypeOf(dstPtr)
	if dstPtr == nil || dstT.Kind() != reflect.Pointer || dstT.Elem().Kind() != reflect.Slice {
		return ErrMustPointerSlice
	}

	isStruct := dstT.Elem().Elem().Kind() == reflect.Struct
	isPointer := dstT.Elem().Elem().Kind() == reflect.Pointer
	if isPointer {
		isStruct = dstT.Elem().Elem().Elem().Kind() == reflect.Struct
	}
	if !isStruct {
		return ErrMustPointerSlice
	}

	sort.Sort(proto.EntityResponseIDSort(entities))
	elemType := dstT.Elem().Elem()

	valuePtr := reflect.ValueOf(dstPtr)
	dst := valuePtr.Elem()
	sliceLen := valuePtr.Elem().Len()
	appendLen := len(entities)
	finalSlice := reflect.MakeSlice(reflect.SliceOf(elemType), sliceLen+appendLen, sliceLen+appendLen)
	for x := 0; x < sliceLen; x++ {
		finalSlice.Index(x).Set(dst.Index(x))
	}

	for i, entity := range entities {
		if entity.Entity == nil || entity.Entity.EntityId == "" {
			return ErrEntityIDMissing
		}
		itm, err := forTypeFromResponse(elemType, entity)
		if err != nil {
			return err
		}
		finalSlice.Index(i + sliceLen).Set(itm)
	}

	dst.Set(finalSlice)
	return nil
}

func UnmarshalToMap(dst any, entities ...*proto.EntityResponse) error {
	if len(entities) == 0 {
		return nil
	}
	dstT := reflect.TypeOf(dst)
	if dst == nil || dstT.Kind() != reflect.Map || dstT.Key().Kind() != reflect.String {
		// Basic map check
		return ErrMustMapStringStruct
	} else if !(dstT.Elem().Kind() == reflect.Struct || (dstT.Elem().Kind() == reflect.Ptr && dstT.Elem().Elem().Kind() == reflect.Struct)) {
		// Allow structs & pointer to structs
		return ErrMustMapStringStruct
	}
	inVal := reflect.ValueOf(dst)
	if inVal.IsValid() && inVal.IsZero() {
		//reflect.MakeMapWithSize(dstT, len(entities))
		return ErrNilMapGiven
	}

	elemType := dstT.Elem()
	for _, entity := range entities {
		if entity.Entity == nil || entity.Entity.EntityId == "" {
			return ErrEntityIDMissing
		}
		mapItm, err := forTypeFromResponse(elemType, entity)
		if err != nil {
			return err
		}
		inVal.SetMapIndex(reflect.ValueOf(entity.Entity.EntityId), mapItm)
	}

	return nil
}

func UnmarshalProperties(data map[Property]*proto.Value, v interface{}) error {
	if v == nil || data == nil || len(data) == 0 {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return ErrMustPassPointer
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

		currentProp, _ := ReflectProperty(field, "")
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
		} else if len(subData) > 0 {
			if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				if currentVal.IsZero() {
					currentVal.Set(reflect.New(field.Type.Elem()))
				}
				if err := UnmarshalProperties(subData, currentVal.Interface()); err != nil {
					return err
				}
			} else if field.Type.Kind() == reflect.Struct {
				if err := UnmarshalProperties(subData, currentVal.Addr().Interface()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
