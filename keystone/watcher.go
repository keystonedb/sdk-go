package keystone

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type Watcher struct {
	knownValues map[Property]*proto.Value
}

// NewDefaultsWatcher creates a new watcher with the default values of the given type
func NewDefaultsWatcher(v interface{}) (*Watcher, error) {
	return NewWatcher(reflect.New(reflect.ValueOf(v).Type()).Interface())
}

// NewWatcher creates a new watcher with the given value
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

// Changes returns the changes between the current value and the previous value.
// If update is true, the current value will be stored as the previous value
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