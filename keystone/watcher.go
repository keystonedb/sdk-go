package keystone

import (
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Watcher struct {
	knownValues map[Property]*proto.Value
}

// NewDefaultsWatcher creates a new watcher with the default values of the given type
func NewDefaultsWatcher(v interface{}) (*Watcher, error) {
	val := reflector.Deref(reflect.ValueOf(v))
	return NewWatcher(reflect.New(val.Type()).Interface())
}

// NewWatcher creates a new watcher with the given Value
func NewWatcher(v interface{}) (*Watcher, error) {
	w := &Watcher{
		knownValues: make(map[Property]*proto.Value),
	}

	current, err := Marshal(v)
	if err != nil {
		return nil, err
	}

	w.knownValues = current
	return w, nil
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

	if w.knownValues == nil || len(w.knownValues) == 0 {
		if update {
			w.knownValues = latest
		}
		return latest, nil
	}

	changes := make(map[Property]*proto.Value)
	for k, lV := range latest {
		prev, ok := w.knownValues[k]
		if !ok || proto.MatchValue(prev, "_", lV) != nil {
			changes[k] = lV
		}
	}

	if update {
		w.knownValues = latest
	}

	return changes, nil
}

func (w *Watcher) ReplaceKnownValues(vals map[Property]*proto.Value) {
	w.knownValues = vals
}

func (w *Watcher) AppendKnownValues(vals map[Property]*proto.Value) {
	if w.knownValues == nil || len(w.knownValues) == 0 {
		w.knownValues = vals
		return
	}

	for k, v := range vals {
		w.knownValues[k] = v
	}
	return
}
