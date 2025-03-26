package keystone

import (
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Watcher struct {
	knownValues map[string]*watcherValue
}

type watcherValue struct {
	Property Property
	Value    *proto.Value
}

// NewDefaultsWatcher creates a new watcher with the default values of the given type
func NewDefaultsWatcher(v interface{}) (*Watcher, error) {
	val := reflector.Deref(reflect.ValueOf(v))
	return NewWatcher(reflect.New(val.Type()).Interface())
}

// NewWatcher creates a new watcher with the given Value
func NewWatcher(v interface{}) (*Watcher, error) {
	current, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	return &Watcher{convert(current)}, nil
}

func convert(current map[Property]*proto.Value) map[string]*watcherValue {
	res := make(map[string]*watcherValue)
	for k, currentV := range current {
		res[k.Name()] = &watcherValue{Property: k, Value: currentV}
	}
	return res
}

func Changes(a, b interface{}) (map[Property]*proto.Value, error) {
	w, err := NewWatcher(a)
	if err != nil {
		return nil, err
	}
	return w.Changes(b, false)
}

func ChangesFromDefault(v interface{}) (map[Property]*proto.Value, error) {
	w, err := NewDefaultsWatcher(v)
	if err != nil {
		return nil, err
	}
	return w.Changes(v, false)
}

// Changes returns the changes between the current Value and the previous Value.
// If update is true, the current Value will be stored as the previous Value
func (w *Watcher) Changes(v interface{}, update bool) (map[Property]*proto.Value, error) {
	latest, err := Marshal(v)
	if err != nil {
		return nil, err
	}

	latestV := convert(latest)

	if w.knownValues == nil || len(w.knownValues) == 0 {
		if update {
			w.knownValues = latestV
		}
		return latest, nil
	}

	changes := make(map[Property]*proto.Value)
	for k, lV := range latest {
		updated := false
		prev, ok := w.knownValues[k.Name()]
		if !ok {
			// If we don't have a previous value, consider changed
			updated = true
		} else {
			matchErr := proto.MatchValue(prev.Value, "_", lV)
			// If the values do not match, consider changed
			updated = matchErr != nil
		}
		if updated {
			changes[k] = lV
		}
	}

	if update {
		w.knownValues = latestV
	}

	return changes, nil
}

func (w *Watcher) ReplaceKnownValues(vals map[Property]*proto.Value) {
	w.knownValues = convert(vals)
}

func (w *Watcher) AppendKnownValues(vals map[Property]*proto.Value) {
	if w.knownValues == nil || len(w.knownValues) == 0 {
		w.knownValues = convert(vals)
		return
	}

	for k, v := range vals {
		w.knownValues[k.Name()] = &watcherValue{Property: k, Value: v}
	}
	return
}
