/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func read(t *testing.T, locker Locker, w, r, finished chan bool) {
	if locker.RLock(TimeoutImmediate) {
		w <- true
		<-r
		locker.RUnlock()
	}
	finished <- true
}

func syncRead(t *testing.T, threads, count int) {
	runtime.GOMAXPROCS(threads)
	locker := NewMemoryLock()
	w, r, finished := make(chan bool, count), make(chan bool, count), make(chan bool, count)
	for i := 0; i < count; i++ {
		go read(t, locker, w, r, finished)
	}
	for i := 0; i < count; i++ {
		<-w
	}
	if len(finished) > 0 {
		t.Fatal("some reader can't lock")
	}
	for i := 0; i < count; i++ {
		r <- true
	}
	for i := 0; i < count; i++ {
		<-finished
	}
}

func TestMemoryReader(t *testing.T) {
	defer runtime.GOMAXPROCS(-1)
	syncRead(t, 1, 1000)
	syncRead(t, 1, 1000)
	syncRead(t, 4, 1000)
	syncRead(t, 4, 1000)
	syncRead(t, 10, 1000)
	syncRead(t, 10, 1000)
}

func reader(t *testing.T, locker Locker, lock, unlock, finished chan bool) {
	if locker.RLock(time.Millisecond * 200) {
		lock <- true
		locker.RUnlock()
	} else {
		unlock <- true
	}
	finished <- true
}

func writer(t *testing.T, locker Locker, readers int, lock, unlock, finished chan bool) {
	if locker.Lock(time.Millisecond * 200) {
		// wait for all readers to exit
		// one second is enough to wait for all locks to finish
		time.Sleep(time.Second)
		if readers != len(lock)+len(unlock) {
			t.Fatal("reader count: ", readers, " blocked: ", len(lock), " timeout: ", len(unlock))
		} else {
			t.Log("reader count: ", readers, " blocked: ", len(lock), " timeout: ", len(unlock))
		}
		locker.Unlock()
	} else {
		t.Log("can't lock")
	}
	finished <- true
}

func syncLock(t *testing.T, threads, count int) {
	runtime.GOMAXPROCS(threads)
	locker := NewMemoryLock()
	lock, unlock, finished := make(chan bool, count), make(chan bool, count), make(chan bool, count+1)
	for i := 0; i < count; i++ {
		go reader(t, locker, lock, unlock, finished)
	}
	go writer(t, locker, count, lock, unlock, finished)
	for i := 0; i < count+1; i++ {
		<-finished
	}
	// check locker status
	memoryLock := locker.(*Lock)
	// wait for all locks to release
	// one second is enough to release all locks
	time.Sleep(time.Second)
	if !reflect.DeepEqual(memoryLock.RWMutex, sync.RWMutex{}) {
		t.Fatal("some locks can't release: ", memoryLock)
	}
}

func TestMemoryLock(t *testing.T) {
	defer runtime.GOMAXPROCS(-1)
	syncLock(t, 1, 1000)
	syncLock(t, 1, 1000)
	syncLock(t, 4, 1000)
	syncLock(t, 4, 1000)
	syncLock(t, 10, 1000)
	syncLock(t, 10, 1000)
}
