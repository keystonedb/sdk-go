package keystone

import "time"

type Locker interface {
	SetLockResult(*LockInfo)
}

type EmbeddedLock struct {
	lockInfo *LockInfo
}

type LockInfo struct {
	ID           string
	LockedUntil  time.Time
	Message      string
	LockAcquired bool
}

func (e *EmbeddedLock) LockData() *LockInfo              { return e.lockInfo }
func (e *EmbeddedLock) SetLockResult(lockInfo *LockInfo) { e.lockInfo = lockInfo }
func (e *EmbeddedLock) AcquiredLock() bool {
	if e.lockInfo == nil {
		return false
	}
	return e.lockInfo.LockAcquired
}
