/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"os"
	"runtime"
	"testing"
	"time"
)

var locker ResourceLocker

func TestMain(m *testing.M) {
	var err error
	locker, err = Create("memory", nil)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func resourceReader(t *testing.T, locker Locker, lock, unlock, finished chan bool) {
	if locker.RLock(time.Millisecond * 200) {
		lock <- true
		locker.RUnlock()
	} else {
		unlock <- true
	}
	finished <- true
}

func resourceWriter(t *testing.T, locker Locker, readers int, lock, unlock, finished chan bool) {
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

func resourceLock(t *testing.T, threads, count int) {
	runtime.GOMAXPROCS(threads)
	lock, unlock, finished := make(chan bool, count), make(chan bool, count), make(chan bool, count+1)
	for i := 0; i < count; i++ {
		go resourceReader(t, locker.Get("a", "b", "c"), lock, unlock, finished)

	}
	go resourceWriter(t, locker.Get("a", "b"), count, lock, unlock, finished)
	for i := 0; i < count+1; i++ {
		<-finished
	}
}

func TestResourceLock(t *testing.T) {
	defer runtime.GOMAXPROCS(-1)
	resourceLock(t, 1, 1000)
	resourceLock(t, 1, 1000)
	resourceLock(t, 4, 1000)
	resourceLock(t, 4, 1000)
	resourceLock(t, 10, 1000)
	resourceLock(t, 10, 1000)
}

func TestLockConflict(t *testing.T) {
	lock := locker.Get("a", "b", "c")
	if lock.Lock(TimeoutImmediate) {
		if lock.Lock(TimeoutImmediate) {
			t.Fatal("locks conflict")
		}
		lock.Unlock()
	} else {
		t.Fatal("can't lock")
	}
	if lock.Lock(TimeoutImmediate) {
		if lock.RLock(TimeoutImmediate) {
			t.Fatal("locks conflict")
		}
		lock.Unlock()
	} else {
		t.Fatal("can't lock")
	}
	if lock.RLock(TimeoutImmediate) {
		if lock.Lock(TimeoutImmediate) {
			t.Fatal("locks conflict")
		}
		lock.RUnlock()
	} else {
		t.Fatal("can't lock")
	}
	if lock.RLock(TimeoutImmediate) {
		if lock.RLock(TimeoutImmediate) {
			lock.RUnlock()
		} else {
			t.Fatal("locks error")
		}
		lock.RUnlock()
	} else {
		t.Fatal("can't lock")
	}
	if lock.Lock(TimeoutImmediate) {
		lock.Unlock()
	} else {
		t.Fatal("lock invalid")
	}
}
