/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"sync"
	"time"
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

// Locks stores an array of locker. Order by parent then child.
type Locks []Locker

// Lock tries lock for writing. If locked, return true
func (locks Locks) Lock(timeout time.Duration) bool {
	maxIndex := len(locks) - 1
	if maxIndex < 0 {
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
		locks.rUnlock(i)
		return false
	}
	return true
}

// rUnlock unlocks locker from length-1 to 0
func (locks Locks) rUnlock(length int) {
	if len(locks) < length || length <= 0 {
		return
	}
	for i := length - 1; i >= 0; i-- {
		locks[i].RUnlock()
	}
}

// Unlock unlock write lock
func (locks Locks) Unlock() {
	maxIndex := len(locks) - 1
	if maxIndex < 0 {
		return
	}
	locks[maxIndex].Unlock()
	locks.rUnlock(maxIndex)
}

// RLock tries lock for reading. If locked, return true
func (locks Locks) RLock(timeout time.Duration) bool {
	length := len(locks)
	if length <= 0 {
		return false
	}
	deadline := newDeadline(timeout)
	for i := 0; i < length; i++ {
		if !locks[i].RLock(deadline.left()) {
			locks.rUnlock(i)
			return false
		}
	}
	return true
}

// RUnlock unlock read lock
func (locks Locks) RUnlock() {
	locks.rUnlock(len(locks))
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
	result := Locks{}
	children := rl.Locks
	for _, r := range res {
		lock, ok := children[r]
		if !ok {
			lock = NewHierarchicalLock(rl.CreateLock())
			children[r] = lock
			children = lock.Children
		}
		result = append(result, lock.Lock)
	}
	return result
}

// Close closes all existing lockers
func (rl *ResourceLock) Close() {
	rl.Locks = make(map[string]*HierarchicalLock)
}
