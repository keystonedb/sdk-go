package keystone

type WatchedEntity interface {
	HasWatcher() bool
	Watcher() *Watcher
}

type SettableWatchedEntity interface {
	WatchedEntity
	SetWatcher(*Watcher)
}

type EmbeddedWatcher struct {
	watcher *Watcher
}

func (e *EmbeddedWatcher) HasWatcher() bool {
	return e.watcher != nil
}

func (e *EmbeddedWatcher) Watcher() *Watcher {
	return e.watcher
}

func (e *EmbeddedWatcher) SetWatcher(watcher *Watcher) {
	e.watcher = watcher
}
