/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/caicloud/helm-registry/pkg/log"
)

// deadline stores the nanoseconds of the deadline.
type deadline int64

// newDeadline creates a deadline.
func newDeadline(limit time.Duration) deadline {
	return deadline(time.Now().UnixNano() + int64(limit))
}

// left returns the remainning time before deadline.
func (dl deadline) left() time.Duration {
	return time.Duration(int64(dl) - time.Now().UnixNano())
}

var counter uint64 = 0

// Locks stores an array of locker. Order by parent then child.
type Locks struct {
	locks []Locker
	name  string
	id    uint64
}

// Name returns name of locks
func (l *Locks) Name() string {
	return fmt.Sprintf("%s(%d)", l.name, l.id)
}

// Lock tries lock for writing. If locked, return true
func (l *Locks) Lock(timeout time.Duration) bool {
	log.Debugf("lock %s", l.Name())
	locks := l.locks
	maxIndex := len(locks) - 1
	if maxIndex < 0 {
		log.Debugf("failed to lock %s due to no underlying lock", l.Name())
		return false
	}
	i := 0
	deadline := newDeadline(timeout)
	for ; i < maxIndex; i++ {
		if !locks[i].RLock(deadline.left()) {
			break
		}
	}
	if i != maxIndex || !locks[maxIndex].Lock(deadline.left()) {
		// rollback
		l.rUnlock(i)
		log.Debugf("failed to lock %s, rollback", l.Name())
		return false
	}
	log.Debugf("lock %s successfully", l.Name())
	return true
}

// rUnlock unlocks locker from length-1 to 0
func (l *Locks) rUnlock(length int) {
	locks := l.locks
	if len(locks) < length || length <= 0 {
		return
	}
	for i := length - 1; i >= 0; i-- {
		locks[i].RUnlock()
	}
}

// Unlock unlock write lock
func (l *Locks) Unlock() {
	log.Debugf("unlock %s", l.Name())
	locks := l.locks
	maxIndex := len(locks) - 1
	if maxIndex < 0 {
		return
	}
	locks[maxIndex].Unlock()
	l.rUnlock(maxIndex)
	log.Debugf("unlock %s successfully", l.Name())
}

// RLock tries lock for reading. If locked, return true
func (l *Locks) RLock(timeout time.Duration) bool {
	log.Debugf("rlock %s", l.Name())
	locks := l.locks
	length := len(locks)
	if length <= 0 {
		log.Debugf("failed to rlock %s due to no underlying lock", l.Name())
		return false
	}
	deadline := newDeadline(timeout)
	for i := 0; i < length; i++ {
		if !locks[i].RLock(deadline.left()) {
			l.rUnlock(i)
			log.Debugf("failed to rlock %s, rollback", l.Name())
			return false
		}
	}
	log.Debugf("rlock %s successfully", l.Name())
	return true
}

// RUnlock unlock read lock
func (l *Locks) RUnlock() {
	log.Debugf("runlock %s", l.Name())
	locks := l.locks
	l.rUnlock(len(locks))
	log.Debugf("runlock %s successfully", l.Name())
}

// HierarchicalLock stores the relationship of lockers
type HierarchicalLock struct {
	Lock     Locker
	Children map[string]*HierarchicalLock
}

// NewHierarchicalLock creates a HierarchicalLock
func NewHierarchicalLock(locker Locker) *HierarchicalLock {
	return &HierarchicalLock{locker, make(map[string]*HierarchicalLock)}
}

// ResourceLock describes a resource locker
// TODO: the resource lock will stores all locks which have been used.
// So it need to release some locks which does not use frequently.
type ResourceLock struct {
	lock       *sync.Mutex
	Locks      map[string]*HierarchicalLock
	CreateLock func() Locker
}

// NewResourceLock creates a ResourceLock
func NewResourceLock(creator func() Locker) *ResourceLock {
	return &ResourceLock{&sync.Mutex{}, make(map[string]*HierarchicalLock), creator}
}

// Get gets a lock for resources. The locker can lock multiple level resources.
func (rl *ResourceLock) Get(res ...string) Locker {
	rl.lock.Lock()
	defer rl.lock.Unlock()
	result := &Locks{
		name:  path.Join(res...),
		locks: make([]Locker, len(res)),
		id:    atomic.AddUint64(&counter, 1),
	}
	children := rl.Locks
	for i, r := range res {
		lock, ok := children[r]
		if !ok {
			lock = NewHierarchicalLock(rl.CreateLock())
			children[r] = lock
			children = lock.Children
		}
		result.locks[i] = lock.Lock
	}
	log.Debugf("get locks %s", result.Name())
	return result
}

// Close closes all existing lockers
func (rl *ResourceLock) Close() {
	rl.Locks = make(map[string]*HierarchicalLock)
}
