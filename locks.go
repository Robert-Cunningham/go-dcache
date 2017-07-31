package dcache2

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	isLocked int32 //unfortunately we don't have CAS operations for bytes or booleans.
}

func (s *SpinLock) Unlock() {
	atomic.StoreInt32(&s.isLocked, 0)
}

func (s *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&s.isLocked, 0, 1) {
		runtime.Gosched()
	}
}

func (s *SpinLock) IsLocked() bool {
	return (atomic.LoadInt32(&s.isLocked) == 1)
}

//Thanks https://github.com/gansidui/go-utils/blob/master/spinlock/spinlock.go for significant amount of the spinlock

//According to https://groups.google.com/forum/#!topic/golang-nuts/7EnEhM3U7B8, the atomic package handles memory barriers and blocking compiler overoptimization for us.

//Thanks http://lxr.free-electrons.com/source/include/linux/seqlock.h for framework of SeqLock

type SeqLock struct {
	count uint32
	writeLock SpinLock
}

func (s *SeqLock) read_seqbegin() uint32 {
	return atomic.LoadUint32(&s.count)
}

func (s *SeqLock) read_seqretry(start uint32) bool { //does the seqlock need to retry the read?
	return start != atomic.LoadUint32(&s.count)
}

func (s *SeqLock) write_seqlock() {
	s.writeLock.Lock()
	atomic.AddUint32(&s.count, 1)
}

func (s *SeqLock) write_sequnlock() {
	atomic.AddUint32(&s.count, 1)
	s.writeLock.Unlock()
}

func (s *SeqLock) raw_write_seqcount_begin(){
	atomic.AddUint32(&s.count, 1)
}

func (s *SeqLock) raw_write_seqcount_end(){
	atomic.AddUint32(&s.count, 1)
}

func (s *SeqLock) write_seqlock_invalidate() {
	atomic.AddUint32(&s.count, 2)
}

type LockRef struct {
	count uint32
	spinLock SpinLock
}

func (l *LockRef) lockref_get(){
	l.spinLock.Lock()
	l.count++
	l.spinLock.Unlock()
}

func (l *LockRef) Unlock() {
	atomic.StoreInt32(&l.spinLock.isLocked, 0)
}

func (l *LockRef) Lock() {
	for !atomic.CompareAndSwapInt32(&l.spinLock.isLocked, 0, 1) {
		runtime.Gosched()
	}
}